<!-- apps/web/pages/search/index.vue -->
<template>
  <div>
    <h1 class="text-3xl font-bold text-slate-900 dark:text-slate-100 border-b-2 border-namu-600 dark:border-namu-500 pb-1">
      Search
    </h1>

    <form @submit.prevent="performSearch" class="mt-3 flex gap-0 max-w-2xl" role="search">
      <label for="search-input" class="sr-only">Search query</label>
      <input
        id="search-input"
        v-model="query"
        type="text"
        placeholder="Search articles, algorithms, topics, or wikitext..."
        class="flex-1 px-2.5 py-1.5 text-sm border border-wiki-border dark:border-slate-700 bg-white dark:bg-slate-950 text-slate-900 dark:text-slate-100 focus:outline-none focus:border-namu-600"
      />
      <button
        type="submit"
        class="px-4 text-sm border border-l-0 border-wiki-border dark:border-slate-700 bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700"
      >
        Search
      </button>
    </form>

    <div class="mt-2 flex flex-wrap items-center gap-x-3 gap-y-1 text-[13px]">
      <span class="text-slate-500">Namespace:</span>
      <button
        v-for="ns in namespaces"
        :key="ns.id"
        type="button"
        @click="selectNamespace(ns.id)"
        :class="selectedNs === ns.id ? 'font-bold text-slate-900 dark:text-slate-100 underline' : 'text-wiki-link dark:text-sky-400 hover:underline'"
      >
        {{ ns.name }}
      </button>
    </div>

    <hr class="my-4 border-wiki-border dark:border-slate-700" />

    <!-- Loading State -->
    <div v-if="loading" class="py-10 text-center text-sm text-slate-500">
      <p>Searching the knowledge index…</p>
    </div>

    <!-- Initial Prompt -->
    <div v-else-if="!query.trim()" class="text-sm text-slate-600 dark:text-slate-300">
      <p>Enter a keyword or title above to begin searching.</p>
      <p class="mt-2 text-slate-500">
        Popular:
        <button @click="quickSearch('Treewiki')" class="text-wiki-link dark:text-sky-400 hover:underline">Treewiki</button> ·
        <button @click="quickSearch('Computer Science')" class="text-wiki-link dark:text-sky-400 hover:underline">Computer Science</button> ·
        <button @click="quickSearch('Algorithms')" class="text-wiki-link dark:text-sky-400 hover:underline">Algorithms</button> ·
        <button @click="quickSearch('Myers Diff')" class="text-wiki-link dark:text-sky-400 hover:underline">Myers Diff</button>
      </p>
    </div>

    <!-- No Results -->
    <div v-else-if="results.length === 0" class="text-sm">
      <h2 class="text-xl font-bold text-slate-900 dark:text-slate-100">No results for “{{ query }}”</h2>
      <p class="text-slate-600 dark:text-slate-400 mt-1">There is no page with this title or content in the selected namespace.</p>
      <p class="mt-3">
        <NuxtLink
          :to="`/edit/${encodeURIComponent(query.trim().replace(/ /g, '_'))}`"
          class="text-wiki-redlink hover:underline font-semibold"
        >
          Create the page “{{ query.trim() }}”
        </NuxtLink>
      </p>
    </div>

    <!-- Results List -->
    <div v-else>
      <p class="text-sm text-slate-500">{{ results.length }} result(s) found, sorted by relevance.</p>
      <ul class="mt-2 space-y-4">
        <li v-for="(item, idx) in results" :key="idx">
          <p class="text-lg">
            <NuxtLink
              :to="`/w/${encodeURIComponent(item.document.title.replace(/ /g, '_'))}`"
              class="text-wiki-link dark:text-sky-400 hover:underline"
            >
              {{ item.document.title.replace(/_/g, ' ') }}
            </NuxtLink>
            <span class="ml-2 text-xs font-mono text-slate-400">{{ getNamespaceLabel(item.document.namespace) }}</span>
          </p>
          <p
            class="text-sm text-slate-600 dark:text-slate-300 leading-relaxed"
            v-html="formatSnippet(item.snippet || item.document.content)"
          ></p>
          <p class="text-[13px] text-slate-500">
            <NuxtLink :to="`/history/${encodeURIComponent(item.document.title.replace(/ /g, '_'))}`" class="text-wiki-link dark:text-sky-400 hover:underline">history</NuxtLink>
            ·
            <NuxtLink :to="`/talk/${encodeURIComponent(item.document.title.replace(/ /g, '_'))}`" class="text-wiki-link dark:text-sky-400 hover:underline">talk</NuxtLink>
            ·
            <NuxtLink :to="`/edit/${encodeURIComponent(item.document.title.replace(/ /g, '_'))}`" class="text-wiki-link dark:text-sky-400 hover:underline">edit</NuxtLink>
          </p>
        </li>
      </ul>
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
    clean = clean.replace(regex, '<mark class="bg-amber-200 dark:bg-amber-900/60 dark:text-amber-200 px-0.5 font-semibold">$1</mark>')
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
