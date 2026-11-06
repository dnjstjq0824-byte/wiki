<!-- apps/web/pages/edit/[...slug].vue -->
<template>
  <div>
    <WikiTabs :slug="slug" active="edit" />

    <!-- Conflict Resolution View (if conflict occurs) -->
    <EditConflict
      v-if="hasConflict"
      :ours="sourceContent"
      :theirs="serverContent"
      @resolve="onConflictResolved"
      @cancel="hasConflict = false"
    />

    <div v-else>
      <h1 class="text-3xl font-bold text-slate-900 dark:text-slate-100 border-b-2 border-namu-600 dark:border-namu-500 pb-1">
        Editing: {{ displayTitle }}
      </h1>
      <p class="text-sm text-slate-500 mt-1">
        <NuxtLink :to="`/w/${slug}`" class="text-wiki-link dark:text-sky-400 hover:underline">← Back to article</NuxtLink>
        · Your changes are released under CC BY-SA 4.0
      </p>

      <!-- Split Pane Component -->
      <div class="mt-4">
        <SplitPaneEditor v-model="sourceContent" />
      </div>

      <!-- Edit Summary & Save Action -->
      <div class="mt-4 space-y-3">
        <div v-if="saveError" class="border border-red-300 dark:border-red-900 bg-red-50 dark:bg-red-950/40 px-3 py-2 text-sm text-red-700 dark:text-red-300">
          {{ saveError }}
        </div>

        <div>
          <label class="block text-sm font-semibold text-slate-700 dark:text-slate-300 mb-1">
            Edit summary
          </label>
          <input
            v-model="summary"
            type="text"
            placeholder="e.g., Added definition, fixed typo, expanded citations, updated topic link"
            class="w-full px-2.5 py-1.5 text-sm border border-wiki-border dark:border-slate-700 bg-white dark:bg-slate-950 text-slate-900 dark:text-slate-100 focus:outline-none focus:border-namu-600"
          />
        </div>

        <div class="flex flex-wrap items-center justify-between gap-3">
          <label class="inline-flex items-center gap-2 text-sm text-slate-600 dark:text-slate-400 cursor-pointer">
            <input type="checkbox" v-model="isMinor" />
            <span>This is a minor edit</span>
          </label>
          <div class="flex gap-2 text-sm">
            <NuxtLink :to="`/w/${slug}`" class="px-4 py-1.5 border border-wiki-border dark:border-slate-700 hover:bg-slate-100 dark:hover:bg-slate-800">
              Cancel
            </NuxtLink>
            <button
              @click="handleSave"
              :disabled="saving"
              class="px-6 py-1.5 bg-namu-600 hover:bg-namu-700 text-white font-semibold disabled:opacity-50"
            >
              {{ saving ? 'Publishing…' : 'Publish Revision' }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useApi } from '../../composables/useApi'

const route = useRoute()
const router = useRouter()
const api = useApi()

const slug = computed(() => {
  const p = route.params.slug
  return Array.isArray(p) ? p.join('/') : p || 'Main_Page'
})

const displayTitle = computed(() => slug.value.replace(/_/g, ' '))

const sourceContent = ref(`== Overview ==
'''${displayTitle.value}''' is a core discipline within modern encyclopedic and epistemological research.

=== Foundational Principles ===
* Principle of Verifiability: Information must be backed by reliable sources.
* Principle of NPOV: Diverse academic viewpoints are documented with neutral balance.

[[Category:General Knowledge]]
`)

const serverContent = ref('')
const parentHash = ref('')
const hasConflict = ref(false)
const summary = ref('')
const isMinor = ref(false)
const saving = ref(false)
const saveError = ref('')

const fetchInitialContent = async () => {
  try {
    const res: any = await api.get(`/pages/0/${slug.value}`)
    const page = res?.page || (res?.title ? res : null)
    if (page && page.content) {
      sourceContent.value = page.content
      parentHash.value = page.current_revision_id || ''
    }
  } catch {
    // New page: keep default template
  }
}

const onConflictResolved = (mergedText: string) => {
  sourceContent.value = mergedText
  hasConflict.value = false
}

const handleSave = async () => {
  saving.value = true
  saveError.value = ''

  try {
    const headers: Record<string, string> = {}
    if (parentHash.value) {
      headers['If-Match'] = parentHash.value
    }

    await api.put(
      `/pages/0/${slug.value}`,
      {
        wikitext: sourceContent.value,
        summary: summary.value || 'Updated page content',
        is_minor: isMinor.value,
      },
      headers
    )

    router.push(`/w/${slug.value}`)
  } catch (err: any) {
    if (err.status === 409 || err.statusCode === 409) {
      serverContent.value = err.data?.latest_content || 'Concurrent update occurred.'
      hasConflict.value = true
    } else {
      saveError.value = err.data?.error || err.message || 'Failed to save revision. Please ensure you are signed in.'
    }
  } finally {
    saving.value = false
  }
}

useSeoMeta({
  title: () => `Editing ${displayTitle.value} — Treewiki`,
  description: 'Wikitext and Markdown article editor with live preview and conflict resolution.',
  robots: 'noindex',
})

onMounted(() => {
  fetchInitialContent()
})
</script>
