<!-- apps/web/error.vue — global 404/500 page -->
<template>
  <div class="min-h-screen bg-wiki-pagebg dark:bg-slate-950 text-slate-900 dark:text-slate-100 font-sans flex flex-col">
    <header class="bg-white dark:bg-slate-900 border-b border-wiki-border dark:border-slate-800">
      <div class="max-w-[72rem] mx-auto px-5 py-2 flex items-center gap-2.5">
        <NuxtLink to="/" class="flex items-center gap-2.5">
          <img src="/logos/treewiki.svg" alt="" class="w-8 h-8" aria-hidden="true" />
          <span class="font-wiki text-xl">Treewiki</span>
        </NuxtLink>
      </div>
    </header>

    <main class="flex-1 max-w-[72rem] w-full mx-auto px-5 py-6">
      <div class="bg-white dark:bg-slate-900 border border-wiki-border dark:border-slate-800 px-6 py-8 max-w-xl">
        <p class="font-mono text-sm text-slate-400">{{ error?.statusCode || 500 }}</p>
        <h1 class="font-wiki text-3xl mt-1">
          {{ error?.statusCode === 404 ? 'Page not found' : 'Something went wrong' }}
        </h1>
        <p class="text-sm text-slate-600 dark:text-slate-400 mt-3 leading-relaxed">
          {{ error?.statusCode === 404
            ? 'This path does not exist on Treewiki. It may have been moved, renamed, or never created.'
            : 'The application hit an unexpected error. Please try again in a moment.' }}
        </p>
        <div class="flex items-center gap-4 mt-4 text-sm">
          <NuxtLink to="/" class="text-wiki-link dark:text-sky-400 hover:underline font-semibold">
            Main page
          </NuxtLink>
          <button @click="handleError" class="text-wiki-link dark:text-sky-400 hover:underline">
            {{ error?.statusCode === 404 ? 'Random article' : 'Try again' }}
          </button>
        </div>
      </div>
    </main>

    <footer class="border-t border-wiki-border dark:border-slate-800 py-4 text-center text-xs text-slate-500">
      Treewiki · CC BY-SA 4.0 ·
      <NuxtLink to="/legal/terms" class="text-wiki-link dark:text-sky-400 hover:underline">Terms</NuxtLink> ·
      <NuxtLink to="/legal/privacy" class="text-wiki-link dark:text-sky-400 hover:underline">Privacy</NuxtLink>
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
