import http from "k6/http";
import { check, sleep } from "k6";
import { Counter, Rate } from "k6/metrics";

const baseURL = (__ENV.BASE_URL || "http://47.80.23.81").replace(/\/$/, "");
const videoID = __ENV.VIDEO_ID || "637853189483266053";
const profile = (__ENV.PROFILE || "smoke").toLowerCase();
const requestInterval = Number(__ENV.REQUEST_INTERVAL || (profile === "rate" ? 0 : 0.1));

const clientErrors = new Counter("video_detail_4xx");
const serverErrors = new Counter("video_detail_5xx");
const businessSuccess = new Rate("video_detail_business_success");

// 默认只执行轻量冒烟测试；需要压测时显式指定 PROFILE=load 或 PROFILE=stress。
const profiles = {
    smoke: [
        { duration: "5s", target: 1 },
        { duration: "10s", target: 1 },
        { duration: "5s", target: 0 },
    ],
    load: [
        { duration: "15s", target: 5 },
        { duration: "30s", target: 10 },
        { duration: "60s", target: 10 },
        { duration: "15s", target: 0 },
    ],
    stress: [
        { duration: "20s", target: 10 },
        { duration: "30s", target: 25 },
        { duration: "30s", target: 50 },
        { duration: "60s", target: 50 },
        { duration: "20s", target: 0 },
    ],
};

if (!profiles[profile] && profile !== "rate") {
    throw new Error(`未知 PROFILE=${profile}，可选值为 smoke、load、stress、rate`);
}

const targetRPS = Number(__ENV.TARGET_RPS || 50);
const rateDuration = __ENV.DURATION || "60s";
const preAllocatedVUs = Number(__ENV.PRE_ALLOCATED_VUS || Math.max(20, Math.ceil(targetRPS / 2)));
const maxVUs = Number(__ENV.MAX_VUS || Math.max(100, targetRPS * 2));

if (profile === "rate" && (!Number.isFinite(targetRPS) || targetRPS <= 0)) {
    throw new Error("TARGET_RPS 必须是大于 0 的数字");
}

const scenario = profile === "rate"
    ? {
        executor: "constant-arrival-rate",
        rate: targetRPS,
        timeUnit: "1s",
        duration: rateDuration,
        preAllocatedVUs,
        maxVUs,
      }
    : {
        executor: "ramping-vus",
        startVUs: 0,
        stages: profiles[profile],
        gracefulRampDown: "10s",
      };

export const options = {
    scenarios: {
        video_detail: scenario,
    },
    thresholds: {
        checks: ["rate>0.99"],
        http_req_failed: ["rate<0.01"],
        http_req_duration: ["p(95)<500", "p(99)<1000"],
        video_detail_business_success: ["rate>0.99"],
    },
};

export default function () {
    const url = `${baseURL}/api/video/detail?video_id=${encodeURIComponent(videoID)}`;
    const response = http.get(url, {
        headers: { Accept: "application/json" },
        tags: { name: "GET /api/video/detail" },
        timeout: "10s",
    });

    if (response.status >= 400 && response.status < 500) {
        clientErrors.add(1);
    } else if (response.status >= 500) {
        serverErrors.add(1);
    }

    let returnedVideoID = "";
    try {
        returnedVideoID = String(response.json("video.id") || "");
    } catch (_) {
        // 非 JSON 响应会在下面的业务校验中计为失败。
    }

    const success = response.status === 200 && returnedVideoID === videoID;
    businessSuccess.add(success);

    check(response, {
        "status is 200": (res) => res.status === 200,
        "content type is json": (res) =>
            String(res.headers["Content-Type"] || "").includes("application/json"),
        "response contains requested video": () => returnedVideoID === videoID,
    });

    if (requestInterval > 0) {
        sleep(requestInterval);
    }
}
