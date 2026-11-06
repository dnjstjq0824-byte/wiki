<!-- apps/web/pages/history/[...slug].vue -->
<template>
  <div class="space-y-6 max-w-5xl mx-auto">
    <div class="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl p-6 shadow-sm space-y-5">
      <!-- Header -->
      <div class="flex flex-wrap items-center justify-between border-b border-slate-200 dark:border-slate-800 pb-4 gap-4">
        <div>
          <h1 class="text-2xl font-bold text-slate-900 dark:text-slate-100 flex items-center gap-2">
            <span>🕒</span> Revision History: <span class="text-forest-700 dark:text-forest-400">{{ displayTitle }}</span>
          </h1>
          <p class="text-xs text-slate-500">Immutable revision ledger with cryptographic Myers diff comparison</p>
        </div>

        <!-- Action Links -->
        <div class="flex items-center gap-2 text-xs font-semibold">
          <NuxtLink :to="`/w/${slug}`" class="px-3 py-1.5 rounded-lg border border-slate-200 dark:border-slate-700 hover:bg-slate-50 dark:hover:bg-slate-800 text-slate-700 dark:text-slate-300">
            ← Back to Article
          </NuxtLink>
          <NuxtLink :to="`/talk/${slug}`" class="px-3 py-1.5 rounded-lg border border-slate-200 dark:border-slate-700 hover:bg-slate-50 dark:hover:bg-slate-800 text-slate-700 dark:text-slate-300">
            💬 Talk
          </NuxtLink>
          <NuxtLink :to="`/edit/${slug}`" class="px-3 py-1.5 rounded-lg bg-forest-600 hover:bg-forest-500 text-white">
            ✏️ Edit Page
          </NuxtLink>
        </div>
      </div>

      <!-- Notification Toast -->
      <div
        v-if="statusMessage"
        class="p-3 rounded-lg text-xs font-semibold flex items-center justify-between"
        :class="statusIsError ? 'bg-rose-50 text-rose-700 border border-rose-200 dark:bg-rose-950 dark:text-rose-300' : 'bg-emerald-50 text-emerald-700 border border-emerald-200 dark:bg-emerald-950 dark:text-emerald-300'"
      >
        <span>{{ statusMessage }}</span>
        <button @click="statusMessage = ''" class="text-slate-400 hover:text-slate-600">✕</button>
      </div>

      <!-- Compare Button Bar -->
      <div class="flex flex-wrap items-center justify-between p-3 rounded-lg bg-slate-50 dark:bg-slate-800/60 border border-slate-200 dark:border-slate-700 text-xs gap-3">
        <span class="text-slate-600 dark:text-slate-300">
          Select any two revisions using the radio buttons to inspect Myers diff changes.
        </span>
        <button
          @click="toggleCompare"
          :disabled="comparing || revisions.length < 2"
          class="px-4 py-1.5 rounded-lg bg-forest-600 hover:bg-forest-500 text-white font-bold transition shadow-xs disabled:opacity-50"
        >
          {{ showInlineDiff ? 'Hide Diff Viewer' : 'Compare Selected Revisions' }}
        </button>
      </div>

      <!-- Inline Diff Viewer (if toggled) -->
      <div v-if="showInlineDiff" class="pt-2 pb-4 space-y-3">
        <div v-if="comparing" class="py-8 text-center text-slate-400 text-xs">
          <span class="inline-block animate-spin mr-1">⌛</span> Computing Myers diff between revisions...
        </div>
        <DiffViewer v-else :chunks="diffChunks" />
      </div>

      <!-- Loading State -->
      <div v-if="loading" class="py-12 text-center text-slate-400 text-xs">
        <div class="inline-block animate-spin text-2xl mb-1">⌛</div>
        <p>Loading revision ledger from Treewiki...</p>
      </div>

      <!-- Empty State -->
      <div v-else-if="revisions.length === 0" class="py-12 text-center text-slate-400 text-xs">
        <p>No revision history found for this page.</p>
      </div>

      <!-- Revision List Table -->
      <div v-else class="border border-slate-200 dark:border-slate-800 rounded-xl overflow-hidden divide-y divide-slate-100 dark:divide-slate-800">
        <div
          v-for="(rev, idx) in formattedRevisions"
          :key="rev.id"
          class="p-4 flex flex-wrap items-center justify-between gap-4 hover:bg-slate-50/80 dark:hover:bg-slate-800/40 transition text-xs"
        >
          <!-- Left: Selection Radios & Metadata -->
          <div class="flex items-center gap-3">
            <div class="flex items-center gap-1.5" title="Left: Base | Right: Target">
              <input type="radio" :value="rev.id" v-model="selectedBase" name="base" title="Base revision (from)" />
              <input type="radio" :value="rev.id" v-model="selectedTarget" name="target" title="Target revision (to)" />
            </div>

            <!-- Timestamp & Hash -->
            <div>
              <div class="flex items-center gap-2">
                <span class="font-bold text-slate-800 dark:text-slate-200">{{ rev.displayTime }}</span>
                <span class="font-mono text-[10px] text-slate-400">#{{ rev.id.substring(0, 8) }}</span>
                <span v-if="idx === 0" class="px-1.5 py-0.5 rounded text-[10px] font-bold bg-forest-100 dark:bg-forest-950 text-forest-700 dark:text-forest-400">
                  CURRENT
                </span>
                <span v-if="rev.is_minor" class="px-1.5 py-0.5 rounded text-[10px] font-mono bg-slate-100 dark:bg-slate-800 text-slate-500">
                  MINOR
                </span>
              </div>
              <p class="text-slate-500 text-[11px] mt-0.5">
                by <strong class="text-slate-700 dark:text-slate-300">{{ rev.author_id }}</strong>
                <span v-if="isMaskedIP(rev.author_id)" class="text-slate-400 font-mono"> (IP Contributor)</span>
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
              v-if="idx !== 0"
              @click="handleRollback(rev.id)"
              :disabled="rollingBack"
              class="px-2.5 py-1 rounded bg-slate-100 dark:bg-slate-800 hover:bg-rose-100 hover:text-rose-700 dark:hover:bg-rose-950 text-slate-600 dark:text-slate-300 text-[11px] font-semibold transition disabled:opacity-50"
              title="Revert back to this clean state"
            >
              ↺ Revert
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useApi } from '../../composables/useApi'

interface Revision {
  id: string
  page_id: string
  content: string
  summary: string
  author_id: string
  is_minor: boolean
  parent_revision_id?: string
  created_at: string
}

interface DiffChunk {
  op: number
  text: string
}

const route = useRoute()
const api = useApi()

const slug = computed(() => {
  const p = route.params.slug
  return Array.isArray(p) ? p.join('/') : p || 'Treewiki'
})

const displayTitle = computed(() => slug.value.replace(/_/g, ' '))

const loading = ref(true)
const revisions = ref<Revision[]>([])
const selectedBase = ref('')
const selectedTarget = ref('')
const showInlineDiff = ref(false)
const comparing = ref(false)
const diffChunks = ref<DiffChunk[]>([])
const rollingBack = ref(false)
const statusMessage = ref('')
const statusIsError = ref(false)

const isMaskedIP = (author: string) => {
  return author && (author.includes('xxx') || author.includes('.'))
}

const loadHistory = async () => {
  loading.value = true
  try {
    const list = await api.get<Revision[]>(`/pages/0/${encodeURIComponent(slug.value)}/history`)
    if (Array.isArray(list)) {
      // Sort newest to oldest
      revisions.value = list.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime())

      if (revisions.value.length > 0) {
        selectedTarget.value = revisions.value[0].id
        if (revisions.value.length > 1) {
          selectedBase.value = revisions.value[1].id
        } else {
          selectedBase.value = revisions.value[0].id
        }
      }
    }
  } catch (err) {
    console.error('Failed to load revision history:', err)
  } finally {
    loading.value = false
  }
}

const formattedRevisions = computed(() => {
  return revisions.value.map((rev, idx, arr) => {
    const nextRev = arr[idx + 1]
    const curLen = (rev.content || '').length
    const prevLen = nextRev ? (nextRev.content || '').length : 0
    const delta = curLen - prevLen

    let displayTime = rev.created_at
    try {
      displayTime = new Date(rev.created_at).toLocaleString()
    } catch {
      // fallback
    }

    return {
      ...rev,
      delta,
      displayTime,
    }
  })
})

const toggleCompare = async () => {
  showInlineDiff.value = !showInlineDiff.value
  if (showInlineDiff.value) {
    await fetchDiff()
  }
}

const fetchDiff = async () => {
  if (!selectedBase.value || !selectedTarget.value) return

  comparing.value = true
  try {
    const data = await api.get<DiffChunk[]>(
      `/pages/0/${encodeURIComponent(slug.value)}/diff`,
      { from: selectedBase.value, to: selectedTarget.value }
    )
    diffChunks.value = data || []
  } catch (err) {
    console.error('Failed to calculate diff:', err)
  } finally {
    comparing.value = false
  }
}

const handleRollback = async (revId: string) => {
  if (!confirm(`Are you sure you want to revert to revision #${revId.substring(0, 8)}? This will append a new reversion record.`)) {
    return
  }

  rollingBack.value = true
  try {
    await api.post(`/pages/0/${encodeURIComponent(slug.value)}/revert`, {
      revision_id: revId,
    })
    statusMessage.value = `Successfully reverted page to revision #${revId.substring(0, 8)}.`
    statusIsError.value = false
    await loadHistory()
  } catch (err: any) {
    statusMessage.value = `Revert failed: ${err.message || 'Permission denied or conflict'}`
    statusIsError.value = true
  } finally {
    rollingBack.value = false
  }
}

useSeoMeta({
  title: () => `Revision History: ${displayTitle.value} — Treewiki`,
  description: () => `Immutable revision ledger and Myers diff comparison for ${displayTitle.value}.`,
  ogTitle: () => `Revision History: ${displayTitle.value} — Treewiki`,
  ogType: 'article',
})

onMounted(() => {
  loadHistory()
})
</script>
