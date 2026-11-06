<!-- apps/web/pages/legal/terms.vue -->
<template>
  <div class="max-w-3xl mx-auto">
    <div class="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl p-6 md:p-8 shadow-sm space-y-6">
      <div class="border-b border-slate-200 dark:border-slate-800 pb-4">
        <h1 class="text-2xl font-bold text-slate-900 dark:text-slate-100">Terms of Service</h1>
        <p class="text-xs text-slate-500 mt-1">Last updated: November 2026 · Treewiki Collaborative Knowledge Platform</p>
      </div>

      <section class="space-y-2 text-sm text-slate-700 dark:text-slate-300 leading-relaxed">
        <h2 class="font-bold text-slate-900 dark:text-slate-100">1. The free knowledge license</h2>
        <p>All article text you contribute to Treewiki is released under the <strong>Creative Commons Attribution-ShareAlike 4.0 International (CC BY-SA 4.0)</strong> license. By pressing “Publish Revision” you agree that your contribution may be freely copied, modified, and redistributed, even commercially, provided attribution is preserved and derivatives carry the same license.</p>
      </section>

      <section class="space-y-2 text-sm text-slate-700 dark:text-slate-300 leading-relaxed">
        <h2 class="font-bold text-slate-900 dark:text-slate-100">2. Editorial standards</h2>
        <p>Contributions must respect a <strong>neutral point of view (NPOV)</strong>, cite verifiable sources, and avoid original research, advertising, defamation, and copyrighted text pasted without permission. Curators and administrators may revert, protect, or remove content that violates these standards.</p>
      </section>

      <section class="space-y-2 text-sm text-slate-700 dark:text-slate-300 leading-relaxed">
        <h2 class="font-bold text-slate-900 dark:text-slate-100">3. Conduct</h2>
        <p>Be civil on talk pages and in edit summaries. Harassment, hate speech, doxxing, spam, vandalism, and attempts to disrupt the service (including abuse-filter evasion) may lead to warnings, editing blocks, or account termination following review.</p>
      </section>

      <section class="space-y-2 text-sm text-slate-700 dark:text-slate-300 leading-relaxed">
        <h2 class="font-bold text-slate-900 dark:text-slate-100">4. Accounts and security</h2>
        <p>You are responsible for activity under your account. Choose a strong password, do not share credentials, and enable two-factor authentication where available. We may lock accounts showing signs of compromise until ownership is verified.</p>
      </section>

      <section class="space-y-2 text-sm text-slate-700 dark:text-slate-300 leading-relaxed">
        <h2 class="font-bold text-slate-900 dark:text-slate-100">5. Copyright complaints</h2>
        <p>Rights holders may file a takedown request identifying the infringing page and their authority. Valid requests are reviewed by administrators; repeat infringers lose editing privileges. Counter-notices follow the same channel.</p>
      </section>

      <section class="space-y-2 text-sm text-slate-700 dark:text-slate-300 leading-relaxed">
        <h2 class="font-bold text-slate-900 dark:text-slate-100">6. Disclaimer</h2>
        <p>Treewiki content is provided “as is” without warranty of accuracy or fitness for any purpose. Always verify critical information — especially medical, legal, or safety topics — against primary sources and professionals.</p>
      </section>

      <div class="pt-2 border-t border-slate-100 dark:border-slate-800 flex items-center gap-3 text-xs">
        <NuxtLink to="/legal/privacy" class="font-semibold text-forest-600 dark:text-forest-400 hover:underline">Privacy Policy →</NuxtLink>
        <NuxtLink to="/" class="font-semibold text-slate-500 hover:underline">Back to Portal</NuxtLink>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
useSeoMeta({
  title: 'Terms of Service — Treewiki',
  description: 'Terms of Service for Treewiki: free knowledge license, editorial standards, conduct, and copyright complaints.',
  ogTitle: 'Terms of Service — Treewiki',
  ogType: 'website',
})
</script>
