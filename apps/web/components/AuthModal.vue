<!-- apps/web/components/AuthModal.vue -->
<template>
  <div v-if="isOpen" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/70 backdrop-blur-xs">
    <div role="dialog" aria-modal="true" aria-labelledby="auth-dialog-title" class="w-full max-w-md bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-2xl p-6 shadow-2xl space-y-5 animate-in fade-in zoom-in-95 duration-150">
      <!-- Header with Tabs -->
      <h2 id="auth-dialog-title" class="sr-only">Sign in or create a Treewiki account</h2>
      <div class="flex items-center justify-between border-b border-slate-100 dark:border-slate-800 pb-3">
        <div class="flex gap-4" role="tablist" aria-label="Authentication mode">
          <button
            @click="tab = 'login'"
            :class="tab === 'login' ? 'text-forest-600 dark:text-forest-400 border-b-2 border-forest-600 font-bold' : 'text-slate-500 font-medium'"
            class="pb-1 text-sm transition"
          >
            Sign In
          </button>
          <button
            @click="tab = 'register'"
            :class="tab === 'register' ? 'text-forest-600 dark:text-forest-400 border-b-2 border-forest-600 font-bold' : 'text-slate-500 font-medium'"
            class="pb-1 text-sm transition"
          >
            Create Account
          </button>
        </div>
        <button @click="close" class="text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 text-sm">✕</button>
      </div>

      <!-- Error Alert -->
      <div v-if="errorMsg" class="p-3 rounded-lg bg-rose-50 dark:bg-rose-950/40 border border-rose-200 dark:border-rose-900 text-xs text-rose-700 dark:text-rose-300">
        {{ errorMsg }}
      </div>

      <!-- Sign In Form -->
      <form v-if="tab === 'login'" @submit.prevent="handleLogin" class="space-y-4 text-xs">
        <div>
          <label class="block font-semibold text-slate-700 dark:text-slate-300 mb-1">Username</label>
          <input
            v-model="username"
            type="text"
            required
            autocomplete="username"
            class="w-full px-3 py-2 rounded-lg border border-slate-300 dark:border-slate-700 bg-slate-50 dark:bg-slate-950 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-forest-500"
          />
        </div>
        <div>
          <label class="block font-semibold text-slate-700 dark:text-slate-300 mb-1">Password</label>
          <input
            v-model="password"
            type="password"
            required
            autocomplete="current-password"
            class="w-full px-3 py-2 rounded-lg border border-slate-300 dark:border-slate-700 bg-slate-50 dark:bg-slate-950 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-forest-500"
          />
        </div>
        <button
          type="submit"
          :disabled="loading"
          class="w-full py-2.5 rounded-lg bg-forest-600 hover:bg-forest-500 text-white font-bold transition disabled:opacity-50"
        >
          {{ loading ? 'Signing in...' : 'Sign In' }}
        </button>
      </form>

      <!-- Register Form -->
      <form v-else @submit.prevent="handleRegister" class="space-y-4 text-xs">
        <div>
          <label class="block font-semibold text-slate-700 dark:text-slate-300 mb-1">Username</label>
          <input
            v-model="regUsername"
            type="text"
            required
            autocomplete="username"
            placeholder="3-32 alphanumeric characters"
            class="w-full px-3 py-2 rounded-lg border border-slate-300 dark:border-slate-700 bg-slate-50 dark:bg-slate-950 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-forest-500"
          />
        </div>
        <div>
          <label class="block font-semibold text-slate-700 dark:text-slate-300 mb-1">Email Address</label>
          <input
            v-model="regEmail"
            type="email"
            required
            autocomplete="email"
            placeholder="curator@example.com"
            class="w-full px-3 py-2 rounded-lg border border-slate-300 dark:border-slate-700 bg-slate-50 dark:bg-slate-950 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-forest-500"
          />
        </div>
        <div>
          <label class="block font-semibold text-slate-700 dark:text-slate-300 mb-1">Password</label>
          <input
            v-model="regPassword"
            type="password"
            required
            autocomplete="new-password"
            placeholder="Min 8 chars, uppercase, lowercase, number"
            class="w-full px-3 py-2 rounded-lg border border-slate-300 dark:border-slate-700 bg-slate-50 dark:bg-slate-950 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-forest-500"
          />
        </div>
        <label class="flex items-start gap-2 cursor-pointer text-slate-600 dark:text-slate-400">
          <input type="checkbox" v-model="acceptedTerms" required class="mt-0.5 rounded text-forest-600" />
          <span>
            I agree to the
            <NuxtLink to="/legal/terms" target="_blank" class="text-forest-600 dark:text-forest-400 hover:underline" @click.stop>Terms of Service</NuxtLink>
            and
            <NuxtLink to="/legal/privacy" target="_blank" class="text-forest-600 dark:text-forest-400 hover:underline" @click.stop>Privacy Policy</NuxtLink>,
            and I release my contributions under CC BY-SA 4.0.
          </span>
        </label>
        <button
          type="submit"
          :disabled="loading"
          class="w-full py-2.5 rounded-lg bg-forest-600 hover:bg-forest-500 text-white font-bold transition disabled:opacity-50"
        >
          {{ loading ? 'Creating account...' : 'Create Account' }}
        </button>
      </form>

      <!-- Social Login -->
      <div v-if="socialProviders.length" class="space-y-3 text-xs">
        <div class="flex items-center gap-3 text-slate-400 dark:text-slate-500">
          <span class="h-px flex-1 bg-slate-200 dark:bg-slate-800"></span>
          <span>or continue with</span>
          <span class="h-px flex-1 bg-slate-200 dark:bg-slate-800"></span>
        </div>
        <div class="grid grid-cols-2 gap-2">
          <button
            v-for="p in socialProviders"
            :key="p.name"
            type="button"
            @click="startOAuth(p.name)"
            class="py-2.5 rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-950 text-slate-700 dark:text-slate-200 font-semibold hover:bg-slate-50 dark:hover:bg-slate-800 transition"
          >
            {{ providerLabel(p.name) }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useApi } from '../composables/useApi'

const props = defineProps<{
  isOpen: boolean
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()

const tab = ref<'login' | 'register'>('login')
const authStore = useAuthStore()
const api = useApi()

const username = ref('')
const password = ref('')

const regUsername = ref('')
const regEmail = ref('')
const regPassword = ref('')
const acceptedTerms = ref(false)

const errorMsg = ref('')
const loading = ref(false)

const socialProviders = ref<Array<{ name: string; configured: boolean }>>([])

watch(
  () => props.isOpen,
  async (open) => {
    if (!open || socialProviders.value.length) return
    try {
      const res = await api.get<{ providers: Array<{ name: string; configured: boolean }> }>('/auth/oauth/providers')
      socialProviders.value = (res.providers || []).filter((p) => p.configured)
    } catch {
      socialProviders.value = []
    }
  },
  { immediate: true }
)

const providerLabel = (name: string) => name.charAt(0).toUpperCase() + name.slice(1)

const startOAuth = (provider: string) => {
  window.location.href = `${api.baseURL}/auth/oauth/${provider}/login`
}

const close = () => {
  errorMsg.value = ''
  emit('close')
}

const handleLogin = async () => {
  errorMsg.value = ''
  loading.value = true
  try {
    const res = await api.post('/auth/login', {
      username: username.value,
      password: password.value,
    })
    authStore.setAuth(res.user, res.access_token)
    close()
  } catch (err: any) {
    errorMsg.value = err.data?.error || err.message || 'Login failed. Please check credentials.'
  } finally {
    loading.value = false
  }
}

const handleRegister = async () => {
  errorMsg.value = ''
  loading.value = true
  try {
    const res = await api.post('/auth/register', {
      username: regUsername.value,
      email: regEmail.value,
      password: regPassword.value,
    })
    authStore.setAuth(res.user, res.access_token)
    close()
  } catch (err: any) {
    errorMsg.value = err.data?.error || err.message || 'Registration failed.'
  } finally {
    loading.value = false
  }
}
</script>

