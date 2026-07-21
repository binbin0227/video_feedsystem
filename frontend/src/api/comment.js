import http from './http'

export async function getCommentList(videoId, { cursor = '', limit = 10 } = {}) {
  const params = {
    video_id: String(videoId),
    limit,
  }

  if (cursor) {
    params.cursor = cursor
  }

  const response = await http.get('/comment/list', { params })
  return response.data
}

export async function publishComment(videoId, content) {
  const response = await http.post('/comment/publish', {
    video_id: String(videoId),
    content,
  })

  return response.data
}

export async function deleteComment(commentId) {
  const response = await http.delete('/comment/delete', {
    params: { comment_id: String(commentId) },
  })

  return response.data
}
