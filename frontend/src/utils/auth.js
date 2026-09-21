const ACCESS_TOKEN_KEY = 'token'
const REFRESH_TOKEN_KEY = 'refresh_token'
const AUTH_CHANGED_EVENT = 'auth-changed'

function notifyAuthChanged() {
  window.dispatchEvent(new Event(AUTH_CHANGED_EVENT))
}

export function saveAuthTokens(accessToken, refreshToken = '') {
  localStorage.setItem(ACCESS_TOKEN_KEY, accessToken)

  if (refreshToken) {
    localStorage.setItem(REFRESH_TOKEN_KEY, refreshToken)
  }

  notifyAuthChanged()
}

export function getAccessToken() {
  return localStorage.getItem(ACCESS_TOKEN_KEY)
}

export function getRefreshToken() {
  return localStorage.getItem(REFRESH_TOKEN_KEY)
}

export function clearAuthTokens() {
  localStorage.removeItem(ACCESS_TOKEN_KEY)
  localStorage.removeItem(REFRESH_TOKEN_KEY)
  notifyAuthChanged()
}

export function isLoggedIn() {
  return Boolean(getAccessToken())
}

// JWT 中的 account_id 是 int64。直接 JSON.parse 会变成 JavaScript Number，
// 对雪花 ID 可能造成精度丢失，因此从原始载荷中提取并始终以字符串返回。
export function getAccountId() {
  const accessToken = getAccessToken()
  const payload = accessToken?.split('.')[1]

  if (!payload) {
    return ''
  }

  try {
    const normalized = payload.replace(/-/g, '+').replace(/_/g, '/')
    const padded = normalized.padEnd(Math.ceil(normalized.length / 4) * 4, '=')
    const decodedPayload = atob(padded)
    const match = decodedPayload.match(/"account_id"\s*:\s*(?:"([0-9]+)"|([0-9]+))/)

    return match?.[1] || match?.[2] || ''
  } catch {
    return ''
  }
}
