import axios from 'axios'
import { API_BASE_URL } from './config'
import { clearAuthTokens, getAccessToken, getRefreshToken, saveAuthTokens } from '../utils/auth'

const http = axios.create({
  baseURL: API_BASE_URL,
  timeout: 15000,
})

const refreshHttp = axios.create({
  baseURL: API_BASE_URL,
  timeout: 15000,
})

let refreshPromise = null
const AUTH_REFRESH_LOCK_NAME = 'video-feedsystem-auth-refresh'

function normalizeError(error) {
  const backendMessage = error.response?.data?.message
  const networkMessage = !error.response
    ? `无法连接后端服务（${API_BASE_URL}），请确认后端已启动且允许当前前端地址跨域访问`
    : ''

  error.message = backendMessage || networkMessage || error.message || '请求失败，请稍后重试'
  return error
}

async function performAccessTokenRefresh(failedAccessToken) {
  const currentAccessToken = getAccessToken()
  if (failedAccessToken && currentAccessToken && currentAccessToken !== failedAccessToken) {
    return currentAccessToken
  }

  const refreshToken = getRefreshToken()

  if (!refreshToken) {
    throw new Error('登录状态已失效，请重新登录')
  }

  const response = await refreshHttp.post('/account/refresh', {
    refresh_token: refreshToken,
  })
  const data = response.data
  const accessToken = data?.access_token
  const newRefreshToken = data?.refresh_token

  if (!accessToken || !newRefreshToken) {
    throw new Error('刷新登录状态的响应不完整')
  }

  saveAuthTokens(accessToken, newRefreshToken)
  return accessToken
}

function refreshAccessToken(failedAccessToken) {
  if (!refreshPromise) {
    const refresh = () => performAccessTokenRefresh(failedAccessToken)
    const locks = globalThis.navigator?.locks

    if (locks) {
      refreshPromise = locks.request(AUTH_REFRESH_LOCK_NAME, refresh)
    } else {
      refreshPromise = refresh().catch((error) => {
        const replacementAccessToken = getAccessToken()
        if (failedAccessToken && replacementAccessToken && replacementAccessToken !== failedAccessToken) {
          return replacementAccessToken
        }
        throw error
      })
    }

    refreshPromise = refreshPromise.finally(() => {
      refreshPromise = null
    })
  }

  return refreshPromise
}

function getRequestAccessToken(config) {
  const headers = config?.headers
  const authorization = typeof headers?.get === 'function'
    ? headers.get('Authorization')
    : headers?.Authorization || headers?.authorization

  if (typeof authorization !== 'string') {
    return ''
  }

  const match = authorization.match(/^Bearer\s+(.+)$/i)
  return match?.[1] || ''
}

function updateLogoutRefreshToken(config) {
  if (config.url !== '/account/logout') {
    return
  }

  const refreshToken = getRefreshToken()
  if (!refreshToken) {
    return
  }

  let data = config.data
  if (typeof data === 'string') {
    try {
      data = JSON.parse(data)
    } catch {
      data = {}
    }
  }

  config.data = JSON.stringify({
    ...data,
    refresh_token: refreshToken,
  })
}

http.interceptors.request.use((config) => {
  const accessToken = getAccessToken()

  if (accessToken) {
    config.headers.Authorization = `Bearer ${accessToken}`
  }

  return config
})

http.interceptors.response.use(
  (response) => response,
  async (error) => {
    const status = error.response?.status
    const originalRequest = error.config
    const canRefresh = status === 401
      && originalRequest
      && !originalRequest._retry
      && !originalRequest.skipAuthRefresh
      && Boolean(getRefreshToken())

    if (canRefresh) {
      originalRequest._retry = true

      try {
        const failedAccessToken = getRequestAccessToken(originalRequest)
        const accessToken = await refreshAccessToken(failedAccessToken)
        originalRequest.headers = originalRequest.headers || {}
        originalRequest.headers.Authorization = `Bearer ${accessToken}`
        updateLogoutRefreshToken(originalRequest)
        return http(originalRequest)
      } catch (refreshError) {
        clearAuthTokens()
        return Promise.reject(normalizeError(refreshError))
      }
    }

    if (status === 401 && !originalRequest?.skipAuthRefresh) {
      clearAuthTokens()
    }

    return Promise.reject(normalizeError(error))
  },
)

export default http
