<template>
  <div class="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl p-6 shadow-sm space-y-4">
    <div class="flex items-center justify-between border-b border-slate-200 dark:border-slate-800 pb-3">
      <div>
        <h1 class="text-xl font-bold text-slate-900 dark:text-slate-100">
          Editing: <span class="text-forest-700 dark:text-forest-400">{{ displayTitle }}</span>
        </h1>
        <p class="text-xs text-slate-500">Dual-syntax Wikitext & Extended Markdown supported</p>
      </div>
      <NuxtLink :to="`/w/${slug}`" class="text-xs font-semibold text-slate-500 hover:text-slate-800 dark:hover:text-slate-200">
        ✕ Cancel
      </NuxtLink>
    </div>

    <!-- Formatting Toolbar -->
    <div class="flex flex-wrap items-center gap-1.5 p-2 bg-slate-100 dark:bg-slate-800/80 rounded-lg text-xs font-semibold">
      <button @click="insertText('\'\'\'', '\'\'\'')" class="px-2.5 py-1 bg-white dark:bg-slate-700 rounded border border-slate-200 dark:border-slate-600 hover:bg-slate-50">B</button>
      <button @click="insertText('\'\'', '\'\'')" class="px-2.5 py-1 bg-white dark:bg-slate-700 rounded border border-slate-200 dark:border-slate-600 italic hover:bg-slate-50">I</button>
      <button @click="insertText('== ', ' ==')" class="px-2.5 py-1 bg-white dark:bg-slate-700 rounded border border-slate-200 dark:border-slate-600 hover:bg-slate-50">H2</button>
      <button @click="insertText('[[', ']]')" class="px-2.5 py-1 bg-white dark:bg-slate-700 rounded border border-slate-200 dark:border-slate-600 hover:bg-slate-50">Link [[ ]]</button>
      <button @click="insertText('<ref>', '</ref>')" class="px-2.5 py-1 bg-white dark:bg-slate-700 rounded border border-slate-200 dark:border-slate-600 hover:bg-slate-50">Ref &lt;ref&gt;</button>
      <button @click="insertText('[[Category:', ']]')" class="px-2.5 py-1 bg-white dark:bg-slate-700 rounded border border-slate-200 dark:border-slate-600 hover:bg-slate-50">Category</button>
    </div>

    <!-- Dual Pane Editor -->
    <div class="grid grid-cols-1 md:grid-cols-2 gap-4 h-[420px]">
      <!-- Textarea input -->
      <div class="flex flex-col">
        <label class="text-xs font-semibold text-slate-500 mb-1">Source Code</label>
        <textarea
          ref="editorRef"
          v-model="sourceContent"
          class="flex-1 w-full p-3 font-mono text-xs rounded-lg border border-slate-300 dark:border-slate-700 bg-slate-50 dark:bg-slate-950 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-forest-500 resize-none"
        ></textarea>
      </div>

      <!-- Live Preview -->
      <div class="flex flex-col">
        <label class="text-xs font-semibold text-slate-500 mb-1">Live Render Preview</label>
        <div class="flex-1 p-4 rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900/60 overflow-y-auto text-sm leading-relaxed prose dark:prose-invert">
          <div v-html="renderedPreview"></div>
        </div>
      </div>
    </div>

    <!-- Edit Summary & Save Action -->
    <div class="pt-4 border-t border-slate-200 dark:border-slate-800 space-y-3">
      <div>
        <label class="block text-xs font-semibold text-slate-600 dark:text-slate-400 mb-1">Edit Summary (Brief explanation of changes)</label>
        <input
          v-model="summary"
          type="text"
          placeholder="e.g., Added definition, fixed typo, updated citation"
          class="w-full px-3 py-1.5 text-xs rounded-lg border border-slate-300 dark:border-slate-700 bg-slate-50 dark:bg-slate-950 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-forest-500"
        />
      </div>

      <div class="flex items-center justify-between">
        <label class="inline-flex items-center gap-2 text-xs text-slate-600 dark:text-slate-400 cursor-pointer">
          <input type="checkbox" v-model="isMinor" class="rounded border-slate-300 text-forest-600 focus:ring-forest-500" />
          <span>This is a minor edit (formatting / typo)</span>
        </label>
        <div class="flex gap-2">
          <NuxtLink :to="`/w/${slug}`" class="px-4 py-2 rounded-lg border border-slate-300 dark:border-slate-700 text-xs font-semibold text-slate-700 dark:text-slate-300 hover:bg-slate-100">
            Cancel
          </NuxtLink>
          <button @click="handleSave" class="px-5 py-2 rounded-lg bg-forest-600 hover:bg-forest-700 text-white font-semibold text-xs shadow-sm transition">
            Publish Revision
          </button>
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
const editorRef = ref<HTMLTextAreaElement | null>(null)

const slug = computed(() => {
  const p = route.params.slug
  return Array.isArray(p) ? p.join('/') : p || 'Main_Page'
})

const displayTitle = computed(() => slug.value.replace(/_/g, ' '))

const sourceContent = ref(`== Overview ==
'''${displayTitle.value}''' is a foundational domain in structured knowledge systems.

== Key Principles ==
* Principle of Verifiability: Information must be backed by reliable sources.
* Principle of NPOV: Multiple valid academic viewpoints must be fairly documented.

[[Category:General Knowledge]]
`)

const summary = ref('')
const isMinor = ref(false)

const renderedPreview = computed(() => {
  // Simple client-side preview regex
  let t = sourceContent.value
  t = t.replace(/== (.*?) ==/g, '<h2 class="text-lg font-bold border-b pb-1 my-2">$1</h2>')
  t = t.replace(/'''(.*?)'''/g, '<strong>$1</strong>')
  t = t.replace(/''(.*?)''/g, '<em>$1</em>')
  t = t.replace(/\[\[Category:(.*?)\]\]/g, '<span class="text-xs bg-slate-100 p-1 rounded mr-1">📁 Category: $1</span>')
  return t
})

const insertText = (prefix: string, suffix: string) => {
  if (!editorRef.value) return
  const el = editorRef.value
  const start = el.selectionStart
  const end = el.selectionEnd
  const val = sourceContent.value
  const selected = val.substring(start, end)
  sourceContent.value = val.substring(0, start) + prefix + selected + suffix + val.substring(end)
}

const handleSave = () => {
  alert('Revision successfully staged to immutable ledger!')
  router.push(`/w/${slug.value}`)
}
</script>
