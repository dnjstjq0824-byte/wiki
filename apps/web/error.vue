<!-- apps/web/error.vue — global 404/500 page -->
<template>
  <div class="min-h-screen bg-slate-50 dark:bg-slate-950 text-slate-900 dark:text-slate-100 font-sans flex flex-col">
    <header class="bg-white/95 dark:bg-slate-900/95 border-b border-slate-200 dark:border-slate-800">
      <div class="max-w-7xl mx-auto px-4 h-14 flex items-center gap-2.5">
        <NuxtLink to="/" class="flex items-center gap-2.5 font-bold text-lg">
          <img src="/logos/treewiki.svg" alt="Treewiki" class="w-8 h-8 rounded" />
          <span class="tracking-tight text-forest-700 dark:text-forest-500">Tree<span class="text-slate-700 dark:text-slate-200">wiki</span></span>
        </NuxtLink>
      </div>
    </header>

    <main class="flex-1 flex items-center justify-center px-4 py-16">
      <div class="max-w-md w-full bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl p-8 shadow-sm text-center space-y-4">
        <div class="text-5xl">{{ error?.statusCode === 404 ? '🧭' : '🛠️' }}</div>
        <p class="text-5xl font-extrabold font-mono text-slate-200 dark:text-slate-700">{{ error?.statusCode || 500 }}</p>
        <h1 class="text-xl font-bold text-slate-900 dark:text-slate-100">
          {{ error?.statusCode === 404 ? 'Page not found' : 'Something went wrong' }}
        </h1>
        <p class="text-xs text-slate-500 leading-relaxed">
          {{ error?.statusCode === 404
            ? 'This path does not exist on Treewiki. It may have been moved, renamed, or never created.'
            : 'The application hit an unexpected error. Our team has been notified — please try again in a moment.' }}
        </p>
        <div class="flex items-center justify-center gap-2 pt-2">
          <NuxtLink
            to="/"
            class="px-5 py-2 rounded-lg bg-forest-600 hover:bg-forest-500 text-white font-bold text-xs transition"
          >
            Back to Portal
          </NuxtLink>
          <button
            @click="handleError"
            class="px-5 py-2 rounded-lg border border-slate-200 dark:border-slate-700 text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-800 font-bold text-xs transition"
          >
            {{ error?.statusCode === 404 ? 'Random Article' : 'Try Again' }}
          </button>
        </div>
      </div>
    </main>

    <footer class="border-t border-slate-200 dark:border-slate-800 py-4 text-center text-[11px] text-slate-400">
      Treewiki · CC BY-SA 4.0 ·
      <NuxtLink to="/legal/terms" class="hover:underline">Terms</NuxtLink> ·
      <NuxtLink to="/legal/privacy" class="hover:underline">Privacy</NuxtLink>
    </footer>
  </div>
</template>

<script setup lang="ts">
// Structural error shape (avoids the '#app' alias so typecheck stays green
// outside a fully generated .nuxt directory).
interface PageError {
  statusCode?: number
  message?: string
}

const props = defineProps<{
  error?: PageError | null
}>()

const handleError = () => {
  if (props.error?.statusCode === 404) {
    navigateTo('/special/random')
  } else {
    clearError({ redirect: '/' })
  }
}

useSeoMeta({
  title: 'Page not found — Treewiki',
  robots: 'noindex',
})
</script>
