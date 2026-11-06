<!-- apps/web/pages/w/[...slug].vue -->
<template>
  <div class="space-y-6">
    <!-- Loading State -->
    <div v-if="pending" class="p-12 text-center text-slate-400">
      <div class="inline-block animate-spin text-2xl mb-2">⌛</div>
      <p class="text-xs">Loading Treewiki document...</p>
    </div>

    <!-- Error / Red Link Page Does Not Exist State -->
    <div
      v-else-if="pageNotFound"
      class="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl p-8 shadow-sm space-y-6"
    >
      <div class="border-b border-slate-200 dark:border-slate-800 pb-4">
        <h1 class="text-3xl font-extrabold text-slate-900 dark:text-slate-100 tracking-tight">
          {{ displayTitle }}
        </h1>
        <p class="text-xs text-rose-600 dark:text-rose-400 mt-1 font-semibold flex items-center gap-1">
          <span>⚠️</span> Document does not exist in the Treewiki database.
        </p>
      </div>

      <div class="p-6 rounded-xl bg-slate-50 dark:bg-slate-800/40 border border-slate-200 dark:border-slate-800 space-y-4 text-xs text-slate-600 dark:text-slate-400 leading-relaxed">
        <p>
          There is currently no article with the title <strong class="text-slate-800 dark:text-slate-200">"{{ displayTitle }}"</strong>.
        </p>
        <ul class="list-disc list-inside space-y-1">
          <li>You can <NuxtLink :to="`/search?q=${encodeURIComponent(displayTitle)}`" class="text-sky-600 dark:text-sky-400 hover:underline">search for this title</NuxtLink> across other encyclopedic articles.</li>
          <li>You can check related topic trees to find where this subject belongs.</li>
          <li>Or, as a collaborative contributor, you can create and publish the first revision!</li>
        </ul>
        <div class="pt-2">
          <NuxtLink
            :to="`/edit/${slug}`"
            class="inline-flex items-center gap-2 px-5 py-2.5 rounded-lg bg-forest-600 hover:bg-forest-500 text-white font-bold text-xs shadow-sm transition"
          >
            <span>✏️</span> Create this Article Now
          </NuxtLink>
        </div>
      </div>
    </div>

    <!-- Service Error State -->
    <div
      v-else-if="loadError"
      class="bg-white dark:bg-slate-900 border border-rose-200 dark:border-rose-900 rounded-xl p-8 shadow-sm space-y-4 text-center"
    >
      <div class="text-3xl">📡</div>
      <h1 class="text-lg font-bold text-slate-800 dark:text-slate-200">Service temporarily unavailable</h1>
      <p class="text-xs text-slate-500">{{ loadError }}</p>
      <button
        @click="fetchArticle"
        class="px-5 py-2 rounded-lg bg-forest-600 hover:bg-forest-500 text-white font-bold text-xs transition"
      >
        Retry
      </button>
    </div>

    <!-- Rendered Article View -->
    <article
      v-else
      class="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl p-6 md:p-8 shadow-sm space-y-6"
    >
      <!-- Article Header & Action Tab Bar -->
      <div class="border-b border-slate-200 dark:border-slate-800 pb-4">
        <div class="flex flex-wrap items-center justify-between gap-4 mb-2">
          <h1 class="text-3xl font-extrabold text-slate-900 dark:text-slate-100 tracking-tight">
            {{ displayTitle }}
          </h1>

          <!-- Tab Actions -->
          <div class="flex items-center gap-1 bg-slate-100 dark:bg-slate-800 p-1 rounded-lg text-xs font-semibold">
            <NuxtLink :to="`/w/${slug}`" class="px-3 py-1.5 rounded-md bg-white dark:bg-slate-700 text-forest-700 dark:text-forest-400 shadow-xs">
              Article
            </NuxtLink>
            <NuxtLink :to="`/talk/${slug}`" class="px-3 py-1.5 rounded-md text-slate-600 dark:text-slate-300 hover:text-slate-900">
              Talk
            </NuxtLink>
            <NuxtLink :to="`/edit/${slug}`" class="px-3 py-1.5 rounded-md text-slate-600 dark:text-slate-300 hover:text-slate-900">
              ✏️ Edit
            </NuxtLink>
            <NuxtLink :to="`/history/${slug}`" class="px-3 py-1.5 rounded-md text-slate-600 dark:text-slate-300 hover:text-slate-900">
              🕒 History
            </NuxtLink>
          </div>
        </div>

        <div class="text-xs text-slate-500 flex items-center gap-2">
          <span>From Treewiki, the free knowledge platform</span>
          <span>•</span>
          <span>Status: Verified CC BY-SA 4.0</span>
        </div>
      </div>

      <!-- Rendered HTML Content from Backend Parser -->
      <div
        class="prose dark:prose-invert max-w-none text-slate-800 dark:text-slate-200 leading-relaxed text-sm"
        v-html="renderedBody"
      ></div>

      <!-- Category Footer -->
      <div v-if="categories.length > 0" class="mt-8 pt-4 border-t border-slate-100 dark:border-slate-800 flex flex-wrap items-center gap-2 text-xs">
        <span class="font-bold text-slate-400">Categories:</span>
        <NuxtLink
          v-for="cat in categories"
          :key="cat"
          :to="`/search?q=${encodeURIComponent(cat)}`"
          class="px-2.5 py-1 rounded-full bg-slate-100 dark:bg-slate-800 text-forest-700 dark:text-forest-400 font-medium hover:bg-forest-50"
        >
          📁 {{ cat }}
        </NuxtLink>
      </div>
    </article>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useApi } from '../../composables/useApi'

const route = useRoute()
const api = useApi()

const slug = computed(() => {
  const p = route.params.slug
  return Array.isArray(p) ? p.join('/') : p || 'Main_Page'
})

const displayTitle = computed(() => slug.value.replace(/_/g, ' '))

const pending = ref(true)
const pageNotFound = ref(false)
const loadError = ref('')
const articleData = ref<any>(null)
const renderedBody = ref('')
const categories = ref<string[]>([])

const fetchArticle = async () => {
  pending.value = true
  pageNotFound.value = false
  loadError.value = ''

  try {
    const res: any = await api.get(`/pages/0/${slug.value}`)
    const page = res?.page || (res?.title ? res : null)
    if (page) {
      articleData.value = page
      categories.value = res.categories || []
      
      // Parse content via preview API
      if (page.content) {
        const preview: any = await api.post('/parser/preview', { wikitext: page.content })
        renderedBody.value = preview?.html || `<p>${page.content}</p>`
        if (preview?.categories && preview.categories.length > 0) {
          categories.value = preview.categories
        }
      } else {
        renderedBody.value = `<p><strong>${displayTitle.value}</strong> is an encyclopedic topic in Treewiki.</p>`
      }
    } else {
      pageNotFound.value = true
    }
  } catch (err: any) {
    if (err.status === 404 || err.statusCode === 404 || err.response?.status === 404) {
      pageNotFound.value = true
    } else {
      loadError.value = 'The article service is unreachable. Check your connection and try again.'
      console.error('Failed to load article:', err)
    }
  } finally {
    pending.value = false
  }
}

useSeoMeta({
  title: () => `${displayTitle.value} — Treewiki`,
  description: () => `Read ${displayTitle.value} on Treewiki, the free collaborative knowledge platform (CC BY-SA 4.0).`,
  ogTitle: () => `${displayTitle.value} — Treewiki`,
  ogDescription: () => `Read ${displayTitle.value} on Treewiki.`,
  ogType: 'article',
})

watch(slug, () => {
  fetchArticle()
})

onMounted(() => {
  fetchArticle()
})
</script>
