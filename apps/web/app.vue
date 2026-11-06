<template>
  <div class="min-h-screen bg-wiki-pagebg dark:bg-slate-950 text-slate-900 dark:text-slate-100 font-sans flex flex-col">
    <a
      href="#main-content"
      class="sr-only focus:not-sr-only focus:absolute focus:top-2 focus:left-2 focus:z-[100] focus:px-4 focus:py-2 focus:bg-white focus:text-wiki-link focus:text-xs focus:font-bold"
    >
      Skip to main content
    </a>
    <WikiHeader />
    <div class="max-w-5xl mx-auto px-2 sm:px-3 py-3 w-full flex-1">
      <main id="main-content" class="bg-white dark:bg-slate-900 border border-wiki-border dark:border-slate-800 px-4 py-4 md:px-6 md:py-5" tabindex="-1">
        <NuxtPage />
      </main>
    </div>
    <footer class="bg-namu-800 dark:bg-slate-900 text-slate-200 dark:text-slate-400 mt-2">
      <div class="max-w-5xl mx-auto px-4 py-5 text-xs space-y-2">
        <p>
          Text is available under the <span class="font-semibold text-white dark:text-slate-200">Creative Commons Attribution-ShareAlike 4.0 License</span>;
          additional terms may apply. Contributions must follow a neutral point of view with verifiable sources.
        </p>
        <div class="flex flex-wrap items-center justify-between gap-3">
          <div class="flex items-center gap-2">
            <img src="/logos/treewiki.svg" alt="" class="w-4 h-4" aria-hidden="true" />
            <span>Treewiki · The free knowledge platform</span>
          </div>
          <nav aria-label="Legal and about">
            <ul class="flex items-center gap-3">
              <li><NuxtLink to="/legal/terms" class="text-namu-100 dark:text-sky-400 hover:underline">Terms of Service</NuxtLink></li>
              <li><NuxtLink to="/legal/privacy" class="text-namu-100 dark:text-sky-400 hover:underline">Privacy Policy</NuxtLink></li>
              <li><NuxtLink to="/special/recent-changes" class="text-namu-100 dark:text-sky-400 hover:underline">Recent Changes</NuxtLink></li>
            </ul>
          </nav>
        </div>
      </div>
    </footer>
  </div>
</template>
