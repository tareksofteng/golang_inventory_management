import axios from 'axios'

// One axios instance for the whole app. Components never call axios directly.
const api = axios.create({
  baseURL: import.meta.env.VITE_API_URL || 'http://localhost:9000/api/v1',
  headers: { 'Content-Type': 'application/json' },
})

// --- Request interceptor: attach the access token to every request ---------
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('access_token')
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

// --- Response interceptor: on 401, try ONE silent refresh then retry -------
let isRefreshing = false
let queue = []

function flushQueue(error, token = null) {
  queue.forEach((p) => (error ? p.reject(error) : p.resolve(token)))
  queue = []
}

api.interceptors.response.use(
  (res) => res,
  async (error) => {
    const original = error.config
    const refreshToken = localStorage.getItem('refresh_token')

    // Only attempt refresh once per request, and never for the auth endpoints.
    if (
      error.response?.status === 401 &&
      !original._retry &&
      refreshToken &&
      !original.url.includes('/auth/')
    ) {
      if (isRefreshing) {
        // Another request is already refreshing — wait for it.
        return new Promise((resolve, reject) => {
          queue.push({ resolve, reject })
        }).then((token) => {
          original.headers.Authorization = `Bearer ${token}`
          return api(original)
        })
      }

      original._retry = true
      isRefreshing = true
      try {
        const { data } = await axios.post(
          `${api.defaults.baseURL}/auth/refresh`,
          { refresh_token: refreshToken },
        )
        const newAccess = data.data.access_token
        localStorage.setItem('access_token', newAccess)
        localStorage.setItem('refresh_token', data.data.refresh_token)
        flushQueue(null, newAccess)
        original.headers.Authorization = `Bearer ${newAccess}`
        return api(original)
      } catch (e) {
        flushQueue(e, null)
        localStorage.clear()
        window.location.href = '/login'
        return Promise.reject(e)
      } finally {
        isRefreshing = false
      }
    }

    return Promise.reject(error)
  },
)

export default api
