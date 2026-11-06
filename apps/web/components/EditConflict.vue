<!-- apps/web/components/EditConflict.vue -->
<template>
  <div class="border border-amber-400 dark:border-amber-800 bg-amber-50 dark:bg-amber-950/20 p-5 space-y-5">
    <!-- Conflict Banner -->
    <div class="border-b border-amber-300 dark:border-amber-900 pb-3">
      <h2 class="font-wiki text-xl text-amber-900 dark:text-amber-200">
        Edit conflict
      </h2>
      <p class="text-sm text-amber-800 dark:text-amber-300 leading-relaxed mt-1">
        Someone else saved a revision while you were editing.
        Review both versions below and resolve the overlapping text before publishing.
      </p>
    </div>

    <!-- Quick Action Resolution Presets -->
    <div class="flex flex-wrap items-center gap-2 text-sm">
      <span class="font-semibold">Resolve:</span>
      <button
        @click="useOurs"
        class="px-3 py-1 bg-white dark:bg-slate-800 border border-wiki-border dark:border-slate-700 hover:bg-slate-100"
      >
        Keep mine
      </button>
      <button
        @click="useTheirs"
        class="px-3 py-1 bg-white dark:bg-slate-800 border border-wiki-border dark:border-slate-700 hover:bg-slate-100"
      >
        Keep server version
      </button>
      <button
        @click="useCombined"
        class="px-3 py-1 bg-wiki-link hover:bg-sky-700 text-white"
      >
        Merge both
      </button>
    </div>

    <!-- Side-by-Side Comparison of Versions -->
    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <div class="space-y-1.5">
        <div class="flex items-center justify-between text-sm font-semibold text-emerald-800 dark:text-emerald-400">
          <span>Your text</span>
          <span class="text-xs font-mono">{{ ours.length }} chars</span>
        </div>
        <textarea
          :value="ours"
          readonly
          class="w-full h-44 p-3 font-mono text-[13px] border border-emerald-400 dark:border-emerald-800 bg-emerald-50/50 dark:bg-emerald-950/20 text-slate-800 dark:text-slate-200 focus:outline-none resize-none"
        ></textarea>
      </div>

      <div class="space-y-1.5">
        <div class="flex items-center justify-between text-sm font-semibold text-sky-800 dark:text-sky-400">
          <span>Server version</span>
          <span class="text-xs font-mono">{{ theirs.length }} chars</span>
        </div>
        <textarea
          :value="theirs"
          readonly
          class="w-full h-44 p-3 font-mono text-[13px] border border-sky-400 dark:border-sky-800 bg-sky-50/50 dark:bg-sky-950/20 text-slate-800 dark:text-slate-200 focus:outline-none resize-none"
        ></textarea>
      </div>
    </div>

    <!-- Final Merged Resolution Editor -->
    <div class="space-y-2 pt-2 border-t border-amber-300 dark:border-amber-900">
      <div class="flex items-center justify-between">
        <label class="text-sm font-bold text-slate-800 dark:text-slate-200">
          Merged text (will be saved as a new revision)
        </label>
        <span class="text-xs text-slate-400 font-mono">{{ mergedText.length }} chars</span>
      </div>
      <textarea
        v-model="mergedText"
        class="w-full h-56 p-3 font-mono text-[13px] border border-wiki-border dark:border-slate-700 bg-white dark:bg-slate-900 text-slate-900 dark:text-slate-100 focus:outline-none focus:border-wiki-link"
      ></textarea>
    </div>

    <!-- Resolution Submit Actions -->
    <div class="flex items-center justify-end gap-2 pt-2 text-sm">
      <button
        @click="$emit('cancel')"
        class="px-4 py-1.5 border border-wiki-border dark:border-slate-700 hover:bg-slate-100 dark:hover:bg-slate-800"
      >
        Cancel
      </button>
      <button
        @click="handleResolve"
        class="px-5 py-1.5 bg-wiki-link hover:bg-sky-700 text-white font-semibold"
      >
        Save merged revision
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
