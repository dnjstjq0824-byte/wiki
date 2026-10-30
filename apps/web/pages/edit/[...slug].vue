<!-- apps/web/pages/edit/[...slug].vue -->
<template>
  <div class="space-y-6">
    <!-- Conflict Resolution View (if conflict occurs) -->
    <EditConflict
      v-if="hasConflict"
      :ours="sourceContent"
      :theirs="serverContent"
      @resolve="onConflictResolved"
      @cancel="hasConflict = false"
    />

    <!-- Main Editor Card -->
    <div v-else class="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl p-6 shadow-sm space-y-5">
      <div class="flex items-center justify-between border-b border-slate-200 dark:border-slate-800 pb-3">
        <div>
          <h1 class="text-xl font-bold text-slate-900 dark:text-slate-100">
            Editing: <span class="text-forest-700 dark:text-forest-400">{{ displayTitle }}</span>
          </h1>
          <p class="text-xs text-slate-500">Live synchronized dual-syntax Wikitext & Extended Markdown editor</p>
        </div>
        <div class="flex items-center gap-2">
          <NuxtLink :to="`/w/${slug}`" class="text-xs font-semibold text-slate-500 hover:text-slate-800 dark:hover:text-slate-200">
            ✕ Cancel
          </NuxtLink>
        </div>
      </div>

      <!-- Split Pane Component -->
      <SplitPaneEditor v-model="sourceContent" />

      <!-- Edit Summary & Save Action -->
      <div class="pt-4 border-t border-slate-200 dark:border-slate-800 space-y-3">
        <div v-if="saveError" class="p-3 rounded-lg bg-rose-50 dark:bg-rose-950/40 border border-rose-200 dark:border-rose-900 text-xs text-rose-700 dark:text-rose-300">
          {{ saveError }}
        </div>

        <div>
          <label class="block text-xs font-semibold text-slate-600 dark:text-slate-400 mb-1">
            Edit Summary (Briefly describe your changes)
          </label>
          <input
            v-model="summary"
            type="text"
            placeholder="e.g., Added definition, fixed typo, expanded citations, updated topic link"
            class="w-full px-3 py-2 text-xs rounded-lg border border-slate-300 dark:border-slate-700 bg-slate-50 dark:bg-slate-950 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-forest-500"
          />
        </div>

        <div class="flex items-center justify-between">
          <label class="inline-flex items-center gap-2 text-xs text-slate-600 dark:text-slate-400 cursor-pointer">
            <input type="checkbox" v-model="isMinor" class="rounded border-slate-300 text-forest-600 focus:ring-forest-500" />
            <span>This is a minor edit (formatting / typo fix)</span>
          </label>
          <div class="flex gap-2">
            <NuxtLink :to="`/w/${slug}`" class="px-4 py-2 rounded-lg border border-slate-300 dark:border-slate-700 text-xs font-semibold text-slate-700 dark:text-slate-300 hover:bg-slate-100">
              Cancel
            </NuxtLink>
            <button
              @click="handleSave"
              :disabled="saving"
              class="px-6 py-2 rounded-lg bg-forest-600 hover:bg-forest-500 text-white font-bold text-xs shadow-sm transition disabled:opacity-50"
            >
              {{ saving ? 'Publishing...' : 'Publish Revision' }}
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

onMounted(() => {
  fetchInitialContent()
})
</script>
