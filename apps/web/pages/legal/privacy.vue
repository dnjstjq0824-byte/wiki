<!-- apps/web/pages/legal/privacy.vue -->
<template>
  <div class="max-w-3xl">
    <h1 class="font-wiki text-3xl text-slate-900 dark:text-slate-100 border-b border-wiki-border dark:border-slate-700 pb-1">Privacy Policy</h1>
    <p class="text-sm text-slate-500 mt-1">Last updated: November 2026 · Treewiki Collaborative Knowledge Platform</p>

    <section class="mt-4 text-[15px] text-slate-800 dark:text-slate-200 leading-relaxed">
      <h2 class="font-wiki text-xl border-b border-wiki-border dark:border-slate-700 pb-0.5 mt-5">1. What we collect</h2>
      <ul class="list-disc list-inside space-y-1 mt-2">
        <li><strong>Account data:</strong> username, email address, and password hash (Argon2id — we never store plain passwords).</li>
        <li><strong>Public contributions:</strong> every edit, comment, and vote is attributed to your username in the permanent revision ledger.</li>
        <li><strong>Operational logs:</strong> request metadata retained briefly for security, abuse prevention, and reliability.</li>
      </ul>
    </section>

    <section class="mt-4 text-[15px] text-slate-800 dark:text-slate-200 leading-relaxed">
      <h2 class="font-wiki text-xl border-b border-wiki-border dark:border-slate-700 pb-0.5 mt-5">2. Anonymous editing</h2>
      <p class="mt-2">You may read without an account. Anonymous edits are attributed to a <strong>masked network identifier</strong> (never your full IP address) to balance attribution with privacy.</p>
    </section>

    <section class="mt-4 text-[15px] text-slate-800 dark:text-slate-200 leading-relaxed">
      <h2 class="font-wiki text-xl border-b border-wiki-border dark:border-slate-700 pb-0.5 mt-5">3. Cookies and sessions</h2>
      <p class="mt-2">We use a strictly-necessary session cookie to keep you signed in, plus local preferences (theme). We run no advertising trackers and sell no personal data.</p>
    </section>

    <section class="mt-4 text-[15px] text-slate-800 dark:text-slate-200 leading-relaxed">
      <h2 class="font-wiki text-xl border-b border-wiki-border dark:border-slate-700 pb-0.5 mt-5">4. Your rights</h2>
      <p class="mt-2">You may update your profile, change or reset credentials, and request account anonymization (public edit history stays attributed to a pseudonym, since the CC BY-SA license requires attribution). Contact the administrators through any talk page to exercise these rights.</p>
    </section>

    <section class="mt-4 text-[15px] text-slate-800 dark:text-slate-200 leading-relaxed">
      <h2 class="font-wiki text-xl border-b border-wiki-border dark:border-slate-700 pb-0.5 mt-5">5. Data retention and security</h2>
      <p class="mt-2">Credentials are hashed, sessions expire, and administrative actions are logged. Backups rotate on a fixed schedule. In the event of a breach affecting account data, we will notify affected users and describe remediation steps.</p>
    </section>

    <p class="mt-6 text-sm">
      <NuxtLink to="/legal/terms" class="text-wiki-link dark:text-sky-400 hover:underline">Terms of Service</NuxtLink>
      ·
      <NuxtLink to="/" class="text-wiki-link dark:text-sky-400 hover:underline">Main page</NuxtLink>
    </p>
  </div>
</template>

<script setup lang="ts">
useSeoMeta({
  title: 'Privacy Policy — Treewiki',
  description: 'Privacy Policy for Treewiki: data collection, anonymous editing, cookies, rights, and retention.',
  ogTitle: 'Privacy Policy — Treewiki',
  ogType: 'website',
})
</script>
