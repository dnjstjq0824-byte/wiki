<template>
  <div class="space-y-8">
    <!-- Hero Banner -->
    <section class="p-6 md:p-8 rounded-2xl bg-gradient-to-br from-forest-800 via-forest-900 to-slate-900 text-white shadow-md relative overflow-hidden">
      <div class="relative z-10 max-w-2xl">
        <div class="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-forest-700/60 text-forest-200 text-xs font-semibold mb-4 border border-forest-500/30">
          <span>🌲</span> Structured Knowledge & Topic Trees
        </div>
        <h1 class="text-3xl md:text-4xl font-extrabold tracking-tight mb-3">
          Welcome to Treewiki
        </h1>
        <p class="text-forest-100 text-sm md:text-base leading-relaxed mb-6">
          A global collaborative knowledge platform designed to organize human understanding into interconnected topic trees, immutable revision ledgers, and neutral encyclopedic references.
        </p>
        <div class="flex flex-wrap gap-3">
          <NuxtLink to="/w/Computer_Science" class="px-4 py-2 rounded-lg bg-forest-500 hover:bg-forest-600 font-semibold text-sm transition shadow-sm">
            Explore Computer Science
          </NuxtLink>
          <NuxtLink to="/tree" class="px-4 py-2 rounded-lg bg-white/10 hover:bg-white/20 font-semibold text-sm transition backdrop-blur border border-white/20">
            Open Knowledge Graph
          </NuxtLink>
        </div>
      </div>
    </section>

    <!-- Live Platform Stats -->
    <section aria-label="Platform statistics" class="grid grid-cols-2 md:grid-cols-4 gap-4">
      <div class="p-4 rounded-xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 shadow-sm text-center">
        <div v-if="statsLoading" class="h-7 flex items-center justify-center text-slate-300"><span class="inline-block animate-spin">⌛</span></div>
        <div v-else class="text-2xl font-extrabold text-forest-700 dark:text-forest-400 font-mono">{{ formatCount(stats?.pages) }}</div>
        <div class="text-[11px] text-slate-500 mt-0.5 font-semibold uppercase tracking-wide">Articles</div>
      </div>
      <div class="p-4 rounded-xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 shadow-sm text-center">
        <div v-if="statsLoading" class="h-7 flex items-center justify-center text-slate-300"><span class="inline-block animate-spin">⌛</span></div>
        <div v-else class="text-2xl font-extrabold text-forest-700 dark:text-forest-400 font-mono">{{ formatCount(stats?.topics) }}</div>
        <div class="text-[11px] text-slate-500 mt-0.5 font-semibold uppercase tracking-wide">Topic Nodes</div>
      </div>
      <div class="p-4 rounded-xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 shadow-sm text-center">
        <div class="text-2xl font-extrabold text-forest-700 dark:text-forest-400">4.0</div>
        <div class="text-[11px] text-slate-500 mt-0.5 font-semibold uppercase tracking-wide">CC BY-SA License</div>
      </div>
      <NuxtLink to="/special/random" class="p-4 rounded-xl bg-forest-600 hover:bg-forest-500 text-white shadow-sm text-center transition block">
        <div class="text-2xl font-extrabold">🎲</div>
        <div class="text-[11px] mt-0.5 font-semibold uppercase tracking-wide">Random Article</div>
      </NuxtLink>
    </section>

    <!-- Main Grid: Featured Article & Topic Trees -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
      <!-- Featured Article Card -->
      <div class="md:col-span-2 p-6 rounded-xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 shadow-sm space-y-4">
        <div class="flex items-center justify-between border-b border-slate-100 dark:border-slate-800 pb-3">
          <span class="text-xs font-bold text-forest-700 dark:text-forest-400 uppercase tracking-wide">⭐ Today's Featured Knowledge</span>
          <span class="text-xs text-slate-400">CC BY-SA 4.0</span>
        </div>
        <h2 class="text-2xl font-bold text-slate-900 dark:text-slate-100">
          <NuxtLink to="/w/Computer_Science" class="hover:text-forest-600 transition">
            Computer Science & Algorithmic Foundations
          </NuxtLink>
        </h2>
        <p class="text-sm text-slate-600 dark:text-slate-400 leading-relaxed">
          Computer Science is the systematic study of computation, information processing, and algorithmic structures. From Myers diff sequence comparison to Directed Acyclic Graph (DAG) taxonomies, explore how modern knowledge representations are built...
        </p>
        <div class="flex items-center justify-between pt-2">
          <div class="flex gap-2">
            <span class="px-2 py-0.5 rounded text-xs bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400">Theory</span>
            <span class="px-2 py-0.5 rounded text-xs bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400">Algorithms</span>
            <span class="px-2 py-0.5 rounded text-xs bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400">Taxonomy</span>
          </div>
          <NuxtLink to="/w/Computer_Science" class="text-xs font-semibold text-forest-600 dark:text-forest-400 hover:underline">
            Read full article →
          </NuxtLink>
        </div>
      </div>

      <!-- Foundational Articles -->
      <div class="p-6 rounded-xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 shadow-sm space-y-4">
        <h3 class="font-bold text-base text-slate-900 dark:text-slate-100 border-b border-slate-100 dark:border-slate-800 pb-3 flex items-center gap-2">
          <span>🌿</span> Foundational Articles
        </h3>
        <ul class="space-y-2.5 text-sm">
          <li v-for="art in foundationArticles" :key="art.slug" class="flex items-center justify-between p-2 rounded-lg hover:bg-slate-50 dark:hover:bg-slate-800/60 transition">
            <NuxtLink :to="`/w/${art.slug}`" class="font-medium text-slate-700 dark:text-slate-300 hover:text-forest-600 flex items-center gap-2">
              <span>{{ art.icon }}</span> {{ art.name }}
            </NuxtLink>
            <span class="text-slate-300 dark:text-slate-600">→</span>
          </li>
        </ul>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useApi } from '../composables/useApi'

interface PlatformStats {
  version: string
  database: string
  pages: number
  topics: number
  timestamp: string
}

const foundationArticles = [
  { name: 'Computer Science', slug: 'Computer_Science', icon: '💻' },
  { name: 'Algorithms', slug: 'Algorithms', icon: '⚙️' },
  { name: 'Data Structures', slug: 'Data_Structures', icon: '🧱' },
  { name: 'Graph Theory', slug: 'Graph_Theory', icon: '🕸️' },
  { name: 'Myers Diff', slug: 'Myers_Diff', icon: '⚖️' },
]

const api = useApi()
const stats = ref<PlatformStats | null>(null)
const statsLoading = ref(true)

const formatCount = (n: number | undefined) => {
  if (n === undefined || n === null) return '—'
  return n.toLocaleString('en-US')
}

useSeoMeta({
  title: 'Treewiki — Collaborative Knowledge Platform',
  description: 'A global collaborative knowledge platform organizing human understanding into topic trees, revision ledgers, and neutral encyclopedic references. Free under CC BY-SA 4.0.',
  ogTitle: 'Treewiki — Collaborative Knowledge Platform',
  ogDescription: 'Structured knowledge, topic trees, and immutable revision ledgers. Free under CC BY-SA 4.0.',
  ogType: 'website',
})

onMounted(async () => {
  try {
    stats.value = await api.get<PlatformStats>('/stats')
  } catch (err) {
    console.error('Failed to load platform stats:', err)
  } finally {
    statsLoading.value = false
  }
})
</script>
