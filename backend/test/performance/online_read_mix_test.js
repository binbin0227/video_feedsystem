import http from 'k6/http'
import { check } from 'k6'
import { Counter, Rate } from 'k6/metrics'

const baseURL = (__ENV.BASE_URL || 'http://47.80.23.81').replace(/\/$/, '')
const targetRPS = Number(__ENV.TARGET_RPS || 5)
const duration = __ENV.DURATION || '60s'
const detailWeight = Number(__ENV.DETAIL_WEIGHT || 0.75)
const videoIDs = (__ENV.VIDEO_IDS || '638423124211138565,638422797122535429,637853189483266053')
  .split(',')
  .map((item) => item.trim())
  .filter(Boolean)

if (!Number.isFinite(targetRPS) || targetRPS <= 0) {
  throw new Error('TARGET_RPS 必须是大于 0 的数字')
}
if (!Number.isFinite(detailWeight) || detailWeight < 0 || detailWeight > 1) {
  throw new Error('DETAIL_WEIGHT 必须是 0 到 1 之间的数字')
}
if (videoIDs.length === 0) {
  throw new Error('VIDEO_IDS 至少需要包含一个视频 ID')
}

const businessSuccess = new Rate('business_success')
const detailRequests = new Counter('detail_requests')
const hotFeedRequests = new Counter('hot_feed_requests')
const clientErrors = new Counter('response_4xx')
const serverErrors = new Counter('response_5xx')

export const options = {
  scenarios: {
    online_read_mix: {
      executor: 'constant-arrival-rate',
      rate: targetRPS,
      timeUnit: '1s',
      duration,
      preAllocatedVUs: Math.max(10, targetRPS * 2),
      maxVUs: Math.max(50, targetRPS * 4),
    },
  },
  thresholds: {
    checks: ['rate>0.99'],
    http_req_failed: ['rate<0.01'],
    http_req_duration: ['p(95)<1500', 'p(99)<3000'],
    business_success: ['rate>0.99'],
    dropped_iterations: ['count==0'],
  },
}

function recordStatus(response) {
  if (response.status >= 400 && response.status < 500) {
    clientErrors.add(1)
  } else if (response.status >= 500) {
    serverErrors.add(1)
  }
}

function requestVideoDetail() {
  detailRequests.add(1)

  const videoID = videoIDs[Math.floor(Math.random() * videoIDs.length)]
  const response = http.get(`${baseURL}/api/video/detail?video_id=${encodeURIComponent(videoID)}`, {
    headers: { Accept: 'application/json' },
    tags: { name: 'GET /api/video/detail' },
    timeout: '10s',
  })

  recordStatus(response)

  let returnedVideoID = ''
  try {
    returnedVideoID = String(response.json('video.id') || '')
  } catch (_) {
    // 非 JSON 响应会在业务校验中计为失败。
  }

  const success = response.status === 200 && returnedVideoID === videoID
  businessSuccess.add(success)
  check(response, {
    'video detail status is 200': (res) => res.status === 200,
    'video detail returns requested video': () => returnedVideoID === videoID,
  })
}

function requestHotFeed() {
  hotFeedRequests.add(1)

  const response = http.get(`${baseURL}/api/feed/hot`, {
    headers: { Accept: 'application/json' },
    tags: { name: 'GET /api/feed/hot' },
    timeout: '10s',
  })

  recordStatus(response)

  let videos
  try {
    videos = response.json('videos')
  } catch (_) {
    // 非 JSON 响应会在业务校验中计为失败。
  }

  const success = response.status === 200 && Array.isArray(videos)
  businessSuccess.add(success)
  check(response, {
    'hot feed status is 200': (res) => res.status === 200,
    'hot feed returns videos array': () => Array.isArray(videos),
  })
}

export default function () {
  if (Math.random() < detailWeight) {
    requestVideoDetail()
    return
  }

  requestHotFeed()
}
