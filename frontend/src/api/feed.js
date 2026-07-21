import http from './http'

export async function getPublicFeed({ cursor = '', limit = 6 } = {}) {
  const params = { limit }

  if (cursor) {
    params.cursor = cursor
  }

  const response = await http.get('/feed/list', { params })
  return response.data
}
