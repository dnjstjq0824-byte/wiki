<!-- apps/web/pages/special/random.vue -->
<template>
  <div class="max-w-2xl mx-auto text-center">
    <h1 class="text-3xl font-bold text-slate-900 dark:text-slate-100 border-b-2 border-namu-600 dark:border-namu-500 pb-1">
      Random Article
    </h1>

    <div v-if="loading" class="py-10 text-sm text-slate-500">
      <p>Choosing a random article…</p>
    </div>

    <div v-else-if="loadError" class="py-8 space-y-4">
      <p class="text-sm text-slate-500">{{ loadError }}</p>
      <button
        @click="rollRandom"
        class="px-5 py-1.5 text-sm bg-namu-600 hover:bg-namu-700 text-white"
      >
        Try again
      </button>
    </div>

    <div v-else-if="page" class="py-8 space-y-3">
      <p class="text-xs text-slate-500 uppercase tracking-wide">Your random destination</p>
      <p class="text-2xl font-bold">
        <NuxtLink :to="`/w/${page.title}`" class="text-wiki-link dark:text-sky-400 hover:underline">
          {{ page.title.replace(/_/g, ' ') }}
        </NuxtLink>
      </p>
      <p v-if="page.summary" class="text-sm text-slate-500 max-w-md mx-auto leading-relaxed">
        {{ page.summary }}
      </p>
      <div class="flex items-center justify-center gap-2 pt-2 text-sm">
        <NuxtLink
          :to="`/w/${page.title}`"
          class="px-5 py-1.5 bg-namu-600 hover:bg-namu-700 text-white"
        >
          Read Article
        </NuxtLink>
        <button
          @click="rollRandom"
          class="px-5 py-1.5 border border-wiki-border dark:border-slate-700 hover:bg-slate-100 dark:hover:bg-slate-800"
        >
          Roll again
        </button>
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
