<template>
  <div class="space-y-6">
    <div class="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl p-6 shadow-sm">
      <div class="flex items-center justify-between border-b border-slate-200 dark:border-slate-800 pb-4 mb-6">
        <div>
          <h1 class="text-2xl font-bold text-slate-900 dark:text-slate-100 flex items-center gap-2">
            <span>🌲</span> Topic Tree Knowledge Taxonomy
          </h1>
          <p class="text-xs text-slate-500">Directed Acyclic Graph (DAG) visualizing prerequisite relationships and topic hierarchies</p>
        </div>
        <div class="text-xs font-semibold px-3 py-1 rounded-full bg-forest-50 dark:bg-slate-800 text-forest-700 dark:text-forest-400 border border-forest-200 dark:border-forest-800">
          Root: Human Knowledge
        </div>
      </div>

      <!-- Tree Nodes Grid -->
      <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
        <!-- Interactive Hierarchy Column -->
        <div class="md:col-span-2 border border-slate-200 dark:border-slate-800 rounded-lg p-4 bg-slate-50/50 dark:bg-slate-950/40 space-y-3">
          <div v-for="node in nodes" :key="node.id" class="p-3 rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-xs cursor-pointer hover:border-forest-500 transition" @click="selectNode(node)">
            <div class="flex items-center justify-between mb-1">
              <span class="font-bold text-sm text-slate-800 dark:text-slate-200 flex items-center gap-2">
                <span>{{ node.icon }}</span> {{ node.title }}
              </span>
              <span class="text-xs font-mono text-slate-400">Depth {{ node.depth }}</span>
            </div>
            <p class="text-xs text-slate-500 line-clamp-1">{{ node.description }}</p>
            <div class="mt-2 flex flex-wrap gap-1.5">
              <span v-for="sub in node.children" :key="sub" class="text-[11px] px-2 py-0.5 rounded bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400">
                ↳ {{ sub }}
              </span>
            </div>
          </div>
        </div>

        <!-- Selected Node Inspector Panel -->
        <div class="border border-slate-200 dark:border-slate-800 rounded-lg p-4 bg-white dark:bg-slate-900 space-y-4">
          <h3 class="font-bold text-sm text-slate-900 dark:text-slate-100 border-b pb-2 flex items-center gap-1.5">
            <span>🔎</span> Node Details
          </h3>
          <div v-if="selectedNode" class="space-y-3">
            <div>
              <span class="text-[11px] font-semibold text-slate-400 uppercase">Selected Topic</span>
              <h4 class="font-bold text-base text-forest-700 dark:text-forest-400">{{ selectedNode.title }}</h4>
              <p class="text-xs text-slate-600 dark:text-slate-400 mt-1">{{ selectedNode.description }}</p>
            </div>
            <div>
              <span class="text-[11px] font-semibold text-slate-400 uppercase">Bound Articles</span>
              <ul class="mt-1 space-y-1">
                <li v-for="art in selectedNode.articles" :key="art" class="text-xs">
                  <NuxtLink :to="`/w/${art.replace(/ /g, '_')}`" class="text-sky-600 dark:text-sky-400 hover:underline flex items-center gap-1">
                    📄 {{ art }}
                  </NuxtLink>
                </li>
              </ul>
            </div>
          </div>
          <div v-else class="text-xs text-slate-400 py-8 text-center">
            Click on any topic node on the left to inspect its children, prerequisites, and bound articles.
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

interface NodeItem {
  id: string
  title: string
  icon: string
  depth: number
  description: string
  children: string[]
  articles: string[]
}

const nodes: NodeItem[] = [
  {
    id: 'cs',
    title: 'Computer Science',
    icon: '💻',
    depth: 1,
    description: 'Study of algorithmic computation, hardware/software systems, and information architectures.',
    children: ['Algorithms & Data Structures', 'Quantum Computing', 'Distributed Systems', 'Cryptography'],
    articles: ['Computer Science', 'Quantum Computing', 'Turing Machine', 'Time Complexity']
  },
  {
    id: 'math',
    title: 'Mathematics',
    icon: '📐',
    depth: 1,
    description: 'Abstract science of numbers, quantity, geometry, spaces, and logical transformations.',
    children: ['Calculus & Analysis', 'Linear Algebra', 'Graph Theory', 'Abstract Algebra'],
    articles: ['Mathematics', 'Linear Algebra', 'Euler Characteristic']
  },
  {
    id: 'phys',
    title: 'Physics',
    icon: '⚛️',
    depth: 1,
    description: 'Empirical natural science exploring matter, energy, spacetime, and fundamental forces.',
    children: ['Quantum Mechanics', 'General Relativity', 'Thermodynamics', 'Electromagnetism'],
    articles: ['Physics', 'Quantum Mechanics', 'Special Relativity']
  }
]

const selectedNode = ref<NodeItem | null>(nodes[0])

const selectNode = (node: NodeItem) => {
  selectedNode.value = node
}
</script>
