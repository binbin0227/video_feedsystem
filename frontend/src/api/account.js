import http from './http'

export async function registerAccount(username, password) {
  const response = await http.post('/account/register', {
    username,
    password,
  })

  return response.data
}

export async function loginAccount(username, password) {
  const response = await http.post('/account/login', {
    username,
    password,
  })

  return response.data
}
