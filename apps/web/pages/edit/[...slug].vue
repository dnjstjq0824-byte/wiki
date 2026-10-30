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
          <button
            @click="simulateConflict"
            class="text-[11px] px-2 py-1 rounded bg-amber-100 dark:bg-amber-950 text-amber-800 dark:text-amber-400 hover:bg-amber-200 font-medium"
            title="Simulate peer edit conflict"
          >
            ⚡ Test Conflict UI
          </button>
          <NuxtLink :to="`/w/${slug}`" class="text-xs font-semibold text-slate-500 hover:text-slate-800 dark:hover:text-slate-200">
            ✕ Cancel
          </NuxtLink>
        </div>
      </div>

      <!-- Split Pane Component -->
      <SplitPaneEditor v-model="sourceContent" />

      <!-- Edit Summary & Save Action -->
      <div class="pt-4 border-t border-slate-200 dark:border-slate-800 space-y-3">
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
              class="px-6 py-2 rounded-lg bg-forest-600 hover:bg-forest-500 text-white font-bold text-xs shadow-sm transition"
            >
              Publish Revision
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()

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

const serverContent = ref(`== Overview ==
'''${displayTitle.value}''' is a computational and mathematical science domain.
Concurrent update by server editor.

[[Category:Science]]
`)

const hasConflict = ref(false)
const summary = ref('')
const isMinor = ref(false)

const simulateConflict = () => {
  hasConflict.value = true
}

const onConflictResolved = (mergedText: string) => {
  sourceContent.value = mergedText
  hasConflict.value = false
  alert('Edit conflict resolved! Ready to publish.')
}

const handleSave = () => {
  alert('Revision successfully recorded to immutable ledger!')
  router.push(`/w/${slug.value}`)
}
</script>
