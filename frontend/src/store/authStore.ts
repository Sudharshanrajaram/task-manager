import { create } from 'zustand'

export interface User {
  id: string
  name: string
  email: string
  created_at: string
}

interface AuthState {
  user: User | null
  accessToken: string | null
  refreshToken: string | null
  isAuthenticated: boolean
  setAuth: (user: User, accessToken: string, refreshToken: string) => void
  setTokens: (accessToken: string, refreshToken: string) => void
  setUser: (user: User) => void
  logout: () => void
}

const STORAGE_KEY_USER = 'taskflow_user'
const STORAGE_KEY_ACCESS = 'taskflow_access_token'
const STORAGE_KEY_REFRESH = 'taskflow_refresh_token'

function getInitialState() {
  try {
    const userStr = localStorage.getItem(STORAGE_KEY_USER)
    const access = localStorage.getItem(STORAGE_KEY_ACCESS)
    const refresh = localStorage.getItem(STORAGE_KEY_REFRESH)
    const user = userStr ? JSON.parse(userStr) : null

    return {
      user,
      accessToken: access,
      refreshToken: refresh,
      isAuthenticated: Boolean(access && user),
    }
  } catch {
    return {
      user: null,
      accessToken: null,
      refreshToken: null,
      isAuthenticated: false,
    }
  }
}

export const useAuthStore = create<AuthState>((set) => ({
  ...getInitialState(),

  setAuth: (user, accessToken, refreshToken) => {
    localStorage.setItem(STORAGE_KEY_USER, JSON.stringify(user))
    localStorage.setItem(STORAGE_KEY_ACCESS, accessToken)
    localStorage.setItem(STORAGE_KEY_REFRESH, refreshToken)
    set({ user, accessToken, refreshToken, isAuthenticated: true })
  },

  setTokens: (accessToken, refreshToken) => {
    localStorage.setItem(STORAGE_KEY_ACCESS, accessToken)
    localStorage.setItem(STORAGE_KEY_REFRESH, refreshToken)
    set({ accessToken, refreshToken, isAuthenticated: true })
  },

  setUser: (user) => {
    localStorage.setItem(STORAGE_KEY_USER, JSON.stringify(user))
    set({ user })
  },

  logout: () => {
    localStorage.removeItem(STORAGE_KEY_USER)
    localStorage.removeItem(STORAGE_KEY_ACCESS)
    localStorage.removeItem(STORAGE_KEY_REFRESH)
    set({ user: null, accessToken: null, refreshToken: null, isAuthenticated: false })
  },
}))

