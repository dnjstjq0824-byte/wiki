<template>
  <aside class="w-60 shrink-0 hidden lg:block text-sm border-r border-slate-200 dark:border-slate-800 pr-4 py-4 space-y-6">
    <!-- Portal Navigation -->
    <div>
      <h3 class="font-semibold text-xs text-slate-400 dark:text-slate-500 uppercase tracking-wider mb-2">Navigation</h3>
      <ul class="space-y-1">
        <li>
          <NuxtLink to="/" class="block px-2.5 py-1.5 rounded-md text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 font-medium">
            Main Portal
          </NuxtLink>
        </li>
        <li>
          <NuxtLink to="/tree" class="block px-2.5 py-1.5 rounded-md text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 font-medium">
            Topic Trees (DAG)
          </NuxtLink>
        </li>
        <li>
          <NuxtLink to="/special/random" class="block px-2.5 py-1.5 rounded-md text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 font-medium">
            Random Article
          </NuxtLink>
        </li>
      </ul>
    </div>

    <!-- Contribute -->
    <div>
      <h3 class="font-semibold text-xs text-slate-400 dark:text-slate-500 uppercase tracking-wider mb-2">Contribute</h3>
      <ul class="space-y-1">
        <li>
          <NuxtLink to="/special/recent-changes" class="block px-2.5 py-1.5 rounded-md text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 font-medium">
            Recent Changes
          </NuxtLink>
        </li>
        <li>
          <NuxtLink to="/w/Treewiki:Community_Portal" class="block px-2.5 py-1.5 rounded-md text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 font-medium">
            Community Guidelines
          </NuxtLink>
        </li>
        <li>
          <NuxtLink to="/w/Treewiki:Sandbox" class="block px-2.5 py-1.5 rounded-md text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 font-medium">
            Editing Sandbox
          </NuxtLink>
        </li>
      </ul>
    </div>

    <!-- Principles -->
    <div class="p-3 rounded-lg bg-forest-50 dark:bg-slate-800/60 border border-forest-200 dark:border-forest-900/50 text-xs text-slate-600 dark:text-slate-400">
      <p class="font-bold text-forest-800 dark:text-forest-400 mb-1">Knowledge Integrity</p>
      <p>All contributions are released under CC BY-SA 4.0 and must adhere to NPOV & verifiable sourcing.</p>
    </div>
  </aside>
</template>
