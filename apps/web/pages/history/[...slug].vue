<!-- apps/web/pages/history/[...slug].vue -->
<template>
  <div>
    <WikiTabs :slug="slug" active="history" />

    <h1 class="font-wiki text-3xl text-slate-900 dark:text-slate-100 border-b border-wiki-border dark:border-slate-700 pb-1">
      Revision History of “{{ displayTitle }}”
    </h1>
    <p class="text-sm text-slate-500 mt-1">Immutable revision ledger · select two revisions to compare</p>

    <!-- Notification -->
    <div
      v-if="statusMessage"
      class="mt-3 px-3 py-2 text-sm flex items-center justify-between border"
      :class="statusIsError ? 'bg-red-50 text-red-700 border-red-300 dark:bg-red-950 dark:text-red-300 dark:border-red-900' : 'bg-emerald-50 text-emerald-700 border-emerald-300 dark:bg-emerald-950 dark:text-emerald-300 dark:border-emerald-900'"
    >
      <span>{{ statusMessage }}</span>
      <button @click="statusMessage = ''" class="text-slate-400 hover:text-slate-600 px-1">✕</button>
    </div>

    <!-- Compare Bar -->
    <div class="mt-3 flex flex-wrap items-center justify-between gap-3 text-sm border border-wiki-border dark:border-slate-700 bg-wiki-pagebg dark:bg-slate-800/40 px-3 py-2">
      <span class="text-slate-600 dark:text-slate-300">
        Select two revisions with the radio buttons to compare them.
      </span>
      <button
        @click="toggleCompare"
        :disabled="comparing || revisions.length < 2"
        class="px-4 py-1 bg-wiki-link hover:bg-sky-700 text-white disabled:opacity-50"
      >
        {{ showInlineDiff ? 'Hide changes' : 'Compare selected revisions' }}
      </button>
    </div>

    <!-- Inline Diff Viewer (if toggled) -->
    <div v-if="showInlineDiff" class="mt-3">
      <div v-if="comparing" class="py-8 text-center text-slate-500 text-sm">
        Computing differences between revisions…
      </div>
      <DiffViewer v-else :chunks="diffChunks" />
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="py-12 text-center text-sm text-slate-500">
      <p>Loading revision history…</p>
    </div>

    <!-- Empty State -->
    <div v-else-if="revisions.length === 0" class="py-12 text-center text-sm text-slate-500">
      <p>No revision history found for this page.</p>
    </div>

    <!-- Revision List -->
    <ul v-else class="mt-2 divide-y divide-slate-200 dark:divide-slate-800 text-sm">
      <li
        v-for="(rev, idx) in formattedRevisions"
        :key="rev.id"
        class="py-2.5 flex flex-wrap items-center gap-x-3 gap-y-1"
      >
        <span class="flex items-center gap-1" title="Left: from · Right: to">
          <input type="radio" :value="rev.id" v-model="selectedBase" name="base" title="Compare from this revision" />
          <input type="radio" :value="rev.id" v-model="selectedTarget" name="target" title="Compare to this revision" />
        </span>
        <span class="text-slate-800 dark:text-slate-200">{{ rev.displayTime }}</span>
        <span class="font-mono text-xs text-slate-400">#{{ rev.id.substring(0, 8) }}</span>
        <span v-if="idx === 0" class="text-[11px] font-bold text-emerald-700 dark:text-emerald-400">(current)</span>
        <span v-if="rev.is_minor" class="text-[11px] font-mono text-slate-500">m</span>
        <span class="text-slate-600 dark:text-slate-300">
          <span class="font-semibold">{{ rev.author_id }}</span><span v-if="isMaskedIP(rev.author_id)" class="text-slate-400"> (unregistered)</span>
        </span>
        <span
          class="font-mono text-xs"
          :class="rev.delta > 0 ? 'text-emerald-700 dark:text-emerald-400' : rev.delta < 0 ? 'text-rose-700 dark:text-rose-400' : 'text-slate-400'"
        >
          ({{ rev.delta > 0 ? `+${rev.delta}` : rev.delta }})
        </span>
        <span class="text-slate-500 italic text-[13px]">“{{ rev.summary || 'No edit summary' }}”</span>
        <button
          v-if="idx !== 0"
          @click="handleRollback(rev.id)"
          :disabled="rollingBack"
          class="text-wiki-link dark:text-sky-400 hover:underline text-[13px] disabled:opacity-50"
          title="Revert to this revision"
        >
          undo
        </button>
      </li>
    </ul>
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
  if (!confirm(`Revert to revision #${revId.substring(0, 8)}? This will append a new reversion record.`)) {
    return
  }

  rollingBack.value = true
  try {
    await api.post(`/pages/0/${encodeURIComponent(slug.value)}/revert`, {
      revision_id: revId,
    })
    statusMessage.value = `Reverted to revision #${revId.substring(0, 8)}.`
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
