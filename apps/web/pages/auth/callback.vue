<!-- apps/web/pages/auth/callback.vue -->
<template>
  <div class="max-w-md mx-auto py-10 text-center">
    <div v-if="status === 'loading'">
      <h1 class="font-wiki text-2xl text-slate-900 dark:text-slate-100">Signing you in…</h1>
      <p class="text-sm text-slate-500 mt-2" role="status">
        Completing sign-in<template v-if="provider"> with {{ providerLabel }}</template>…
      </p>
    </div>
    <div v-else>
      <h1 class="font-wiki text-2xl text-slate-900 dark:text-slate-100">Sign-in failed</h1>
      <p class="text-sm text-red-700 dark:text-red-300 mt-2">{{ errorMsg }}</p>
      <p class="mt-4">
        <NuxtLink to="/" class="text-sm text-wiki-link dark:text-sky-400 hover:underline">
          Back to the main page
        </NuxtLink>
      </p>
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
