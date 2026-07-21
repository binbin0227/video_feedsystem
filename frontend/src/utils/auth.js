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
