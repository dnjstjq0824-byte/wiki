<template>
  <div>
    <h1 class="text-3xl font-bold text-slate-900 dark:text-slate-100 border-b-2 border-namu-600 dark:border-namu-500 pb-1 text-center">
      Welcome to Treewiki
    </h1>
    <p class="text-center text-sm text-slate-600 dark:text-slate-400 mt-2">
      The free knowledge platform that anyone can edit ·
      <span class="whitespace-nowrap">{{ formatCount(stats?.pages) }} articles</span> ·
      <span class="whitespace-nowrap">{{ formatCount(stats?.topics) }} topic nodes</span>
    </p>

    <div class="grid grid-cols-1 md:grid-cols-2 gap-4 mt-5">
      <!-- Featured article -->
      <section aria-label="Featured article" class="border border-wiki-border dark:border-slate-700 bg-wiki-pagebg dark:bg-slate-800/30">
        <h2 class="px-3 py-1.5 bg-slate-100 dark:bg-slate-800 border-b border-wiki-border dark:border-slate-700 font-bold text-[15px]">
          Featured article
        </h2>
        <div class="p-4">
          <p class="text-xl font-bold">
            <NuxtLink to="/w/Computer_Science" class="text-wiki-link dark:text-sky-400 hover:underline">
              Computer Science
            </NuxtLink>
          </p>
          <p class="text-sm text-slate-700 dark:text-slate-300 leading-relaxed mt-2">
            Computer science is the systematic study of computation, information processing, and algorithmic
            structures. From Myers diff sequence comparison to directed acyclic graph taxonomies, explore how
            modern knowledge representations are built.
          </p>
          <p class="text-sm mt-2">
            <NuxtLink to="/w/Computer_Science" class="text-wiki-link dark:text-sky-400 hover:underline">
              Read the full article…
            </NuxtLink>
          </p>
        </div>
      </section>

      <!-- Statistics -->
      <section aria-label="Platform statistics" class="border border-wiki-border dark:border-slate-700 bg-wiki-pagebg dark:bg-slate-800/30">
        <h2 class="px-3 py-1.5 bg-slate-100 dark:bg-slate-800 border-b border-wiki-border dark:border-slate-700 font-bold text-[15px]">
          Treewiki at a glance
        </h2>
        <div class="p-4 text-sm">
          <div v-if="statsLoading" class="text-slate-500">Loading statistics…</div>
          <ul v-else class="space-y-1.5 text-slate-700 dark:text-slate-300">
            <li><strong class="font-mono">{{ formatCount(stats?.pages) }}</strong> Articles</li>
            <li><strong class="font-mono">{{ formatCount(stats?.topics) }}</strong> Topic Nodes</li>
            <li>Content license: <strong>CC BY-SA 4.0</strong></li>
          </ul>
          <ul class="mt-3 pt-3 border-t border-wiki-border dark:border-slate-700 space-y-1">
            <li><NuxtLink to="/special/random" class="text-wiki-link dark:text-sky-400 hover:underline">Random article</NuxtLink></li>
            <li><NuxtLink to="/tree" class="text-wiki-link dark:text-sky-400 hover:underline">Browse the topic tree</NuxtLink></li>
            <li><NuxtLink to="/special/recent-changes" class="text-wiki-link dark:text-sky-400 hover:underline">Recent changes</NuxtLink></li>
          </ul>
        </div>
      </section>
    </div>

    <!-- Foundations -->
    <section aria-label="Foundational topics" class="mt-4 border border-wiki-border dark:border-slate-700">
      <h2 class="px-3 py-1.5 bg-slate-100 dark:bg-slate-800 border-b border-wiki-border dark:border-slate-700 font-bold text-[15px]">
        Foundational topics
      </h2>
      <ul class="grid grid-cols-2 md:grid-cols-5 gap-px bg-slate-200 dark:bg-slate-800 text-sm">
        <li v-for="art in foundationArticles" :key="art.slug" class="bg-white dark:bg-slate-900 px-3 py-2">
          <NuxtLink :to="`/w/${art.slug}`" class="text-wiki-link dark:text-sky-400 hover:underline">
            {{ art.name }}
          </NuxtLink>
        </li>
      </ul>
    </section>
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
  { name: 'Computer Science', slug: 'Computer_Science' },
  { name: 'Algorithms', slug: 'Algorithms' },
  { name: 'Data Structures', slug: 'Data_Structures' },
  { name: 'Graph Theory', slug: 'Graph_Theory' },
  { name: 'Myers Diff', slug: 'Myers_Diff' },
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
