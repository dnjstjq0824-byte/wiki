<!-- apps/web/pages/special/recent-changes.vue -->
<template>
  <div>
    <h1 class="text-3xl font-bold text-slate-900 dark:text-slate-100 border-b-2 border-namu-600 dark:border-namu-500 pb-1">
      Recent Changes
    </h1>
    <p class="text-sm text-slate-500 mt-1">Edits, new pages, and editorial activity across Treewiki.</p>

    <div class="mt-3 flex flex-wrap items-center gap-3 text-sm border border-wiki-border dark:border-slate-700 bg-wiki-pagebg dark:bg-slate-800/40 px-3 py-2">
      <label class="flex items-center gap-2 flex-1 min-w-[12rem]">
        <span class="sr-only">Filter by title or author</span>
        <input
          v-model="searchFilter"
          type="text"
          placeholder="Filter by title or author…"
          class="flex-1 px-2 py-1 border border-wiki-border dark:border-slate-700 bg-white dark:bg-slate-950 text-slate-900 dark:text-slate-100 focus:outline-none focus:border-namu-600"
        />
      </label>
      <label class="flex items-center gap-1.5 text-slate-600 dark:text-slate-400 cursor-pointer">
        <input type="checkbox" v-model="hideMinor" />
        <span>Hide minor edits</span>
      </label>
      <button
        @click="loadChanges"
        class="px-3 py-1 border border-wiki-border dark:border-slate-700 bg-white dark:bg-slate-900 hover:bg-slate-100 dark:hover:bg-slate-800"
      >
        Refresh
      </button>
      <span class="text-slate-500 font-mono text-xs">{{ filteredChanges.length }} events</span>
    </div>

    <div v-if="loading" class="py-12 text-center text-sm text-slate-500">
      <p>Fetching recent changes…</p>
    </div>

    <div v-else-if="filteredChanges.length === 0" class="py-12 text-center">
      <p class="text-xl font-bold text-slate-800 dark:text-slate-200">No changes match your filter</p>
      <p class="text-sm text-slate-500 mt-1">New edits will appear here as they happen.</p>
    </div>

    <ul v-else class="mt-2 divide-y divide-slate-200 dark:divide-slate-800 text-sm">
      <li
        v-for="(item, idx) in filteredChanges"
        :key="idx"
        class="py-2 flex flex-wrap items-baseline gap-x-2.5 gap-y-1"
      >
        <span class="font-mono text-xs text-slate-500">{{ formatTimestamp(item.timestamp) }}</span>
        <span
          class="font-mono text-[11px] px-1 border uppercase"
          :class="getActionClass(item.action)"
        >
          {{ item.action }}
        </span>
        <NuxtLink
          :to="`/w/${encodeURIComponent(item.title.replace(/ /g, '_'))}`"
          class="font-semibold text-wiki-link dark:text-sky-400 hover:underline"
        >
          {{ item.title.replace(/_/g, ' ') }}
        </NuxtLink>
        <span class="text-[13px] text-slate-400">
          (<NuxtLink :to="`/history/${encodeURIComponent(item.title.replace(/ /g, '_'))}`" class="text-wiki-link dark:text-sky-400 hover:underline">hist</NuxtLink>
          ·
          <NuxtLink :to="`/talk/${encodeURIComponent(item.title.replace(/ /g, '_'))}`" class="text-wiki-link dark:text-sky-400 hover:underline">talk</NuxtLink>)
        </span>
        <span class="text-slate-500 italic text-[13px]">“{{ item.summary || 'Content updated' }}”</span>
        <span class="text-slate-600 dark:text-slate-300">{{ item.author }}</span>
        <span v-if="isMaskedIP(item.author)" class="font-mono text-[11px] text-slate-400">(unregistered)</span>
      </li>
    </ul>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useApi } from '../../composables/useApi'

interface ChangeItem {
  title: string
  namespace: number
  author: string
  summary: string
  action: string
  timestamp: string
}

const api = useApi()
const loading = ref(true)
const rawChanges = ref<ChangeItem[]>([])
const searchFilter = ref('')
const hideMinor = ref(false)

const isMaskedIP = (author: string) => {
  return author && (author.includes('xxx') || author.includes('.') || author.includes(':'))
}

const loadChanges = async () => {
  loading.value = true
  try {
    const list = await api.get<any[]>('/recentchanges')
    if (Array.isArray(list)) {
      rawChanges.value = list.map((item) => {
        if (typeof item === 'string') {
          // Check if it is a JSON string
          if (item.startsWith('{')) {
            try {
              const parsed = JSON.parse(item)
              return {
                title: parsed.title || 'Untitled',
                namespace: parsed.namespace || 0,
                author: parsed.author || 'anonymous',
                summary: parsed.summary || '',
                action: parsed.action || 'Edited',
                timestamp: parsed.timestamp || new Date().toISOString(),
              }
            } catch {
              // fallback text
            }
          }
          // Parse textual format: "Page Edited: Title by Author"
          const parts = item.split(' by ')
          const actionPart = parts[0] || 'Page Edited: Article'
          const author = parts[1] || 'anonymous'
          const title = actionPart.replace(/^(Page Edited: |Page Created: |)/i, '').trim()
          const action = actionPart.toLowerCase().includes('create') ? 'Created' : 'Edited'

          return {
            title: title || 'Treewiki',
            namespace: 0,
            author,
            summary: item,
            action,
            timestamp: new Date().toISOString(),
          }
        }
        return item
      })
    }
  } catch (err) {
    console.error('Failed to load recent changes:', err)
  } finally {
    loading.value = false
  }
}

const filteredChanges = computed(() => {
  let list = rawChanges.value
  const q = searchFilter.value.trim().toLowerCase()
  if (q) {
    list = list.filter(item =>
      item.title.toLowerCase().includes(q) ||
      item.author.toLowerCase().includes(q) ||
      item.summary.toLowerCase().includes(q)
    )
  }
  if (hideMinor.value) {
    list = list.filter(item => !item.summary.toLowerCase().includes('minor'))
  }
  return list
})

const getActionClass = (action: string) => {
  switch (action?.toUpperCase()) {
    case 'CREATED':
      return 'text-emerald-700 border-emerald-400 dark:text-emerald-400 dark:border-emerald-800'
    case 'EDITED':
      return 'text-sky-700 border-sky-400 dark:text-sky-400 dark:border-sky-800'
    case 'ROLLBACK':
      return 'text-rose-700 border-rose-400 dark:text-rose-400 dark:border-rose-800'
    default:
      return 'text-slate-600 border-slate-300'
  }
}

const formatTimestamp = (ts: string) => {
  if (!ts) return ''
  try {
    const d = new Date(ts)
    if (isNaN(d.getTime())) return ts
    return d.toLocaleTimeString()
  } catch {
    return ts
  }
}

useSeoMeta({
  title: 'Recent Changes — Treewiki',
  description: 'Real-time ledger of page edits, creations, and editorial activity across Treewiki.',
  ogTitle: 'Recent Changes — Treewiki',
  ogType: 'website',
})

onMounted(() => {
  loadChanges()
})
</script>
