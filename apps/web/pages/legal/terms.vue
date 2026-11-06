<!-- apps/web/pages/legal/terms.vue -->
<template>
  <div class="max-w-3xl">
    <h1 class="font-wiki text-3xl text-slate-900 dark:text-slate-100 border-b border-wiki-border dark:border-slate-700 pb-1">Terms of Service</h1>
    <p class="text-sm text-slate-500 mt-1">Last updated: November 2026 · Treewiki Collaborative Knowledge Platform</p>

    <section class="mt-4 text-[15px] text-slate-800 dark:text-slate-200 leading-relaxed">
      <h2 class="font-wiki text-xl border-b border-wiki-border dark:border-slate-700 pb-0.5 mt-5">1. The free knowledge license</h2>
      <p class="mt-2">All article text you contribute to Treewiki is released under the <strong>Creative Commons Attribution-ShareAlike 4.0 International (CC BY-SA 4.0)</strong> license. By pressing “Publish Revision” you agree that your contribution may be freely copied, modified, and redistributed, even commercially, provided attribution is preserved and derivatives carry the same license.</p>
    </section>

    <section class="mt-4 text-[15px] text-slate-800 dark:text-slate-200 leading-relaxed">
      <h2 class="font-wiki text-xl border-b border-wiki-border dark:border-slate-700 pb-0.5 mt-5">2. Editorial standards</h2>
      <p class="mt-2">Contributions must respect a <strong>neutral point of view (NPOV)</strong>, cite verifiable sources, and avoid original research, advertising, defamation, and copyrighted text pasted without permission. Curators and administrators may revert, protect, or remove content that violates these standards.</p>
    </section>

    <section class="mt-4 text-[15px] text-slate-800 dark:text-slate-200 leading-relaxed">
      <h2 class="font-wiki text-xl border-b border-wiki-border dark:border-slate-700 pb-0.5 mt-5">3. Conduct</h2>
      <p class="mt-2">Be civil on talk pages and in edit summaries. Harassment, hate speech, doxxing, spam, vandalism, and attempts to disrupt the service (including abuse-filter evasion) may lead to warnings, editing blocks, or account termination following review.</p>
    </section>

    <section class="mt-4 text-[15px] text-slate-800 dark:text-slate-200 leading-relaxed">
      <h2 class="font-wiki text-xl border-b border-wiki-border dark:border-slate-700 pb-0.5 mt-5">4. Accounts and security</h2>
      <p class="mt-2">You are responsible for activity under your account. Choose a strong password, do not share credentials, and enable two-factor authentication where available. We may lock accounts showing signs of compromise until ownership is verified.</p>
    </section>

    <section class="mt-4 text-[15px] text-slate-800 dark:text-slate-200 leading-relaxed">
      <h2 class="font-wiki text-xl border-b border-wiki-border dark:border-slate-700 pb-0.5 mt-5">5. Copyright complaints</h2>
      <p class="mt-2">Rights holders may file a takedown request identifying the infringing page and their authority. Valid requests are reviewed by administrators; repeat infringers lose editing privileges. Counter-notices follow the same channel.</p>
    </section>

    <section class="mt-4 text-[15px] text-slate-800 dark:text-slate-200 leading-relaxed">
      <h2 class="font-wiki text-xl border-b border-wiki-border dark:border-slate-700 pb-0.5 mt-5">6. Disclaimer</h2>
      <p class="mt-2">Treewiki content is provided “as is” without warranty of accuracy or fitness for any purpose. Always verify critical information — especially medical, legal, or safety topics — against primary sources and professionals.</p>
    </section>

    <p class="mt-6 text-sm">
      <NuxtLink to="/legal/privacy" class="text-wiki-link dark:text-sky-400 hover:underline">Privacy Policy</NuxtLink>
      ·
      <NuxtLink to="/" class="text-wiki-link dark:text-sky-400 hover:underline">Main page</NuxtLink>
    </p>
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
