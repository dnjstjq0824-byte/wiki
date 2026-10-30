<!-- apps/web/components/EditConflict.vue -->
<template>
  <div class="rounded-xl border border-amber-300 dark:border-amber-800/80 bg-amber-50/50 dark:bg-amber-950/20 p-6 space-y-6 shadow-sm">
    <!-- Conflict Banner -->
    <div class="flex items-start gap-3 border-b border-amber-200 dark:border-amber-900/60 pb-4">
      <span class="text-2xl">⚠️</span>
      <div>
        <h2 class="text-base font-bold text-amber-900 dark:text-amber-200">
          Concurrent Edit Conflict Detected
        </h2>
        <p class="text-xs text-amber-800/80 dark:text-amber-300/80 leading-relaxed mt-0.5">
          Another user authored and committed a revision while you were composing your changes.
          Please review the diff below and choose how to resolve the overlapping text before publishing.
        </p>
      </div>
    </div>

    <!-- Quick Action Resolution Presets -->
    <div class="flex flex-wrap items-center gap-2">
      <span class="text-xs font-semibold text-slate-700 dark:text-slate-300">Quick Resolution:</span>
      <button
        @click="useOurs"
        class="px-3 py-1.5 text-xs font-semibold rounded-lg bg-white dark:bg-slate-800 border border-slate-300 dark:border-slate-700 hover:bg-slate-50 text-slate-800 dark:text-slate-200 shadow-2xs transition"
      >
        Keep My Version (Ours)
      </button>
      <button
        @click="useTheirs"
        class="px-3 py-1.5 text-xs font-semibold rounded-lg bg-white dark:bg-slate-800 border border-slate-300 dark:border-slate-700 hover:bg-slate-50 text-slate-800 dark:text-slate-200 shadow-2xs transition"
      >
        Keep Server Version (Theirs)
      </button>
      <button
        @click="useCombined"
        class="px-3 py-1.5 text-xs font-semibold rounded-lg bg-forest-600 hover:bg-forest-500 text-white shadow-2xs transition"
      >
        Auto-Merge Both
      </button>
    </div>

    <!-- Side-by-Side Comparison of Versions -->
    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <!-- Your Local Edit -->
      <div class="space-y-1.5">
        <div class="flex items-center justify-between text-xs font-semibold text-emerald-800 dark:text-emerald-400">
          <span>Your Edits (Local Draft)</span>
          <span class="text-[11px] font-mono">{{ ours.length }} chars</span>
        </div>
        <textarea
          :value="ours"
          readonly
          class="w-full h-44 p-3 font-mono text-xs rounded-lg border border-emerald-300 dark:border-emerald-800/60 bg-emerald-50/30 dark:bg-emerald-950/20 text-slate-800 dark:text-slate-200 focus:outline-none resize-none"
        ></textarea>
      </div>

      <!-- Server Incoming Edit -->
      <div class="space-y-1.5">
        <div class="flex items-center justify-between text-xs font-semibold text-sky-800 dark:text-sky-400">
          <span>Server Latest (Committed by Peer)</span>
          <span class="text-[11px] font-mono">{{ theirs.length }} chars</span>
        </div>
        <textarea
          :value="theirs"
          readonly
          class="w-full h-44 p-3 font-mono text-xs rounded-lg border border-sky-300 dark:border-sky-800/60 bg-sky-50/30 dark:bg-sky-950/20 text-slate-800 dark:text-slate-200 focus:outline-none resize-none"
        ></textarea>
      </div>
    </div>

    <!-- Final Merged Resolution Editor -->
    <div class="space-y-2 pt-2 border-t border-amber-200 dark:border-amber-900/60">
      <div class="flex items-center justify-between">
        <label class="text-xs font-bold text-slate-800 dark:text-slate-200">
          Resolved Document Text (Will be committed as new revision):
        </label>
        <span class="text-xs text-slate-400 font-mono">{{ mergedText.length }} chars</span>
      </div>
      <textarea
        v-model="mergedText"
        class="w-full h-56 p-3 font-mono text-xs rounded-lg border border-slate-300 dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-forest-500"
      ></textarea>
    </div>

    <!-- Resolution Submit Actions -->
    <div class="flex items-center justify-end gap-3 pt-2">
      <button
        @click="$emit('cancel')"
        class="px-4 py-2 text-xs font-semibold rounded-lg border border-slate-300 dark:border-slate-700 text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 transition"
      >
        Discard & Abort
      </button>
      <button
        @click="handleResolve"
        class="px-5 py-2 text-xs font-bold rounded-lg bg-forest-600 hover:bg-forest-500 text-white shadow-sm transition"
      >
        Submit Resolved Revision
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const props = withDefaults(
  defineProps<{
    ours?: string
    theirs?: string
    base?: string
  }>(),
  {
    ours: '== Overview ==\nQuantum computing is an emerging field with exponential speedup.\n[[Category:Science]]',
    theirs: '== Overview ==\nQuantum computing harnesses qubits to solve complex optimization problems.\n[[Category:Physics]]',
    base: '== Overview ==\nQuantum computing is a computational field.\n[[Category:Science]]',
  }
)

const emit = defineEmits<{
  (e: 'resolve', merged: string): void
  (e: 'cancel'): void
}>()

const mergedText = ref(props.ours)

const useOurs = () => {
  mergedText.value = props.ours
}

const useTheirs = () => {
  mergedText.value = props.theirs
}

const useCombined = () => {
  // Simple automatic section concatenation
  mergedText.value = `${props.ours}\n\n<!-- Merged Content from Peer Revision -->\n${props.theirs}`
}

const handleResolve = () => {
  emit('resolve', mergedText.value)
}
</script>
