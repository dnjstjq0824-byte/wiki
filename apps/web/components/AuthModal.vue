<!-- apps/web/components/AuthModal.vue -->
<template>
  <div v-if="isOpen" class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-slate-950/70">
    <div role="dialog" aria-modal="true" aria-labelledby="auth-dialog-title" class="w-full max-w-md bg-white dark:bg-slate-900 border border-wiki-border dark:border-slate-700 p-6 shadow-xl">
      <h2 id="auth-dialog-title" class="sr-only">Log in or create a Treewiki account</h2>
      <div class="flex items-center gap-0 text-sm border-b border-wiki-border dark:border-slate-700 -mb-px" role="tablist" aria-label="Authentication mode">
        <button
          @click="tab = 'login'"
          role="tab"
          :aria-selected="tab === 'login'"
          class="px-3 py-1 border border-b-0 -mb-px mr-1"
          :class="tab === 'login' ? 'bg-white dark:bg-slate-900 border-wiki-border dark:border-slate-700 font-semibold' : 'bg-slate-100 dark:bg-slate-800 border-slate-200 dark:border-slate-700 text-wiki-link dark:text-sky-400'"
        >
          Log in
        </button>
        <button
          @click="tab = 'register'"
          role="tab"
          :aria-selected="tab === 'register'"
          class="px-3 py-1 border border-b-0 -mb-px mr-1"
          :class="tab === 'register' ? 'bg-white dark:bg-slate-900 border-wiki-border dark:border-slate-700 font-semibold' : 'bg-slate-100 dark:bg-slate-800 border-slate-200 dark:border-slate-700 text-wiki-link dark:text-sky-400'"
        >
          Create account
        </button>
        <span class="flex-1"></span>
        <button @click="close" class="text-slate-400 hover:text-slate-600 px-1" aria-label="Close">✕</button>
      </div>

      <div class="pt-4 space-y-4 text-sm">
        <!-- Error Alert -->
        <div v-if="errorMsg" class="px-3 py-2 bg-red-50 dark:bg-red-950/40 border border-red-300 dark:border-red-900 text-red-700 dark:text-red-300">
          {{ errorMsg }}
        </div>

        <!-- Log In Form -->
        <form v-if="tab === 'login'" @submit.prevent="handleLogin" class="space-y-3">
          <div>
            <label class="block font-semibold mb-1">Username</label>
            <input
              v-model="username"
              type="text"
              required
              autocomplete="username"
              class="w-full px-2.5 py-1.5 border border-wiki-border dark:border-slate-700 bg-white dark:bg-slate-950 text-slate-900 dark:text-slate-100 focus:outline-none focus:border-namu-600"
            />
          </div>
          <div>
            <label class="block font-semibold mb-1">Password</label>
            <input
              v-model="password"
              type="password"
              required
              autocomplete="current-password"
              class="w-full px-2.5 py-1.5 border border-wiki-border dark:border-slate-700 bg-white dark:bg-slate-950 text-slate-900 dark:text-slate-100 focus:outline-none focus:border-namu-600"
            />
          </div>
          <button
            type="submit"
            :disabled="loading"
            class="w-full py-2 bg-namu-600 hover:bg-namu-700 text-white font-semibold disabled:opacity-50"
          >
            {{ loading ? 'Logging in…' : 'Log in' }}
          </button>
        </form>

        <!-- Register Form -->
        <form v-else @submit.prevent="handleRegister" class="space-y-3">
          <div>
            <label class="block font-semibold mb-1">Username</label>
            <input
              v-model="regUsername"
              type="text"
              required
              autocomplete="username"
              placeholder="3-32 alphanumeric characters"
              class="w-full px-2.5 py-1.5 border border-wiki-border dark:border-slate-700 bg-white dark:bg-slate-950 text-slate-900 dark:text-slate-100 focus:outline-none focus:border-namu-600"
            />
          </div>
          <div>
            <label class="block font-semibold mb-1">Email address</label>
            <input
              v-model="regEmail"
              type="email"
              required
              autocomplete="email"
              placeholder="curator@example.com"
              class="w-full px-2.5 py-1.5 border border-wiki-border dark:border-slate-700 bg-white dark:bg-slate-950 text-slate-900 dark:text-slate-100 focus:outline-none focus:border-namu-600"
            />
          </div>
          <div>
            <label class="block font-semibold mb-1">Password</label>
            <input
              v-model="regPassword"
              type="password"
              required
              autocomplete="new-password"
              placeholder="Min 8 chars, uppercase, lowercase, number"
              class="w-full px-2.5 py-1.5 border border-wiki-border dark:border-slate-700 bg-white dark:bg-slate-950 text-slate-900 dark:text-slate-100 focus:outline-none focus:border-namu-600"
            />
          </div>
          <label class="flex items-start gap-2 cursor-pointer text-[13px] text-slate-600 dark:text-slate-400">
            <input type="checkbox" v-model="acceptedTerms" required class="mt-1" />
            <span>
              I agree to the
              <NuxtLink to="/legal/terms" target="_blank" class="text-wiki-link dark:text-sky-400 hover:underline" @click.stop>Terms of Service</NuxtLink>
              and
              <NuxtLink to="/legal/privacy" target="_blank" class="text-wiki-link dark:text-sky-400 hover:underline" @click.stop>Privacy Policy</NuxtLink>,
              and I release my contributions under CC BY-SA 4.0.
            </span>
          </label>
          <button
            type="submit"
            :disabled="loading"
            class="w-full py-2 bg-namu-600 hover:bg-namu-700 text-white font-semibold disabled:opacity-50"
          >
            {{ loading ? 'Creating account…' : 'Create account' }}
          </button>
        </form>

        <!-- Social Login -->
        <div v-if="socialProviders.length" class="pt-3 border-t border-wiki-border dark:border-slate-700">
          <p class="text-[13px] text-slate-500 mb-2">Or continue with:</p>
          <div class="grid grid-cols-2 gap-2">
            <button
              v-for="p in socialProviders"
              :key="p.name"
              type="button"
              @click="startOAuth(p.name)"
              class="py-2 border border-wiki-border dark:border-slate-700 bg-white dark:bg-slate-950 font-semibold hover:bg-slate-100 dark:hover:bg-slate-800"
            >
              {{ providerLabel(p.name) }}
            </button>
          </div>
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
