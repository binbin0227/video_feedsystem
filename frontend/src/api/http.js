import axios from 'axios'
import { API_BASE_URL } from './config'
import { getToken, removeToken } from '../utils/auth'

const http = axios.create({
  baseURL: API_BASE_URL,
  timeout: 15000,
})

http.interceptors.request.use((config) => {
  const token = getToken()

  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }

  return config
})

http.interceptors.response.use(
  (response) => response,
  (error) => {
    const status = error.response?.status
    const backendMessage = error.response?.data?.message
    const networkMessage = !error.response
      ? `无法连接后端服务（${API_BASE_URL}），请确认后端已启动且允许当前前端地址跨域访问`
      : ''

    if (status === 401) {
      removeToken()
    }

    error.message = backendMessage || networkMessage || error.message || '请求失败，请稍后重试'
    return Promise.reject(error)
  },
)

export default http
