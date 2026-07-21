import { API_BASE_URL } from '../api/config'

export function getMediaUrl(path) {
  if (!path) {
    return ''
  }

  if (/^https?:\/\//i.test(path)) {
    return path
  }

  if (path.startsWith('/uploads/')) {
    return `${API_BASE_URL}${path}`
  }

  if (path.startsWith('uploads/')) {
    return `${API_BASE_URL}/${path}`
  }

  return path
}
