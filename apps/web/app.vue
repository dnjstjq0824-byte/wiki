<template>
  <div class="min-h-screen bg-slate-50 dark:bg-slate-950 text-slate-900 dark:text-slate-100 font-sans transition-colors flex flex-col">
    <a
      href="#main-content"
      class="sr-only focus:not-sr-only focus:absolute focus:top-2 focus:left-2 focus:z-[100] focus:px-4 focus:py-2 focus:rounded-lg focus:bg-forest-600 focus:text-white focus:text-xs focus:font-bold"
    >
      Skip to main content
    </a>
    <WikiHeader />
    <div class="max-w-7xl mx-auto px-4 py-6 flex gap-6 w-full flex-1">
      <WikiSidebar />
      <main id="main-content" class="flex-1 min-w-0" tabindex="-1">
        <NuxtPage />
      </main>
    </div>
    <footer class="border-t border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900">
      <div class="max-w-7xl mx-auto px-4 py-5 flex flex-wrap items-center justify-between gap-3 text-[11px] text-slate-500">
        <div class="flex items-center gap-2">
          <img src="/logos/treewiki.svg" alt="" class="w-5 h-5 rounded" aria-hidden="true" />
          <span>Treewiki · Free knowledge under <span class="font-semibold">CC BY-SA 4.0</span></span>
        </div>
        <nav aria-label="Legal and about" class="flex items-center gap-4">
          <NuxtLink to="/legal/terms" class="hover:text-forest-600 dark:hover:text-forest-400 hover:underline">Terms of Service</NuxtLink>
          <NuxtLink to="/legal/privacy" class="hover:text-forest-600 dark:hover:text-forest-400 hover:underline">Privacy Policy</NuxtLink>
          <NuxtLink to="/special/recent-changes" class="hover:text-forest-600 dark:hover:text-forest-400 hover:underline">Recent Changes</NuxtLink>
        </nav>
      </div>
    </footer>
  </div>
</template>
