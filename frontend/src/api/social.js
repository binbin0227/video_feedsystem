import http from './http'

export async function getFollowStatus(accountId) {
  const response = await http.get('/social/status', {
    params: { vlogger_id: String(accountId) },
  })

  return response.data
}

export async function followAccount(accountId) {
  const response = await http.post('/social/follow', {
    vlogger_id: String(accountId),
  })

  return response.data
}

export async function unfollowAccount(accountId) {
  const response = await http.post('/social/unfollow', {
    vlogger_id: String(accountId),
  })

  return response.data
}

export async function getFollowingAccounts({ cursor = '', limit = 12 } = {}) {
  const params = { limit }

  if (cursor) {
    params.cursor = cursor
  }

  const response = await http.get('/social/following', { params })
  return response.data
}

export async function getFollowerAccounts({ cursor = '', limit = 12 } = {}) {
  const params = { limit }

  if (cursor) {
    params.cursor = cursor
  }

  const response = await http.get('/social/followers', { params })
  return response.data
}
