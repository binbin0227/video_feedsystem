const TOKEN_KEY = 'token'
const AUTH_CHANGED_EVENT = 'auth-changed'

function notifyAuthChanged() {
  window.dispatchEvent(new Event(AUTH_CHANGED_EVENT))
}

export function saveToken(token) {
  localStorage.setItem(TOKEN_KEY, token)
  notifyAuthChanged()
}

export function getToken() {
  return localStorage.getItem(TOKEN_KEY)
}

export function removeToken() {
  localStorage.removeItem(TOKEN_KEY)
  notifyAuthChanged()
}

export function isLoggedIn() {
  return Boolean(getToken())
}

// JWT 中的 account_id 是 int64。直接 JSON.parse 会变成 JavaScript Number，
// 对雪花 ID 可能造成精度丢失，因此从原始载荷中提取并始终以字符串返回。
export function getAccountId() {
  const token = getToken()
  const payload = token?.split('.')[1]

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
