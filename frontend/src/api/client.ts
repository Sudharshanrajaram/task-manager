import axios from 'axios'
import { useAuthStore } from '../store/authStore'

const BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'

export const apiClient = axios.create({
  baseURL: `${BASE_URL}/api`,
  headers: { 'Content-Type': 'application/json' },
  timeout: 15000,
})

// Attach JWT access token to outbound requests
apiClient.interceptors.request.use((config) => {
  const token = localStorage.getItem('taskflow_access_token')
  if (token && config.headers) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// Intercept 401 and attempt token refresh
let isRefreshing = false
let failedQueue: Array<{
  resolve: (value?: unknown) => void
  reject: (reason?: unknown) => void
}> = []

const processQueue = (error: unknown, token: string | null = null) => {
  failedQueue.forEach((prom) => {
    if (error) {
      prom.reject(error)
    } else {
      prom.resolve(token)
    }
  })
  failedQueue = []
}

apiClient.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config

    if (error.response?.status === 401 && !originalRequest._retry && !originalRequest.url?.includes('/auth/')) {
      if (isRefreshing) {
        return new Promise((resolve, reject) => {
          failedQueue.push({ resolve, reject })
        })
          .then((token) => {
            originalRequest.headers.Authorization = `Bearer ${token}`
            return apiClient(originalRequest)
          })
          .catch((err) => Promise.reject(err))
      }

      originalRequest._retry = true
      isRefreshing = true

      const refreshToken = localStorage.getItem('taskflow_refresh_token')
      if (!refreshToken) {
        useAuthStore.getState().logout()
        return Promise.reject(new Error('Session expired. Please log in again.'))
      }

      try {
        const { data } = await axios.post(`${BASE_URL}/api/auth/refresh`, {
          refresh_token: refreshToken,
        })

        const newAccessToken = data.access_token
        const newRefreshToken = data.refresh_token
        useAuthStore.getState().setTokens(newAccessToken, newRefreshToken)

        processQueue(null, newAccessToken)
        originalRequest.headers.Authorization = `Bearer ${newAccessToken}`
        return apiClient(originalRequest)
      } catch (refreshErr) {
        processQueue(refreshErr, null)
        useAuthStore.getState().logout()
        return Promise.reject(new Error('Session expired. Please log in again.'))
      } finally {
        isRefreshing = false
      }
    }

    const message =
      error.response?.data?.message ||
      error.message ||
      'An unexpected error occurred'
    return Promise.reject(new Error(message))
  }
)
