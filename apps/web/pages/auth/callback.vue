<!-- apps/web/pages/auth/callback.vue -->
<template>
  <div class="min-h-[60vh] flex items-center justify-center p-4">
    <div class="w-full max-w-sm bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-2xl p-6 shadow-xl text-center space-y-4">
      <div v-if="status === 'loading'" class="space-y-3">
        <div class="mx-auto w-8 h-8 rounded-full border-2 border-forest-600 border-t-transparent animate-spin" aria-hidden="true"></div>
        <p class="text-sm text-slate-600 dark:text-slate-300" role="status">
          Completing sign-in<template v-if="provider"> with {{ providerLabel }}</template>...
        </p>
      </div>
      <div v-else class="space-y-3">
        <h1 class="text-base font-bold text-slate-900 dark:text-slate-100">Sign-in failed</h1>
        <p class="text-xs text-rose-700 dark:text-rose-300">{{ errorMsg }}</p>
        <NuxtLink to="/" class="inline-block text-xs font-semibold text-forest-600 dark:text-forest-400 hover:underline">
          Back to home
        </NuxtLink>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useAuthStore } from '../../stores/auth'
import { useApi } from '../../composables/useApi'

useSeoMeta({
  title: 'Completing sign-in — Treewiki',
  robots: 'noindex',
})

const route = useRoute()
const api = useApi()
const authStore = useAuthStore()

const status = ref<'loading' | 'error'>('loading')
const errorMsg = ref('')

const provider = computed(() => {
  const p = route.query.provider
  return typeof p === 'string' && p ? p : ''
})
const providerLabel = computed(() => (provider.value ? provider.value.charAt(0).toUpperCase() + provider.value.slice(1) : ''))

onMounted(async () => {
  const q = route.query
  if (typeof q.error === 'string' && q.error) {
    status.value = 'error'
    errorMsg.value = q.error
    return
  }
  const code = q.code
  if (typeof code !== 'string' || !code) {
    status.value = 'error'
    errorMsg.value = 'Sign-in was not completed. Please try again.'
    return
  }
  try {
    const res = await api.post('/auth/oauth/consume', { code })
    authStore.setAuth(res.user, res.access_token)
    await navigateTo('/')
  } catch (err: any) {
    status.value = 'error'
    errorMsg.value = err.data?.error || err.message || 'Could not complete social sign-in.'
  }
})
</script>
