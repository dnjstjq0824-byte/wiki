<template>
  <div class="space-y-8">
    <!-- Hero Banner -->
    <section class="p-6 md:p-8 rounded-2xl bg-gradient-to-br from-forest-800 via-forest-900 to-slate-900 text-white shadow-md relative overflow-hidden">
      <div class="relative z-10 max-w-2xl">
        <div class="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-forest-700/60 text-forest-200 text-xs font-semibold mb-4 border border-forest-500/30">
          <span>🌲</span> Structured Knowledge & Topic Trees
        </div>
        <h1 class="text-3xl md:text-4xl font-extrabold tracking-tight mb-3">
          Welcome to Treewiki
        </h1>
        <p class="text-forest-100 text-sm md:text-base leading-relaxed mb-6">
          A global collaborative knowledge platform designed to organize human understanding into interconnected topic trees, immutable revision ledgers, and neutral encyclopedic references.
        </p>
        <div class="flex flex-wrap gap-3">
          <NuxtLink to="/w/Computer_Science" class="px-4 py-2 rounded-lg bg-forest-500 hover:bg-forest-600 font-semibold text-sm transition shadow-sm">
            Explore Computer Science
          </NuxtLink>
          <NuxtLink to="/tree" class="px-4 py-2 rounded-lg bg-white/10 hover:bg-white/20 font-semibold text-sm transition backdrop-blur border border-white/20">
            Open Knowledge Graph
          </NuxtLink>
        </div>
      </div>
    </section>

    <!-- Main Grid: Featured Article & Topic Trees -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
      <!-- Featured Article Card -->
      <div class="md:col-span-2 p-6 rounded-xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 shadow-sm space-y-4">
        <div class="flex items-center justify-between border-b border-slate-100 dark:border-slate-800 pb-3">
          <span class="text-xs font-bold text-forest-700 dark:text-forest-400 uppercase tracking-wide">⭐ Today's Featured Knowledge</span>
          <span class="text-xs text-slate-400">CC BY-SA 4.0</span>
        </div>
        <h2 class="text-2xl font-bold text-slate-900 dark:text-slate-100">
          <NuxtLink to="/w/Quantum_Computing" class="hover:text-forest-600 transition">
            Quantum Computing
          </NuxtLink>
        </h2>
        <p class="text-sm text-slate-600 dark:text-slate-400 leading-relaxed">
          Quantum computing is a rapidly-emerging technology that harnesses the laws of quantum mechanics to solve problems too complex for classical computers. By utilizing superposition and entanglement, qubits process multidimensional computational spaces at exponential scale...
        </p>
        <div class="flex items-center justify-between pt-2">
          <div class="flex gap-2">
            <span class="px-2 py-0.5 rounded text-xs bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400">Physics</span>
            <span class="px-2 py-0.5 rounded text-xs bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400">Computation</span>
          </div>
          <NuxtLink to="/w/Quantum_Computing" class="text-xs font-semibold text-forest-600 dark:text-forest-400 hover:underline">
            Read full article →
          </NuxtLink>
        </div>
      </div>

      <!-- Quick Knowledge Taxonomy -->
      <div class="p-6 rounded-xl bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 shadow-sm space-y-4">
        <h3 class="font-bold text-base text-slate-900 dark:text-slate-100 border-b border-slate-100 dark:border-slate-800 pb-3 flex items-center gap-2">
          <span>🌿</span> Major Knowledge Trees
        </h3>
        <ul class="space-y-2.5 text-sm">
          <li v-for="tree in majorTrees" :key="tree.slug" class="flex items-center justify-between p-2 rounded-lg hover:bg-slate-50 dark:hover:bg-slate-800/60 transition">
            <NuxtLink :to="`/w/${tree.slug}`" class="font-medium text-slate-700 dark:text-slate-300 hover:text-forest-600 flex items-center gap-2">
              <span>{{ tree.icon }}</span> {{ tree.name }}
            </NuxtLink>
            <span class="text-xs text-slate-400 font-mono">{{ tree.count }} arts</span>
          </li>
        </ul>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const majorTrees = [
  { name: 'Computer Science', slug: 'Computer_Science', icon: '💻', count: '1,420' },
  { name: 'Mathematics', slug: 'Mathematics', icon: '📐', count: '980' },
  { name: 'Physics & Astronomy', slug: 'Physics', icon: '⚛️', count: '1,150' },
  { name: 'Biology & Medicine', slug: 'Biology', icon: '🧬', count: '890' },
  { name: 'Philosophy & Ethics', slug: 'Philosophy', icon: '🏛️', count: '640' }
]
</script>
