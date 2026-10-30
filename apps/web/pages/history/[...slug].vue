<!-- apps/web/pages/history/[...slug].vue -->
<template>
  <div class="space-y-6">
    <div class="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl p-6 shadow-sm space-y-5">
      <!-- Header -->
      <div class="flex flex-wrap items-center justify-between border-b border-slate-200 dark:border-slate-800 pb-4 gap-4">
        <div>
          <h1 class="text-2xl font-bold text-slate-900 dark:text-slate-100 flex items-center gap-2">
            <span>🕒</span> Revision History: <span class="text-forest-700 dark:text-forest-400">{{ displayTitle }}</span>
          </h1>
          <p class="text-xs text-slate-500">Immutable revision ledger with cryptographic SHA-256 integrity</p>
        </div>

        <!-- Action Links -->
        <div class="flex items-center gap-2 text-xs font-semibold">
          <NuxtLink :to="`/w/${slug}`" class="px-3 py-1.5 rounded-lg border border-slate-200 dark:border-slate-700 hover:bg-slate-50 dark:hover:bg-slate-800 text-slate-700 dark:text-slate-300">
            ← Back to Article
          </NuxtLink>
          <NuxtLink :to="`/edit/${slug}`" class="px-3 py-1.5 rounded-lg bg-forest-600 hover:bg-forest-500 text-white">
            ✏️ Edit Page
          </NuxtLink>
        </div>
      </div>

      <!-- Compare Button Bar -->
      <div class="flex items-center justify-between p-3 rounded-lg bg-slate-50 dark:bg-slate-800/60 border border-slate-200 dark:border-slate-700 text-xs">
        <span class="text-slate-600 dark:text-slate-300">
          Select any two revisions using the radio buttons to inspect differences.
        </span>
        <button
          @click="showInlineDiff = !showInlineDiff"
          class="px-4 py-1.5 rounded-lg bg-forest-600 hover:bg-forest-500 text-white font-bold transition shadow-xs"
        >
          {{ showInlineDiff ? 'Hide Diff Viewer' : 'Compare Selected Revisions' }}
        </button>
      </div>

      <!-- Inline Diff Viewer (if toggled) -->
      <div v-if="showInlineDiff" class="pt-2 pb-4">
        <DiffViewer />
      </div>

      <!-- Revision List Table -->
      <div class="border border-slate-200 dark:border-slate-800 rounded-xl overflow-hidden divide-y divide-slate-100 dark:divide-slate-800">
        <div
          v-for="rev in revisions"
          :key="rev.id"
          class="p-4 flex flex-wrap items-center justify-between gap-4 hover:bg-slate-50/80 dark:hover:bg-slate-800/40 transition text-xs"
        >
          <!-- Left: Selection Radios & Metadata -->
          <div class="flex items-center gap-3">
            <div class="flex items-center gap-1.5">
              <input type="radio" :value="rev.id" v-model="selectedBase" name="base" title="Base revision" />
              <input type="radio" :value="rev.id" v-model="selectedTarget" name="target" title="Target revision" />
            </div>

            <!-- Timestamp & Hash -->
            <div>
              <div class="flex items-center gap-2">
                <span class="font-bold text-slate-800 dark:text-slate-200">{{ rev.timeAgo }}</span>
                <span class="font-mono text-[10px] text-slate-400">#{{ rev.sha256.substring(0, 8) }}</span>
                <span v-if="rev.isLatest" class="px-1.5 py-0.5 rounded text-[10px] font-bold bg-forest-100 dark:bg-forest-950 text-forest-700 dark:text-forest-400">
                  CURRENT
                </span>
              </div>
              <p class="text-slate-500 text-[11px] mt-0.5">
                by <strong class="text-slate-700 dark:text-slate-300">{{ rev.author }}</strong>
                <span v-if="rev.isIP" class="text-slate-400 font-mono"> ({{ rev.maskedIP }})</span>
              </p>
            </div>
          </div>

          <!-- Middle: Summary Note -->
          <div class="flex-1 max-w-md text-slate-600 dark:text-slate-300 italic text-[11px]">
            "{{ rev.summary || 'No edit summary provided' }}"
          </div>

          <!-- Right: Byte Delta & Rollback -->
          <div class="flex items-center gap-3">
            <span
              class="font-mono font-semibold"
              :class="rev.delta > 0 ? 'text-emerald-600 dark:text-emerald-400' : rev.delta < 0 ? 'text-rose-600 dark:text-rose-400' : 'text-slate-400'"
            >
              {{ rev.delta > 0 ? `+${rev.delta}` : rev.delta }} bytes
            </span>

            <button
              v-if="!rev.isLatest"
              @click="handleRollback(rev.id)"
              class="px-2.5 py-1 rounded bg-slate-100 dark:bg-slate-800 hover:bg-rose-100 hover:text-rose-700 dark:hover:bg-rose-950 text-slate-600 dark:text-slate-300 text-[11px] font-semibold transition"
              title="Revert back to this clean state"
            >
              ↺ Rollback
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()

const slug = computed(() => {
  const p = route.params.slug
  return Array.isArray(p) ? p.join('/') : p || 'Main_Page'
})

const displayTitle = computed(() => slug.value.replace(/_/g, ' '))

const selectedBase = ref('rev-2')
const selectedTarget = ref('rev-1')
const showInlineDiff = ref(false)

const revisions = ref([
  {
    id: 'rev-1',
    timeAgo: '14 minutes ago',
    sha256: 'a1b2c3d4e5f67890123456789abcdef0123456789abcdef0123456789abcdef0',
    author: 'CuratorAlpha',
    isIP: false,
    summary: 'Expanded definitions of quantum gates and asymptotic complexity',
    delta: 340,
    isLatest: true,
  },
  {
    id: 'rev-2',
    timeAgo: '2 hours ago',
    sha256: 'e8f7a6b5c4d32109876543210fedcba9876543210fedcba9876543210fedcba9',
    author: 'Anonymous Contributor',
    isIP: true,
    maskedIP: '198.51.100.xxx',
    summary: 'Added bibliography citations for Shor and Grover',
    delta: 120,
    isLatest: false,
  },
  {
    id: 'rev-3',
    timeAgo: '1 day ago',
    sha256: '0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef',
    author: 'FoundingEditor',
    isIP: false,
    summary: 'Initial draft article creation with core taxonomy',
    delta: 1450,
    isLatest: false,
  },
])

const handleRollback = (revId: string) => {
  if (confirm(`Are you sure you want to rollback to revision #${revId}? This will append a new reversal revision.`)) {
    alert('Fast rollback executed successfully!')
  }
}
</script>
