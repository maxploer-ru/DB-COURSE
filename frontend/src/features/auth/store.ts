import { create } from 'zustand'
import type { AuthTokens, User } from '../../shared/api/types'

type AuthState = {
  user: User | null
  accessToken: string | null
  isAuthenticated: boolean
  setSession: (payload: { user: User; tokens: AuthTokens }) => void
  logout: () => void
}

export const TOKEN_KEY = 'zvideo_tokens'
export const USER_KEY = 'zvideo_user'

function loadAuthFromStorage(): Pick<AuthState, 'user' | 'accessToken' | 'isAuthenticated'> {
  const userRaw = localStorage.getItem(USER_KEY)
  const tokensRaw = localStorage.getItem(TOKEN_KEY)

  const parsedUser = userRaw ? (JSON.parse(userRaw) as Partial<User>) : null
  const user = parsedUser
    ? ({
        id: parsedUser.id ?? 0,
        username: parsedUser.username ?? '',
        email: parsedUser.email ?? '',
        role: parsedUser.role ?? 'user',
        notificationsEnabled: parsedUser.notificationsEnabled ?? true,
      } as User)
    : null
  const tokens = tokensRaw ? (JSON.parse(tokensRaw) as AuthTokens) : null

  return {
    user,
    accessToken: tokens?.accessToken ?? null,
    isAuthenticated: Boolean(tokens?.accessToken),
  }
}

export const useAuthStore = create<AuthState>((set) => ({
  ...loadAuthFromStorage(),
  setSession: ({ user, tokens }) => {
    localStorage.setItem(USER_KEY, JSON.stringify(user))
    localStorage.setItem(TOKEN_KEY, JSON.stringify(tokens))

    set({
      user,
      accessToken: tokens.accessToken,
      isAuthenticated: true,
    })
  },
  logout: () => {
    localStorage.removeItem(USER_KEY)
    localStorage.removeItem(TOKEN_KEY)

    set({
      user: null,
      accessToken: null,
      isAuthenticated: false,
    })
  },
}))


