import { apiClient } from './client'
import type { User } from '../store/authStore'

export interface AuthResponse {
  user: User
  access_token: string
  refresh_token: string
}

export const authApi = {
  register: (name: string, email: string, password: string) =>
    apiClient.post<AuthResponse>('/auth/register', { name, email, password }).then((r) => r.data),

  login: (email: string, password: string) =>
    apiClient.post<AuthResponse>('/auth/login', { email, password }).then((r) => r.data),

  me: () => apiClient.get<{ user: User }>('/auth/me').then((r) => r.data.user),
}

