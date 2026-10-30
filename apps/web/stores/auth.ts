// apps/web/stores/auth.ts
import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export interface UserProfile {
  id: string
  username: string
  role: string
  edit_count: number
  profile_bio?: string
  avatar_url?: string
}

export const useAuthStore = defineStore('auth', () => {
  const user = ref<UserProfile | null>(null)
  const token = ref<string | null>(null)
  const isLoaded = ref(false)

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

  return {
    user,
    token,
    isLoaded,
    isAuthenticated,
    userRole,
    setAuth,
    clearAuth,
    initAuth,
  }
})

