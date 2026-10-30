<!-- apps/web/components/DiffViewer.vue -->
<template>
  <div class="rounded-xl border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 overflow-hidden shadow-sm">
    <!-- Diff Header Toolbar -->
    <div class="px-4 py-3 bg-slate-50 dark:bg-slate-800/80 border-b border-slate-200 dark:border-slate-700 flex flex-wrap items-center justify-between gap-3 text-xs">
      <div class="flex items-center gap-3">
        <span class="font-bold text-slate-700 dark:text-slate-200 flex items-center gap-1.5">
          <span>⚖️</span> Revision Comparison
        </span>
        <span class="px-2 py-0.5 rounded bg-emerald-100 dark:bg-emerald-950 text-emerald-700 dark:text-emerald-400 font-mono font-semibold">
          +{{ additionsCount }}
        </span>
        <span class="px-2 py-0.5 rounded bg-rose-100 dark:bg-rose-950 text-rose-700 dark:text-rose-400 font-mono font-semibold">
          -{{ deletionsCount }}
        </span>
      </div>

      <!-- Mode Toggle -->
      <div class="flex items-center gap-1 bg-slate-200 dark:bg-slate-700 p-0.5 rounded-lg font-medium">
        <button
          @click="viewMode = 'unified'"
          :class="viewMode === 'unified' ? 'bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 shadow-xs' : 'text-slate-600 dark:text-slate-300'"
          class="px-2.5 py-1 rounded transition"
        >
          Unified View
        </button>
        <button
          @click="viewMode = 'split'"
          :class="viewMode === 'split' ? 'bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 shadow-xs' : 'text-slate-600 dark:text-slate-300'"
          class="px-2.5 py-1 rounded transition"
        >
          Side-by-Side (Split)
        </button>
      </div>
    </div>

    <!-- 1. Unified Inline View -->
    <div v-if="viewMode === 'unified'" class="font-mono text-xs overflow-x-auto">
      <div
        v-for="(chunk, idx) in formattedChunks"
        :key="idx"
        class="flex items-stretch border-b border-slate-100 dark:border-slate-800/50 hover:brightness-95 transition"
        :class="getUnifiedRowClass(chunk.op)"
      >
        <span class="w-12 py-1 px-2 text-right text-slate-400 select-none bg-slate-50 dark:bg-slate-800/30 border-r border-slate-200 dark:border-slate-800">
          {{ chunk.oldLine ?? '' }}
        </span>
        <span class="w-12 py-1 px-2 text-right text-slate-400 select-none bg-slate-50 dark:bg-slate-800/30 border-r border-slate-200 dark:border-slate-800">
          {{ chunk.newLine ?? '' }}
        </span>
        <span class="w-8 py-1 text-center font-bold select-none" :class="getOpSignColor(chunk.op)">
          {{ chunk.op === 1 ? '+' : chunk.op === 2 ? '-' : ' ' }}
        </span>
        <span class="flex-1 py-1 px-2 whitespace-pre-wrap break-all text-slate-800 dark:text-slate-200">
          {{ chunk.text }}
        </span>
      </div>
    </div>

    <!-- 2. Side-by-Side Split View -->
    <div v-else class="font-mono text-xs grid grid-cols-2 divide-x divide-slate-200 dark:divide-slate-800 overflow-x-auto">
      <!-- Left Column: Old / Deleted -->
      <div>
        <div class="p-2 bg-slate-100 dark:bg-slate-800 font-bold text-slate-500 border-b border-slate-200 dark:border-slate-700">
          Previous Revision (Base)
        </div>
        <div
          v-for="(row, idx) in splitRows"
          :key="`left-${idx}`"
          class="flex items-stretch min-h-[24px] border-b border-slate-100 dark:border-slate-800/40"
          :class="row.left?.op === 2 ? 'bg-rose-50 dark:bg-rose-950/40' : ''"
        >
          <span class="w-10 py-1 px-1 text-right text-slate-400 select-none bg-slate-50 dark:bg-slate-800/30 border-r border-slate-200 dark:border-slate-800">
            {{ row.left?.line ?? '' }}
          </span>
          <span class="flex-1 py-1 px-2 whitespace-pre-wrap break-all text-slate-800 dark:text-slate-200">
            {{ row.left?.text ?? '' }}
          </span>
        </div>
      </div>

      <!-- Right Column: New / Added -->
      <div>
        <div class="p-2 bg-slate-100 dark:bg-slate-800 font-bold text-slate-500 border-b border-slate-200 dark:border-slate-700">
          Current Revision (Target)
        </div>
        <div
          v-for="(row, idx) in splitRows"
          :key="`right-${idx}`"
          class="flex items-stretch min-h-[24px] border-b border-slate-100 dark:border-slate-800/40"
          :class="row.right?.op === 1 ? 'bg-emerald-50 dark:bg-emerald-950/40' : ''"
        >
          <span class="w-10 py-1 px-1 text-right text-slate-400 select-none bg-slate-50 dark:bg-slate-800/30 border-r border-slate-200 dark:border-slate-800">
            {{ row.right?.line ?? '' }}
          </span>
          <span class="flex-1 py-1 px-2 whitespace-pre-wrap break-all text-slate-800 dark:text-slate-200">
            {{ row.right?.text ?? '' }}
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'

export interface DiffChunk {
  op: number // 0 = Equal, 1 = Insert, 2 = Delete
  text: string
}

const props = withDefaults(
  defineProps<{
    chunks?: DiffChunk[]
  }>(),
  {
    chunks: () => [
      { op: 0, text: '== Introduction ==' },
      { op: 2, text: 'Quantum computing is an experimental laboratory topic.' },
      { op: 1, text: 'Quantum computing is a rapidly-emerging commercial and scientific discipline.' },
      { op: 0, text: 'It leverages qubits, superposition, and quantum entanglement.' },
      { op: 1, text: 'Algorithms such as Shor and Grover demonstrate asymptotic polynomial and quadratic speedups.' },
      { op: 0, text: '[[Category:Science]]' },
    ],
  }
)

const viewMode = ref<'unified' | 'split'>('unified')

const additionsCount = computed(() => props.chunks.filter((c) => c.op === 1).length)
const deletionsCount = computed(() => props.chunks.filter((c) => c.op === 2).length)

interface FormattedChunk extends DiffChunk {
  oldLine?: number
  newLine?: number
}

const formattedChunks = computed(() => {
  let oldLineCounter = 1
  let newLineCounter = 1

  return props.chunks.map((chunk) => {
    const item: FormattedChunk = { ...chunk }
    if (chunk.op === 0) {
      item.oldLine = oldLineCounter++
      item.newLine = newLineCounter++
    } else if (chunk.op === 2) {
      item.oldLine = oldLineCounter++
    } else if (chunk.op === 1) {
      item.newLine = newLineCounter++
    }
    return item
  })
})

interface SplitRow {
  left?: { line: number; text: string; op: number }
  right?: { line: number; text: string; op: number }
}

const splitRows = computed(() => {
  const rows: SplitRow[] = []
  let oldL = 1
  let newL = 1

  for (let i = 0; i < props.chunks.length; i++) {
    const chunk = props.chunks[i]
    if (chunk.op === 0) {
      rows.push({
        left: { line: oldL++, text: chunk.text, op: 0 },
        right: { line: newL++, text: chunk.text, op: 0 },
      })
    } else if (chunk.op === 2) {
      // Check if followed by insert (replacement)
      const next = props.chunks[i + 1]
      if (next && next.op === 1) {
        rows.push({
          left: { line: oldL++, text: chunk.text, op: 2 },
          right: { line: newL++, text: next.text, op: 1 },
        })
        i++ // skip next
      } else {
        rows.push({
          left: { line: oldL++, text: chunk.text, op: 2 },
        })
      }
    } else if (chunk.op === 1) {
      rows.push({
        right: { line: newL++, text: chunk.text, op: 1 },
      })
    }
  }

  return rows
})

const getUnifiedRowClass = (op: number) => {
  if (op === 1) return 'bg-emerald-50/70 dark:bg-emerald-950/40 text-emerald-950 dark:text-emerald-200'
  if (op === 2) return 'bg-rose-50/70 dark:bg-rose-950/40 text-rose-950 dark:text-rose-200'
  return ''
}

const getOpSignColor = (op: number) => {
  if (op === 1) return 'text-emerald-600 dark:text-emerald-400'
  if (op === 2) return 'text-rose-600 dark:text-rose-400'
  return 'text-slate-300 dark:text-slate-600'
}
</script>
