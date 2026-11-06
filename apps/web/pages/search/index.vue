<!-- apps/web/pages/search/index.vue -->
<template>
  <div class="space-y-6 max-w-5xl mx-auto">
    <!-- Search Query Bar & Filters Card -->
    <div class="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl p-6 shadow-sm space-y-4">
      <div class="flex items-center justify-between border-b border-slate-100 dark:border-slate-800 pb-3">
        <h1 class="text-2xl font-bold text-slate-900 dark:text-slate-100 flex items-center gap-2">
          <span>🔎</span> Encyclopedic Search
        </h1>
        <span class="text-xs text-slate-400">Trigram & Faceted Full-Text Engine</span>
      </div>

      <!-- Main Search Input -->
      <form @submit.prevent="performSearch" class="flex gap-2">
        <div class="relative flex-1">
          <input
            v-model="query"
            type="text"
            placeholder="Search articles, algorithms, topics, or wikitext..."
            class="w-full pl-10 pr-4 py-2.5 text-sm rounded-lg bg-slate-50 dark:bg-slate-800 border border-slate-300 dark:border-slate-700 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-forest-500 transition"
          />
          <span class="absolute left-3.5 top-3 text-slate-400 text-sm">🔍</span>
          <button
            v-if="query"
            type="button"
            @click="query = ''; performSearch()"
            class="absolute right-3 top-2.5 text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 text-sm"
          >
            ✕
          </button>
        </div>

        <button
          type="submit"
          class="px-5 py-2.5 rounded-lg bg-forest-600 hover:bg-forest-500 text-white font-bold text-sm shadow-xs transition flex items-center gap-1.5"
        >
          <span>Search</span>
        </button>
      </form>

      <!-- Namespace Filter Chips -->
      <div class="flex flex-wrap items-center gap-2 text-xs pt-1">
        <span class="font-semibold text-slate-500 dark:text-slate-400">Namespace:</span>
        <button
          v-for="ns in namespaces"
          :key="ns.id"
          type="button"
          @click="selectNamespace(ns.id)"
          :class="selectedNs === ns.id ? 'bg-forest-600 text-white font-bold' : 'bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 hover:bg-slate-200 dark:hover:bg-slate-700'"
          class="px-2.5 py-1 rounded-md transition font-medium"
        >
          {{ ns.name }}
        </button>
      </div>
    </div>

    <!-- Search Results Status & Content -->
    <div class="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl p-6 shadow-sm min-h-[300px]">
      <!-- Loading State -->
      <div v-if="loading" class="py-16 text-center text-slate-400 space-y-2">
        <div class="inline-block animate-spin text-3xl">⌛</div>
        <p class="text-xs">Searching Treewiki knowledge index...</p>
      </div>

      <!-- Initial Prompt (No Query Entered) -->
      <div v-else-if="!query.trim()" class="py-16 text-center text-slate-400 space-y-3">
        <span class="text-4xl block">📚</span>
        <p class="text-sm font-semibold text-slate-600 dark:text-slate-300">Enter a keyword or title above to begin searching.</p>
        <p class="text-xs text-slate-400">Popular searches:
          <button @click="quickSearch('Treewiki')" class="text-forest-600 dark:text-forest-400 underline mx-1">Treewiki</button>,
          <button @click="quickSearch('Computer Science')" class="text-forest-600 dark:text-forest-400 underline mx-1">Computer Science</button>,
          <button @click="quickSearch('Algorithms')" class="text-forest-600 dark:text-forest-400 underline mx-1">Algorithms</button>,
          <button @click="quickSearch('Myers Diff')" class="text-forest-600 dark:text-forest-400 underline mx-1">Myers Diff</button>
        </p>
      </div>

      <!-- No Results Found State (with Red Link direct creation button) -->
      <div v-else-if="results.length === 0" class="py-12 px-6 text-center space-y-5">
        <div class="inline-flex p-4 rounded-full bg-amber-50 dark:bg-amber-950/40 text-amber-600 dark:text-amber-400 text-3xl">
          🔍
        </div>
        <div class="space-y-1">
          <h2 class="text-lg font-bold text-slate-800 dark:text-slate-200">
            No matching documents found for <span class="text-rose-600 dark:text-rose-400">"{{ query }}"</span>
          </h2>
          <p class="text-xs text-slate-500 max-w-md mx-auto">
            There is currently no page with this title or content in the selected namespace.
          </p>
        </div>

        <div class="pt-2">
          <NuxtLink
            :to="`/edit/${encodeURIComponent(query.trim().replace(/ /g, '_'))}`"
            class="inline-flex items-center gap-2 px-5 py-2.5 rounded-lg bg-forest-600 hover:bg-forest-500 text-white font-bold text-xs shadow-sm transition"
          >
            <span>✏️</span> Create "{{ query.trim() }}" on Treewiki
          </NuxtLink>
        </div>
      </div>

      <!-- Results List -->
      <div v-else class="space-y-4">
        <div class="flex items-center justify-between pb-3 border-b border-slate-100 dark:border-slate-800 text-xs text-slate-500">
          <span>Found <strong class="text-slate-800 dark:text-slate-200">{{ results.length }}</strong> matching documents</span>
          <span>Sorted by BM25 relevance & trigram match</span>
        </div>

        <div class="divide-y divide-slate-100 dark:divide-slate-800">
          <div
            v-for="(item, idx) in results"
            :key="idx"
            class="py-4 hover:bg-slate-50/70 dark:hover:bg-slate-800/40 px-3 rounded-lg transition space-y-1.5"
          >
            <!-- Title, Namespace Badge & Relevance Score -->
            <div class="flex items-center justify-between gap-3">
              <NuxtLink
                :to="`/w/${encodeURIComponent(item.document.title.replace(/ /g, '_'))}`"
                class="text-base font-bold text-forest-700 dark:text-forest-400 hover:underline flex items-center gap-2"
              >
                <span>📄</span>
                <span>{{ item.document.title.replace(/_/g, ' ') }}</span>
              </NuxtLink>

              <div class="flex items-center gap-2">
                <span class="text-[10px] font-mono px-2 py-0.5 rounded bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400">
                  {{ getNamespaceLabel(item.document.namespace) }}
                </span>
                <span class="text-[10px] font-mono px-1.5 py-0.5 rounded bg-forest-50 dark:bg-forest-950 text-forest-700 dark:text-forest-400 font-semibold">
                  Score {{ item.score.toFixed(1) }}
                </span>
              </div>
            </div>

            <!-- Snippet with Highlight -->
            <p
              class="text-xs text-slate-600 dark:text-slate-300 leading-relaxed font-sans line-clamp-3"
              v-html="formatSnippet(item.snippet || item.document.content)"
            ></p>

            <!-- Bottom metadata -->
            <div class="flex items-center gap-4 text-[11px] text-slate-400 pt-1">
              <NuxtLink
                :to="`/history/${encodeURIComponent(item.document.title.replace(/ /g, '_'))}`"
                class="hover:text-slate-600 dark:hover:text-slate-300 underline"
              >
                View History
              </NuxtLink>
              <NuxtLink
                :to="`/talk/${encodeURIComponent(item.document.title.replace(/ /g, '_'))}`"
                class="hover:text-slate-600 dark:hover:text-slate-300 underline"
              >
                Discussion Talk
              </NuxtLink>
              <NuxtLink
                :to="`/edit/${encodeURIComponent(item.document.title.replace(/ /g, '_'))}`"
                class="hover:text-slate-600 dark:hover:text-slate-300 underline"
              >
                Edit Page
              </NuxtLink>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useApi } from '../../composables/useApi'

interface SearchDoc {
  id: string
  namespace: number
  title: string
  content: string
}

interface SearchResultItem {
  document: SearchDoc
  score: number
  snippet: string
}

const route = useRoute()
const router = useRouter()
const api = useApi()

const query = ref('')
const selectedNs = ref<number>(-1)
const loading = ref(false)
const results = ref<SearchResultItem[]>([])

const namespaces = [
  { id: -1, name: 'All' },
  { id: 0, name: 'Main (0)' },
  { id: 1, name: 'Talk (1)' },
  { id: 2, name: 'User (2)' },
  { id: 4, name: 'Treewiki (4)' },
  { id: 14, name: 'Category (14)' },
]

const getNamespaceLabel = (ns: number) => {
  const match = namespaces.find(n => n.id === ns)
  return match ? match.name : `NS ${ns}`
}

const selectNamespace = (nsId: number) => {
  selectedNs.value = nsId
  performSearch()
}

const quickSearch = (term: string) => {
  query.value = term
  performSearch()
}

const performSearch = async () => {
  const q = query.value.trim()
  router.replace({
    path: '/search',
    query: {
      q: q || undefined,
      ns: selectedNs.value >= 0 ? selectedNs.value : undefined,
    },
  })

  if (!q) {
    results.value = []
    return
  }

  loading.value = true
  try {
    const params: Record<string, any> = { q }
    if (selectedNs.value >= 0) {
      params.namespace = selectedNs.value
    }
    const data = await api.get<SearchResultItem[]>('/search', params)
    results.value = data || []
  } catch (err) {
    console.error('Search failed:', err)
    results.value = []
  } finally {
    loading.value = false
  }
}

const formatSnippet = (raw: string) => {
  if (!raw) return ''
  const q = query.value.trim()
  let clean = raw.replace(/[<>{}]/g, '')
  if (q) {
    const escaped = q.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
    const regex = new RegExp(`(${escaped})`, 'gi')
    clean = clean.replace(regex, '<mark class="bg-amber-200 dark:bg-amber-900/60 dark:text-amber-200 px-0.5 rounded font-semibold">$1</mark>')
  }
  return clean
}

useSeoMeta({
  title: 'Search — Treewiki',
  description: 'Full-text faceted search across Treewiki articles, topics, and revisions.',
  ogTitle: 'Search — Treewiki',
  ogType: 'website',
})

onMounted(() => {
  if (route.query.q) {
    query.value = String(route.query.q)
  }
  if (route.query.ns !== undefined) {
    selectedNs.value = Number(route.query.ns)
  }
  if (query.value) {
    performSearch()
  }
})

watch(() => route.query.q, (newQ) => {
  if (newQ !== undefined && newQ !== query.value) {
    query.value = String(newQ)
    performSearch()
  }
})
</script>

