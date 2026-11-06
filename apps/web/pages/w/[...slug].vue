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
      <h1 class="text-3xl font-bold text-slate-900 dark:text-slate-100 border-b-2 border-namu-600 dark:border-namu-500 pb-1">
        {{ displayTitle }}
      </h1>
      <p class="text-sm text-slate-500 mt-1">This article does not exist yet. You can create it below.</p>

      <div class="mt-4 border border-wiki-border dark:border-slate-700 bg-wiki-pagebg dark:bg-slate-800/40 px-4 py-3 text-sm text-slate-700 dark:text-slate-300">
        <p>
          There is currently no article with the title <strong>“{{ displayTitle }}”</strong>.
          You can <NuxtLink :to="`/search?q=${encodeURIComponent(displayTitle)}`" class="text-wiki-link dark:text-sky-400 hover:underline">search for this title</NuxtLink>,
          browse related topic trees, or create the first revision:
        </p>
        <p class="mt-3">
          <NuxtLink :to="`/edit/${slug}`" class="inline-block px-4 py-1.5 bg-namu-600 hover:bg-namu-700 text-white text-sm font-semibold">
            Create “{{ displayTitle }}”
          </NuxtLink>
        </p>
      </div>
    </div>

    <!-- Service Error State -->
    <div v-else-if="loadError" class="text-center py-10">
      <h1 class="text-2xl font-bold text-slate-900 dark:text-slate-100">Service temporarily unavailable</h1>
      <p class="text-sm text-slate-500 mt-2">{{ loadError }}</p>
      <button
        @click="fetchArticle"
        class="mt-4 px-4 py-1.5 text-sm bg-namu-600 hover:bg-namu-700 text-white"
      >
        Retry
      </button>
    </div>

    <!-- Rendered Article View -->
    <article v-else>
      <h1 class="text-3xl font-bold text-slate-900 dark:text-slate-100 border-b-2 border-namu-600 dark:border-namu-500 pb-1">
        {{ displayTitle }}
      </h1>
      <p class="text-[13px] text-slate-500 mt-1">From Treewiki, the free knowledge platform · CC BY-SA 4.0</p>

      <!-- Table of Contents -->
      <nav v-if="toc.length > 0" aria-label="Contents" class="md:float-right md:ml-4 my-3 border border-wiki-border dark:border-slate-700 bg-white dark:bg-slate-900 text-sm min-w-[12rem]">
        <p class="px-3 py-1.5 bg-namu-600 text-white font-bold">Contents</p>
        <ol class="px-3 py-2 space-y-1">
          <li v-for="(entry, i) in toc" :key="entry.id" :class="entry.level === 3 ? 'pl-4' : ''">
            <a :href="`#${entry.id}`" class="text-wiki-link dark:text-sky-400 hover:underline">
              <span class="text-slate-400 mr-1">{{ i + 1 }}.</span>{{ entry.text }}
            </a>
          </li>
        </ol>
      </nav>

      <div
        ref="contentRef"
        class="mt-3 text-[15px] leading-7 text-slate-900 dark:text-slate-200
          [&_h2]:text-[22px] [&_h2]:font-bold [&_h2]:border-b [&_h2]:border-namu-600 [&_h2]:dark:border-namu-500 [&_h2]:mt-6 [&_h2]:pb-1
          [&_h3]:text-lg [&_h3]:font-bold [&_h3]:mt-4
          [&_a]:text-wiki-link [&_a]:dark:text-sky-400 [&_a]:hover:underline
          [&_p]:my-2 [&_ul]:list-disc [&_ul]:pl-6 [&_ul]:my-2"
        v-html="renderedBody"
      ></div>

      <div v-if="categories.length > 0" class="mt-6 border border-wiki-border dark:border-slate-700 text-sm clear-both">
        <p class="px-3 py-1.5 bg-slate-100 dark:bg-slate-800 border-b border-wiki-border dark:border-slate-700 font-bold">Categories</p>
        <p class="px-3 py-2">
          <span v-for="(cat, i) in categories" :key="cat"><span v-if="i > 0"> · </span><NuxtLink :to="`/search?q=${encodeURIComponent(cat)}`" class="text-wiki-link dark:text-sky-400 hover:underline">{{ cat }}</NuxtLink></span>
        </p>
      </div>
    </article>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch, nextTick } from 'vue'
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
const contentRef = ref<HTMLElement | null>(null)

interface TocEntry { id: string; text: string; level: number }
const toc = ref<TocEntry[]>([])

const buildToc = () => {
  toc.value = []
  if (!contentRef.value) return
  const heads = contentRef.value.querySelectorAll('h2, h3')
  heads.forEach((el, i) => {
    const id = `section-${i}`
    el.setAttribute('id', id)
    toc.value.push({ id, text: (el.textContent || '').trim().slice(0, 60), level: el.tagName === 'H3' ? 3 : 2 })
  })
}

const fetchArticle = async () => {
  pending.value = true
  pageNotFound.value = false
  loadError.value = ''
  toc.value = []

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
  // Build TOC only after the article branch has rendered.
  if (!pageNotFound.value && !loadError.value) {
    await nextTick()
    buildToc()
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
