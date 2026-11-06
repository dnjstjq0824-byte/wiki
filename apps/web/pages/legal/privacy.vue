<!-- apps/web/pages/legal/privacy.vue -->
<template>
  <div class="max-w-3xl mx-auto">
    <div class="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl p-6 md:p-8 shadow-sm space-y-6">
      <div class="border-b border-slate-200 dark:border-slate-800 pb-4">
        <h1 class="text-2xl font-bold text-slate-900 dark:text-slate-100">Privacy Policy</h1>
        <p class="text-xs text-slate-500 mt-1">Last updated: November 2026 · Treewiki Collaborative Knowledge Platform</p>
      </div>

      <section class="space-y-2 text-sm text-slate-700 dark:text-slate-300 leading-relaxed">
        <h2 class="font-bold text-slate-900 dark:text-slate-100">1. What we collect</h2>
        <ul class="list-disc list-inside space-y-1">
          <li><strong>Account data:</strong> username, email address, and password hash (Argon2id — we never store plain passwords).</li>
          <li><strong>Public contributions:</strong> every edit, comment, and vote is attributed to your username in the permanent revision ledger.</li>
          <li><strong>Operational logs:</strong> request metadata retained briefly for security, abuse prevention, and reliability.</li>
        </ul>
      </section>

      <section class="space-y-2 text-sm text-slate-700 dark:text-slate-300 leading-relaxed">
        <h2 class="font-bold text-slate-900 dark:text-slate-100">2. Anonymous editing</h2>
        <p>You may read without an account. Anonymous edits are attributed to a <strong>masked network identifier</strong> (never your full IP address) to balance attribution with privacy.</p>
      </section>

      <section class="space-y-2 text-sm text-slate-700 dark:text-slate-300 leading-relaxed">
        <h2 class="font-bold text-slate-900 dark:text-slate-100">3. Cookies and sessions</h2>
        <p>We use a strictly-necessary session cookie to keep you signed in, plus local preferences (theme). We run no advertising trackers and sell no personal data.</p>
      </section>

      <section class="space-y-2 text-sm text-slate-700 dark:text-slate-300 leading-relaxed">
        <h2 class="font-bold text-slate-900 dark:text-slate-100">4. Your rights</h2>
        <p>You may update your profile, change or reset credentials, and request account anonymization (public edit history stays attributed to a pseudonym, since the CC BY-SA license requires attribution). Contact the administrators through any talk page to exercise these rights.</p>
      </section>

      <section class="space-y-2 text-sm text-slate-700 dark:text-slate-300 leading-relaxed">
        <h2 class="font-bold text-slate-900 dark:text-slate-100">5. Data retention and security</h2>
        <p>Credentials are hashed, sessions expire, and administrative actions are logged. Backups rotate on a fixed schedule. In the event of a breach affecting account data, we will notify affected users and describe remediation steps.</p>
      </section>

      <div class="pt-2 border-t border-slate-100 dark:border-slate-800 flex items-center gap-3 text-xs">
        <NuxtLink to="/legal/terms" class="font-semibold text-forest-600 dark:text-forest-400 hover:underline">Terms of Service →</NuxtLink>
        <NuxtLink to="/" class="font-semibold text-slate-500 hover:underline">Back to Portal</NuxtLink>
      </div>
    </div>
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
