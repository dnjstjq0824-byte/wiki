<template>
  <div class="min-h-screen bg-slate-50 dark:bg-slate-950 text-slate-900 dark:text-slate-100 font-sans transition-colors">
    <WikiHeader />
    <div class="max-w-7xl mx-auto px-4 py-6 flex gap-6">
      <WikiSidebar />
      <main class="flex-1 min-w-0">
        <NuxtPage />
      </main>
    </div>
  </div>
</template>

