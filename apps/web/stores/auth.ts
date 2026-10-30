// apps/web/stores/auth.ts
import { computed, reactive } from 'vue'

export interface UserProfile {
  id: string
  username: string
  role: string
  edit_count: number
  profile_bio?: string
  avatar_url?: string
}

export const useAuthStore = () => {
  const user = useState<UserProfile | null>('auth_user', () => null)
  const token = useState<string | null>('auth_token', () => null)
  const isLoaded = useState<boolean>('auth_is_loaded', () => false)

  const isAuthenticated = computed(() => !!user.value)
  const userRole = computed(() => user.value?.role || 'ANONYMOUS')

  const setAuth = (userData: UserProfile, tokenStr: string) => {
    user.value = userData
    token.value = tokenStr
    if (import.meta.client) {
      localStorage.setItem('treewiki_token', tokenStr)
      localStorage.setItem('treewiki_user', JSON.stringify(userData))
    }
  }

  const clearAuth = () => {
    user.value = null
    token.value = null
    if (import.meta.client) {
      localStorage.removeItem('treewiki_token')
      localStorage.removeItem('treewiki_user')
    }
  }

  const initAuth = () => {
    if (import.meta.client) {
      const storedToken = localStorage.getItem('treewiki_token')
      const storedUser = localStorage.getItem('treewiki_user')
      if (storedToken && storedUser) {
        try {
          token.value = storedToken
          user.value = JSON.parse(storedUser)
        } catch {
          clearAuth()
        }
      }
      isLoaded.value = true
    }
  }

  return reactive({
    user,
    token,
    isLoaded,
    isAuthenticated,
    userRole,
    setAuth,
    clearAuth,
    initAuth,
  })
}
