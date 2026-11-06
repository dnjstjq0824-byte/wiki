<!-- apps/web/pages/w/[...slug].vue -->
<template>
  <div>
    <WikiTabs :slug="slug" active="article" />

    <!-- Loading State -->
    <div v-if="pending" class="py-12 text-center text-sm text-slate-500">
      <p>Loading article…</p>
    </div>

    <!-- Red Link: Page Does Not Exist -->
    <div v-else-if="pageNotFound">
      <h1 class="font-wiki text-3xl text-slate-900 dark:text-slate-100 border-b border-wiki-border dark:border-slate-700 pb-1">
        {{ displayTitle }}
      </h1>
      <p class="text-sm text-slate-500 mt-1">From Treewiki, the free knowledge platform</p>

      <div class="mt-4 border border-wiki-border dark:border-slate-700 bg-wiki-pagebg dark:bg-slate-800/40 px-4 py-3 text-sm text-slate-700 dark:text-slate-300">
        <p>
          There is currently no article with the title <strong>“{{ displayTitle }}”</strong>.
          You can <NuxtLink :to="`/search?q=${encodeURIComponent(displayTitle)}`" class="text-wiki-link dark:text-sky-400 hover:underline">search for this title</NuxtLink>,
          browse related topic trees, or create the first revision:
        </p>
        <p class="mt-2">
          <NuxtLink :to="`/edit/${slug}`" class="text-wiki-link dark:text-sky-400 hover:underline font-semibold">
            Create “{{ displayTitle }}”
          </NuxtLink>
        </p>
      </div>
    </div>

    <!-- Service Error State -->
    <div v-else-if="loadError" class="text-center py-10">
      <h1 class="font-wiki text-2xl text-slate-900 dark:text-slate-100">Service temporarily unavailable</h1>
      <p class="text-sm text-slate-500 mt-2">{{ loadError }}</p>
      <button
        @click="fetchArticle"
        class="mt-4 px-4 py-1 text-sm border border-wiki-border dark:border-slate-700 bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700"
      >
        Retry
      </button>
    </div>

    <!-- Rendered Article View -->
    <article v-else>
      <h1 class="font-wiki text-3xl text-slate-900 dark:text-slate-100 border-b border-wiki-border dark:border-slate-700 pb-1">
        {{ displayTitle }}
      </h1>
      <p class="text-sm text-slate-500 mt-1">From Treewiki, the free knowledge platform · CC BY-SA 4.0</p>

      <div
        class="mt-3 text-[15px] leading-7 text-slate-900 dark:text-slate-200
          [&_h2]:font-wiki [&_h2]:text-2xl [&_h2]:border-b [&_h2]:border-wiki-border [&_h2]:dark:border-slate-700 [&_h2]:mt-6 [&_h2]:pb-1
          [&_h3]:font-bold [&_h3]:mt-4
          [&_a]:text-wiki-link [&_a]:dark:text-sky-400 [&_a]:hover:underline
          [&_p]:my-2"
        v-html="renderedBody"
      ></div>

      <div v-if="categories.length > 0" class="mt-6 border border-wiki-border dark:border-slate-700 bg-wiki-pagebg dark:bg-slate-800/40 px-3 py-2 text-sm">
        <span class="font-semibold">Categories:&#32;</span><span v-for="(cat, i) in categories" :key="cat"><span v-if="i > 0"> · </span><NuxtLink :to="`/search?q=${encodeURIComponent(cat)}`" class="text-wiki-link dark:text-sky-400 hover:underline">{{ cat }}</NuxtLink></span>
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
