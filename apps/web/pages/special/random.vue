<!-- apps/web/pages/special/random.vue -->
<template>
  <div class="max-w-3xl mx-auto">
    <div class="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl p-8 shadow-sm text-center space-y-5">
      <div class="text-4xl">🎲</div>
      <h1 class="text-2xl font-bold text-slate-900 dark:text-slate-100">Random Article</h1>

      <div v-if="loading" class="py-8 text-slate-400 text-xs space-y-2">
        <div class="inline-block animate-spin text-2xl">⌛</div>
        <p>Rolling the knowledge dice...</p>
      </div>

      <div v-else-if="loadError" class="space-y-4">
        <p class="text-xs text-slate-500">{{ loadError }}</p>
        <button
          @click="rollRandom"
          class="px-5 py-2 rounded-lg bg-forest-600 hover:bg-forest-500 text-white font-bold text-xs transition"
        >
          Try Again
        </button>
      </div>

      <div v-else-if="page" class="space-y-4">
        <div>
          <p class="text-[11px] font-bold text-forest-700 dark:text-forest-400 uppercase tracking-wider mb-1">Your random destination</p>
          <NuxtLink :to="`/w/${page.title}`" class="text-xl font-bold text-slate-900 dark:text-slate-100 hover:text-forest-600 dark:hover:text-forest-400 transition">
            {{ page.title.replace(/_/g, ' ') }}
          </NuxtLink>
          <p v-if="page.summary" class="text-xs text-slate-500 mt-2 max-w-md mx-auto leading-relaxed">
            {{ page.summary }}
          </p>
        </div>
        <div class="flex items-center justify-center gap-2">
          <NuxtLink
            :to="`/w/${page.title}`"
            class="px-5 py-2 rounded-lg bg-forest-600 hover:bg-forest-500 text-white font-bold text-xs transition"
          >
            Read Article →
          </NuxtLink>
          <button
            @click="rollRandom"
            class="px-5 py-2 rounded-lg border border-slate-200 dark:border-slate-700 text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-800 font-bold text-xs transition"
          >
            🎲 Roll Again
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useApi } from '../../composables/useApi'

interface RandomPage {
  id: string
  namespace: number
  title: string
  summary: string
}

const api = useApi()
const loading = ref(true)
const loadError = ref('')
const page = ref<RandomPage | null>(null)

const rollRandom = async () => {
  loading.value = true
  loadError.value = ''
  page.value = null
  try {
    page.value = await api.get<RandomPage>('/pages/random')
  } catch (err) {
    console.error('Failed to roll random article:', err)
    loadError.value = 'The article service is unreachable. Check your connection and try again.'
  } finally {
    loading.value = false
  }
}

useSeoMeta({
  title: 'Random Article — Treewiki',
  description: 'Discover a random article from the Treewiki knowledge base.',
  ogTitle: 'Random Article — Treewiki',
  ogType: 'website',
})

onMounted(() => {
  rollRandom()
})
</script>
