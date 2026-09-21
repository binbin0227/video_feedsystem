import http from './http'

export async function registerAccount(username, password) {
  const response = await http.post('/account/register', {
    username,
    password,
  }, {
    skipAuthRefresh: true,
  })

  return response.data
}

export async function loginAccount(username, password) {
  const response = await http.post('/account/login', {
    username,
    password,
  }, {
    skipAuthRefresh: true,
  })

  return response.data
}

export async function logoutAccount(refreshToken) {
  const response = await http.post('/account/logout', {
    refresh_token: refreshToken,
  })

  return response.data
}

export async function getAccountProfile(accountId) {
  const response = await http.get('/account/profile', {
    params: { account_id: String(accountId) },
  })

  return response.data
}

export async function searchAccounts(keyword) {
  const response = await http.get('/account/search', {
    params: { keyword },
  })

  return response.data
}
