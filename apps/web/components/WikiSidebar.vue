<template>
  <nav aria-label="Site" class="w-44 shrink-0 hidden lg:block text-[13px] pr-3 py-1 space-y-4">
    <section aria-label="Navigation" class="border border-wiki-border dark:border-slate-800 bg-white dark:bg-slate-900">
      <h2 class="px-2.5 py-1 text-xs text-slate-500 border-b border-wiki-border dark:border-slate-800">Navigation</h2>
      <ul class="px-2.5 py-1.5 space-y-0.5">
        <li><NuxtLink to="/" class="text-wiki-link dark:text-sky-400 hover:underline">Main page</NuxtLink></li>
        <li><NuxtLink to="/tree" class="text-wiki-link dark:text-sky-400 hover:underline">Topic trees</NuxtLink></li>
        <li><NuxtLink to="/special/random" class="text-wiki-link dark:text-sky-400 hover:underline">Random article</NuxtLink></li>
      </ul>
    </section>

    <section aria-label="Contribute" class="border border-wiki-border dark:border-slate-800 bg-white dark:bg-slate-900">
      <h2 class="px-2.5 py-1 text-xs text-slate-500 border-b border-wiki-border dark:border-slate-800">Contribute</h2>
      <ul class="px-2.5 py-1.5 space-y-0.5">
        <li><NuxtLink to="/special/recent-changes" class="text-wiki-link dark:text-sky-400 hover:underline">Recent changes</NuxtLink></li>
        <li><NuxtLink to="/w/Treewiki:Community_Portal" class="text-wiki-link dark:text-sky-400 hover:underline">Community portal</NuxtLink></li>
        <li><NuxtLink to="/w/Treewiki:Sandbox" class="text-wiki-link dark:text-sky-400 hover:underline">Sandbox</NuxtLink></li>
      </ul>
    </section>

    <section v-if="isAdmin" aria-label="Administration" class="border border-wiki-border dark:border-slate-800 bg-white dark:bg-slate-900">
      <h2 class="px-2.5 py-1 text-xs text-slate-500 border-b border-wiki-border dark:border-slate-800">Administration</h2>
      <ul class="px-2.5 py-1.5 space-y-0.5">
        <li><NuxtLink to="/admin" class="text-wiki-link dark:text-sky-400 hover:underline">Dashboard</NuxtLink></li>
      </ul>
    </section>

    <section aria-label="Legal" class="border border-wiki-border dark:border-slate-800 bg-white dark:bg-slate-900">
      <h2 class="px-2.5 py-1 text-xs text-slate-500 border-b border-wiki-border dark:border-slate-800">Legal</h2>
      <ul class="px-2.5 py-1.5 space-y-0.5">
        <li><NuxtLink to="/legal/terms" class="text-wiki-link dark:text-sky-400 hover:underline">Terms of Service</NuxtLink></li>
        <li><NuxtLink to="/legal/privacy" class="text-wiki-link dark:text-sky-400 hover:underline">Privacy Policy</NuxtLink></li>
      </ul>
    </section>

    <section aria-label="License" class="border border-wiki-border dark:border-slate-800 bg-wiki-notice dark:bg-slate-900 px-2.5 py-2 text-xs text-slate-600 dark:text-slate-400">
      <p class="font-semibold text-slate-700 dark:text-slate-300">Free content</p>
      <p>CC BY-SA 4.0 · NPOV · Verifiable sources.</p>
    </section>
  </nav>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useAuthStore } from '../stores/auth'

const authStore = useAuthStore()
const isAdmin = computed(() =>
  ['SYSOP', 'OVERSIGHT', 'BUREAUCRAT', 'SYSTEM_ADMIN'].includes(authStore.userRole)
)
</script>
