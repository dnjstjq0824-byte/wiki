<!-- apps/web/components/SplitPaneEditor.vue -->
<template>
  <div class="border border-wiki-border dark:border-slate-700 bg-white dark:bg-slate-900 flex flex-col h-[650px]">
    <!-- Formatting Toolbar -->
    <div class="px-3 py-2 bg-slate-100 dark:bg-slate-800 border-b border-wiki-border dark:border-slate-700 flex flex-wrap items-center justify-between gap-2 text-[13px]">
      <div class="flex flex-wrap items-center gap-1">
        <button
          @click="insertFormat('\'\'\'', '\'\'\'')"
          class="px-2.5 py-1 bg-white dark:bg-slate-700 border border-wiki-border dark:border-slate-600 hover:bg-slate-100 text-slate-800 dark:text-slate-200"
          title="Bold ('''text''')"
        >
          <strong>B</strong>
        </button>
        <button
          @click="insertFormat('\'\'', '\'\'')"
          class="px-2.5 py-1 bg-white dark:bg-slate-700 border border-wiki-border dark:border-slate-600 hover:bg-slate-100 text-slate-800 dark:text-slate-200"
          title="Italic (''text'')"
        >
          <em class="font-wiki">I</em>
        </button>
        <button
          @click="insertFormat('== ', ' ==')"
          class="px-2.5 py-1 bg-white dark:bg-slate-700 border border-wiki-border dark:border-slate-600 hover:bg-slate-100 text-slate-800 dark:text-slate-200"
          title="Section Heading H2"
        >
          H2
        </button>
        <button
          @click="insertFormat('=== ', ' ===')"
          class="px-2.5 py-1 bg-white dark:bg-slate-700 border border-wiki-border dark:border-slate-600 hover:bg-slate-100 text-slate-800 dark:text-slate-200"
          title="Subsection Heading H3"
        >
          H3
        </button>
        <button
          @click="insertFormat('[[', ']]')"
          class="px-2.5 py-1 bg-white dark:bg-slate-700 border border-wiki-border dark:border-slate-600 hover:bg-slate-100 text-wiki-link dark:text-sky-400"
          title="Wiki Internal Link [[Target|Label]]"
        >
          [[ Link ]]
        </button>
        <button
          @click="insertFormat('<ref>', '</ref>')"
          class="px-2.5 py-1 bg-white dark:bg-slate-700 border border-wiki-border dark:border-slate-600 hover:bg-slate-100 text-sky-700 dark:text-sky-400"
          title="Citation Reference <ref>Source</ref>"
        >
          &lt;ref&gt;
        </button>
        <button
          @click="insertFormat('[[Category:', ']]')"
          class="px-2.5 py-1 bg-white dark:bg-slate-700 border border-wiki-border dark:border-slate-600 hover:bg-slate-100 text-slate-700 dark:text-slate-300"
          title="Add Category"
        >
          Category
        </button>
        <button
          @click="insertFormat('$$\n', '\n$$')"
          class="px-2.5 py-1 bg-white dark:bg-slate-700 border border-wiki-border dark:border-slate-600 hover:bg-slate-100 text-amber-700 dark:text-amber-400 font-mono"
          title="LaTeX Math Formula"
        >
          ∑ Math
        </button>
      </div>

      <!-- Live Statistics -->
      <div class="flex items-center gap-3 text-slate-500 text-xs font-mono">
        <span>{{ wordCount }} words</span>
        <span>{{ text.length }} chars</span>
        <span>{{ lineCount }} lines</span>
      </div>
    </div>

    <!-- Dual Pane Body -->
    <div class="flex-1 grid grid-cols-1 md:grid-cols-2 divide-y md:divide-y-0 md:divide-x divide-slate-200 dark:divide-slate-800 overflow-hidden">
      <!-- Left: Raw Wikitext Code Input -->
      <div class="flex flex-col h-full bg-slate-50 dark:bg-slate-950/40">
        <div class="px-3 py-1 bg-slate-100 dark:bg-slate-800/40 border-b border-wiki-border dark:border-slate-800 text-xs text-slate-500">
          Source
        </div>
        <textarea
          ref="textareaRef"
          v-model="text"
          placeholder="Type wikitext or markdown here..."
          class="flex-1 p-4 font-mono text-[13px] text-slate-900 dark:text-slate-100 bg-transparent resize-none focus:outline-none leading-relaxed"
          @keydown.tab.prevent="insertTab"
        ></textarea>
      </div>

      <!-- Right: Real-time HTML Preview -->
      <div class="flex flex-col h-full bg-white dark:bg-slate-900 overflow-hidden">
        <div class="px-3 py-1 bg-slate-100 dark:bg-slate-800/40 border-b border-wiki-border dark:border-slate-800 text-xs text-slate-500 flex items-center justify-between">
          <span>Preview</span>
          <span class="text-emerald-700 dark:text-emerald-400">Live</span>
        </div>
        <div class="flex-1 p-6 overflow-y-auto text-slate-800 dark:text-slate-200 text-sm leading-relaxed">
          <div v-html="renderedHtml"></div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'

const props = withDefaults(
  defineProps<{
    modelValue?: string
  }>(),
  {
    modelValue: `== Quantum Computing ==
'''Quantum computing''' is a multidisciplinary field comprising aspects of computer science, physics, and mathematics that utilizes [[Quantum Mechanics|quantum mechanics]] to solve complex problems faster than on classical computers.

=== Superposition and Entanglement ===
Unlike classical bits that are either 0 or 1, a ''qubit'' can exist in a multidimensional state of superposition:
$$\\lvert \\psi \\rangle = \\alpha \\lvert 0 \\rangle + \\beta \\lvert 1 \\rangle$$

Notable algorithms include Shor's algorithm for integer factorization<ref>P. W. Shor, Algorithms for quantum computation: discrete logarithms and factoring, IEEE 1994.</ref>.

[[Category:Quantum Information]]
[[Category:Computer Science]]
`,
  }
)

const emit = defineEmits<{
  (e: 'update:modelValue', val: string): void
}>()

const textareaRef = ref<HTMLTextAreaElement | null>(null)

const text = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val),
})

const wordCount = computed(() => {
  const matches = text.value.trim().match(/\S+/g)
  return matches ? matches.length : 0
})

const lineCount = computed(() => {
  return text.value.split('\n').length
})

const insertFormat = (prefix: string, suffix: string) => {
  const el = textareaRef.value
  if (!el) return
  const start = el.selectionStart
  const end = el.selectionEnd
  const current = text.value
  const selected = current.substring(start, end)
  text.value = current.substring(0, start) + prefix + selected + suffix + current.substring(end)
  // Restore cursor position
  setTimeout(() => {
    el.focus()
    el.setSelectionRange(start + prefix.length, end + prefix.length)
  }, 0)
}

const insertTab = () => {
  insertFormat('  ', '')
}

// Client-side instant regex parser for fast preview
const renderedHtml = computed(() => {
  let html = text.value

  // Headings
  html = html.replace(/^=== (.*?) ===$/gm, '<h3 class="text-base font-bold mt-4 mb-2">$1</h3>')
  html = html.replace(/^== (.*?) ==$/gm, '<h2 class="text-xl font-bold mt-5 mb-3 pb-1 border-b-2 border-namu-600 dark:border-namu-500">$1</h2>')

  // Bold & Italic
  html = html.replace(/'''(.*?)'''/g, '<strong class="font-bold">$1</strong>')
  html = html.replace(/''(.*?)''/g, '<em class="italic">$1</em>')

  // Internal Links [[Target|Label]] or [[Target]]
  html = html.replace(/\[\[([^\]|]+)(?:\|([^\]]+))?\]\]/g, (_, target, label) => {
    const textLabel = label || target
    return `<a href="/w/${encodeURIComponent(target.trim().replace(/ /g, '_'))}" class="text-wiki-link dark:text-sky-400 hover:underline">${textLabel}</a>`
  })

  // Footnote <ref>
  let refCount = 1
  const references: string[] = []
  html = html.replace(/<ref>(.*?)<\/ref>/g, (_, refText) => {
    references.push(refText)
    const idx = refCount++
    return `<sup class="text-wiki-link font-semibold">[${idx}]</sup>`
  })

  // Categories
  html = html.replace(/\[\[Category:(.*?)\]\]/g, (_, cat) => {
    return `<span class="inline-block px-2 py-0.5 text-xs bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300 mr-1.5 mb-1">${cat}</span>`
  })

  // Math Blocks
  html = html.replace(/\$\$([\s\S]*?)\$\$/g, (_, math) => {
    return `<div class="p-3 my-2 border border-wiki-border dark:border-slate-700 bg-wiki-pagebg dark:bg-slate-800 font-mono text-xs text-center overflow-x-auto">${math.trim()}</div>`
  })

  // Append References list if exists
  if (references.length > 0) {
    html += '<div class="mt-6 pt-4 border-t border-wiki-border dark:border-slate-700"><h4 class="text-xs font-bold text-slate-500 mb-2">References</h4><ol class="list-decimal list-inside text-xs text-slate-500 space-y-1">'
    references.forEach((r) => {
      html += `<li>${r}</li>`
    })
    html += '</ol></div>'
  }

  return html
})
</script>
