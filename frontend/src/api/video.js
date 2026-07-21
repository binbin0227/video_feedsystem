import http from './http'

export async function uploadVideoFile(file, onUploadProgress) {
  const formData = new FormData()
  formData.append('file', file)

  const response = await http.post('/video/upload-video', formData, {
    timeout: 0,
    onUploadProgress,
  })

  return response.data
}

export async function uploadCoverFile(file, onUploadProgress) {
  const formData = new FormData()
  formData.append('file', file)

  const response = await http.post('/video/upload-cover', formData, {
    timeout: 0,
    onUploadProgress,
  })

  return response.data
}

export async function publishVideo({ title, description, playUrl, coverUrl }) {
  const response = await http.post('/video/publish', {
    title,
    description,
    play_url: playUrl,
    cover_url: coverUrl,
  })

  return response.data
}

export async function getVideosByAuthor(accountId, { cursor = '', limit = 6 } = {}) {
  const params = {
    author_id: String(accountId),
    limit,
  }

  if (cursor) {
    params.cursor = cursor
  }

  const response = await http.get('/video/list-by-author-id', { params })
  return response.data
}

export async function getLikedVideos({ cursor = '', limit = 6 } = {}) {
  const params = { limit }

  if (cursor) {
    params.cursor = cursor
  }

  const response = await http.get('/video/liked', { params })
  return response.data
}

export async function getVideoDetail(videoId) {
  const response = await http.get('/video/detail', {
    params: { video_id: String(videoId) },
  })

  return response.data
}

export async function getVideoLikeStatus(videoId) {
  const response = await http.get('/video/like-status', {
    params: { video_id: String(videoId) },
  })

  return response.data
}

export async function likeVideo(videoId) {
  const response = await http.post('/video/like', {
    video_id: String(videoId),
  })

  return response.data
}

export async function unlikeVideo(videoId) {
  const response = await http.post('/video/unlike', {
    video_id: String(videoId),
  })

  return response.data
}
