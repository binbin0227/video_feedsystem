import http from "k6/http";
import { check } from "k6";
import { Counter } from "k6/metrics";

const status500 = new Counter("status_500");
const statusOther = new Counter("status_other");

const vus = Number(__ENV.VUS || 10);
const duration = __ENV.DURATION || "60s";

export const options = {
    discardResponseBodies: true,

    scenarios: {
        detail_test: {
            executor: "constant-vus",
            vus: vus,
            duration: duration,
        },
    },
};

export default function () {
    const videoID = "629324078238408507";
    const url = `http://127.0.0.1:8080/video/detail?video_id=${videoID}`;

    const response = http.get(url);

    if (response.status === 500) {
        status500.add(1);
    } else if (response.status !== 200) {
        statusOther.add(1);
    }

    check(response, {
        "status is 200": (res) => res.status === 200,
    });
}
