<!-- apps/web/pages/special/recent-changes.vue -->
<template>
  <div class="space-y-6 max-w-5xl mx-auto">
    <!-- Header Card -->
    <div class="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl p-6 shadow-sm space-y-4">
      <div class="flex flex-wrap items-center justify-between gap-4 border-b border-slate-100 dark:border-slate-800 pb-4">
        <div>
          <h1 class="text-2xl font-bold text-slate-900 dark:text-slate-100 flex items-center gap-2">
            <span>⚡</span> Recent Changes
          </h1>
          <p class="text-xs text-slate-500 mt-1">
            Real-time ledger of page edits, new article creations, and editorial governance across Treewiki.
          </p>
        </div>

        <div class="flex items-center gap-2">
          <button
            @click="loadChanges"
            class="px-3.5 py-1.5 rounded-lg border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 hover:bg-slate-100 dark:hover:bg-slate-700 text-slate-700 dark:text-slate-300 text-xs font-semibold flex items-center gap-1.5 transition"
          >
            <span :class="{ 'animate-spin': loading }">🔄</span> Refresh Feed
          </button>
        </div>
      </div>

      <!-- Filters & Controls Bar -->
      <div class="flex flex-wrap items-center justify-between gap-3 text-xs pt-1">
        <div class="flex items-center gap-2 flex-1 max-w-sm">
          <input
            v-model="searchFilter"
            type="text"
            placeholder="Filter by title or author..."
            class="w-full px-3 py-1.5 rounded-lg bg-slate-50 dark:bg-slate-800 border border-slate-200 dark:border-slate-700 text-slate-800 dark:text-slate-200 focus:outline-none focus:ring-1 focus:ring-forest-500 text-xs"
          />
        </div>

        <div class="flex items-center gap-3">
          <label class="flex items-center gap-1.5 text-slate-600 dark:text-slate-400 cursor-pointer">
            <input type="checkbox" v-model="hideMinor" class="rounded text-forest-600" />
            <span>Hide minor edits</span>
          </label>
          <span class="text-slate-400 font-mono">Total events: {{ filteredChanges.length }}</span>
        </div>
      </div>
    </div>

    <!-- Changes Feed -->
    <div class="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl overflow-hidden shadow-sm">
      <div v-if="loading" class="py-16 text-center text-slate-400 space-y-2">
        <div class="inline-block animate-spin text-3xl">⌛</div>
        <p class="text-xs">Fetching recent changes from ledger...</p>
      </div>

      <div v-else-if="filteredChanges.length === 0" class="py-16 text-center text-slate-400 space-y-2">
        <span class="text-3xl block">📋</span>
        <p class="text-sm font-semibold text-slate-600 dark:text-slate-300">No recent changes match your filter.</p>
        <p class="text-xs text-slate-400">All updates and edits will be displayed here in real time.</p>
      </div>

      <div v-else class="divide-y divide-slate-100 dark:divide-slate-800">
        <div
          v-for="(item, idx) in filteredChanges"
          :key="idx"
          class="p-4 hover:bg-slate-50/70 dark:hover:bg-slate-800/40 transition flex flex-wrap items-center justify-between gap-4 text-xs"
        >
          <!-- Left: Action Tag, Title, and Quick Links -->
          <div class="flex items-center gap-3 min-w-[300px]">
            <span
              class="px-2 py-0.5 rounded font-mono text-[10px] font-bold uppercase border"
              :class="getActionClass(item.action)"
            >
              {{ item.action }}
            </span>

            <div class="space-y-0.5">
              <div class="flex items-center gap-2">
                <NuxtLink
                  :to="`/w/${encodeURIComponent(item.title.replace(/ /g, '_'))}`"
                  class="font-bold text-slate-900 dark:text-slate-100 hover:text-forest-600 dark:hover:text-forest-400 transition"
                >
                  {{ item.title.replace(/_/g, ' ') }}
                </NuxtLink>

                <div class="flex items-center gap-1 text-[10px] text-slate-400">
                  <span>(</span>
                  <NuxtLink :to="`/history/${encodeURIComponent(item.title.replace(/ /g, '_'))}`" class="hover:underline">hist</NuxtLink>
                  <span>|</span>
                  <NuxtLink :to="`/talk/${encodeURIComponent(item.title.replace(/ /g, '_'))}`" class="hover:underline">talk</NuxtLink>
                  <span>)</span>
                </div>
              </div>

              <!-- Summary Note -->
              <p class="text-[11px] text-slate-500 italic">
                "{{ item.summary || 'Content updated' }}"
              </p>
            </div>
          </div>

          <!-- Right: Author & Time -->
          <div class="flex items-center gap-4 text-right">
            <div>
              <div class="flex items-center gap-1.5 justify-end">
                <span class="font-semibold text-slate-700 dark:text-slate-300">
                  {{ item.author }}
                </span>
                <span v-if="isMaskedIP(item.author)" class="px-1.5 py-0.2 rounded bg-slate-100 dark:bg-slate-800 text-slate-400 font-mono text-[9px]">
                  IP
                </span>
              </div>
              <span class="text-[10px] text-slate-400 font-mono">
                {{ formatTimestamp(item.timestamp) }}
              </span>
            </div>

            <NuxtLink
              :to="`/w/${encodeURIComponent(item.title.replace(/ /g, '_'))}`"
              class="px-2.5 py-1 rounded bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700 text-slate-600 dark:text-slate-300 text-[11px] font-semibold transition"
            >
              View →
            </NuxtLink>
          </div>
        </div>
      </div>
    </div>
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
      return 'bg-emerald-50 text-emerald-700 border-emerald-300 dark:bg-emerald-950 dark:text-emerald-400 dark:border-emerald-800'
    case 'EDITED':
      return 'bg-sky-50 text-sky-700 border-sky-300 dark:bg-sky-950 dark:text-sky-400 dark:border-sky-800'
    case 'ROLLBACK':
      return 'bg-rose-50 text-rose-700 border-rose-300 dark:bg-rose-950 dark:text-rose-400 dark:border-rose-800'
    default:
      return 'bg-slate-100 text-slate-700 border-slate-200'
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

onMounted(() => {
  loadChanges()
})
</script>
