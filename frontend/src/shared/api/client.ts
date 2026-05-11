import axios from 'axios'
import { TOKEN_KEY, USER_KEY, useAuthStore } from '../../features/auth/store'
import type { User } from './types'

type TokensStorage = {
  accessToken?: string
}

type RetriableRequestConfig = {
  _retry?: boolean
}

export const apiClient = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL ?? '/api',
  timeout: 10_000,
  withCredentials: true,
})

let refreshPromise: Promise<string> | null = null

apiClient.interceptors.request.use((config) => {
  const tokensRaw = localStorage.getItem(TOKEN_KEY)
  if (!tokensRaw) {
    return config
  }

  const tokens = JSON.parse(tokensRaw) as TokensStorage
  if (!tokens.accessToken) {
    return config
  }

  config.headers.Authorization = `Bearer ${tokens.accessToken}`
  return config
})

apiClient.interceptors.response.use(
  (response) => response,
  async (error) => {
    const originalRequest = error.config as RetriableRequestConfig & typeof error.config
    if (!error.response || error.response.status !== 401 || originalRequest?._retry) {
      return Promise.reject(error)
    }

    try {
      originalRequest._retry = true
      if (!refreshPromise) {
        refreshPromise = axios
          .post(
            `${import.meta.env.VITE_API_BASE_URL ?? '/api'}/refresh`,
            {},
            {
              withCredentials: true,
            },
          )
          .then((refreshResponse) => {
            const payload = refreshResponse.data as {
              user: {
                id: number
                username: string
                email: string
                role: string
                notifications_enabled: boolean
              }
              access_token: string
            }

            const user: User = {
              id: payload.user.id,
              username: payload.user.username,
              email: payload.user.email,
              role: payload.user.role,
              notificationsEnabled: payload.user.notifications_enabled,
            }

            useAuthStore.getState().setSession({
              user,
              tokens: {
                accessToken: payload.access_token,
              },
            })

            return payload.access_token
          })
          .finally(() => {
            refreshPromise = null
          })
      }

      const newAccessToken = await refreshPromise
      originalRequest.headers.Authorization = `Bearer ${newAccessToken}`
      return apiClient.request(originalRequest)
    } catch (refreshError) {
      localStorage.removeItem(TOKEN_KEY)
      localStorage.removeItem(USER_KEY)
      useAuthStore.getState().logout()
      return Promise.reject(refreshError)
    }
  },
)




