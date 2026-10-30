<!-- apps/web/pages/tree/index.vue -->
<template>
  <div class="space-y-6 max-w-6xl mx-auto">
    <div class="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl p-6 shadow-sm">
      <!-- Title & Tab Bar -->
      <div class="flex flex-wrap items-center justify-between border-b border-slate-200 dark:border-slate-800 pb-4 mb-6 gap-4">
        <div>
          <h1 class="text-2xl font-bold text-slate-900 dark:text-slate-100 flex items-center gap-2">
            <span>🌲</span> Topic Tree Knowledge Taxonomy
          </h1>
          <p class="text-xs text-slate-500">Directed Acyclic Graph (DAG) visualizing prerequisite relationships and topic hierarchies</p>
        </div>

        <!-- View Switcher -->
        <div class="flex items-center gap-1 bg-slate-100 dark:bg-slate-800 p-1 rounded-lg text-xs font-semibold">
          <button
            @click="activeTab = 'graph'"
            :class="activeTab === 'graph' ? 'bg-white dark:bg-slate-700 text-forest-700 dark:text-forest-400 shadow-xs' : 'text-slate-600 dark:text-slate-400'"
            class="px-3 py-1.5 rounded-md transition flex items-center gap-1.5"
          >
            <span>🕸️</span> Network Graph
          </button>
          <button
            @click="activeTab = 'tree'"
            :class="activeTab === 'tree' ? 'bg-white dark:bg-slate-700 text-forest-700 dark:text-forest-400 shadow-xs' : 'text-slate-600 dark:text-slate-400'"
            class="px-3 py-1.5 rounded-md transition flex items-center gap-1.5"
          >
            <span>🌳</span> Hierarchy Tree
          </button>
          <button
            @click="activeTab = 'directory'"
            :class="activeTab === 'directory' ? 'bg-white dark:bg-slate-700 text-forest-700 dark:text-forest-400 shadow-xs' : 'text-slate-600 dark:text-slate-400'"
            class="px-3 py-1.5 rounded-md transition flex items-center gap-1.5"
          >
            <span>📋</span> Topic Directory
          </button>
        </div>
      </div>

      <!-- Tab 1: Force-Directed Knowledge Graph -->
      <div v-if="activeTab === 'graph'">
        <div v-if="loading" class="h-[600px] flex items-center justify-center bg-slate-900 rounded-xl text-slate-400 text-xs">
          <span class="inline-block animate-spin mr-2">⌛</span> Loading DAG taxonomy...
        </div>
        <ForceDirectedGraph
          v-else
          :nodes="graphNodes"
          :edges="graphEdges"
          @select-node="handleGraphSelect"
        />
      </div>

      <!-- Tab 2: D3 Hierarchy Tree Viewer -->
      <div v-else-if="activeTab === 'tree'">
        <div v-if="loading" class="h-[600px] flex items-center justify-center bg-slate-900 rounded-xl text-slate-400 text-xs">
          <span class="inline-block animate-spin mr-2">⌛</span> Loading hierarchical taxonomy...
        </div>
        <D3TreeViewer v-else />
      </div>

      <!-- Tab 3: Topic Directory Grid -->
      <div v-else class="grid grid-cols-1 md:grid-cols-3 gap-6">
        <!-- Interactive Hierarchy Column -->
        <div class="md:col-span-2 border border-slate-200 dark:border-slate-800 rounded-lg p-4 bg-slate-50/50 dark:bg-slate-950/40 space-y-3">
          <div v-if="loading" class="py-12 text-center text-slate-400 text-xs">
            <span class="inline-block animate-spin text-xl mb-1">⌛</span>
            <p>Loading topic taxonomy...</p>
          </div>

          <div
            v-else
            v-for="node in displayNodes"
            :key="node.id"
            class="p-3.5 rounded-lg border border-slate-200 dark:border-slate-800 bg-white dark:bg-slate-900 shadow-xs cursor-pointer hover:border-forest-500 transition"
            :class="{ 'border-forest-500 ring-1 ring-forest-500': selectedNode?.id === node.id }"
            @click="selectedNode = node"
          >
            <div class="flex items-center justify-between mb-1">
              <span class="font-bold text-sm text-slate-800 dark:text-slate-200 flex items-center gap-2">
                <span>{{ node.icon }}</span> {{ node.title }}
              </span>
              <span class="text-xs font-mono text-slate-400">Difficulty {{ node.difficulty }}</span>
            </div>
            <p class="text-xs text-slate-500 line-clamp-2 leading-relaxed">{{ node.description }}</p>
            <div v-if="node.children.length > 0" class="mt-2.5 flex flex-wrap gap-1.5">
              <span v-for="sub in node.children" :key="sub" class="text-[11px] px-2 py-0.5 rounded bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-400 font-mono">
                ↳ {{ sub }}
              </span>
            </div>
          </div>
        </div>

        <!-- Selected Node Inspector Panel -->
        <div class="border border-slate-200 dark:border-slate-800 rounded-lg p-5 bg-white dark:bg-slate-900 space-y-4">
          <h3 class="font-bold text-sm text-slate-900 dark:text-slate-100 border-b border-slate-100 dark:border-slate-800 pb-2.5 flex items-center gap-1.5">
            <span>🔎</span> Node Details
          </h3>
          <div v-if="selectedNode" class="space-y-4">
            <div>
              <span class="text-[10px] font-bold text-slate-400 uppercase tracking-wider">Selected Topic</span>
              <h4 class="font-bold text-base text-forest-700 dark:text-forest-400 mt-0.5">{{ selectedNode.title }}</h4>
              <p class="text-xs text-slate-600 dark:text-slate-400 mt-1 leading-relaxed">{{ selectedNode.description }}</p>
            </div>

            <div>
              <span class="text-[10px] font-bold text-slate-400 uppercase tracking-wider">Bound Articles</span>
              <ul class="mt-1.5 space-y-1.5">
                <li v-for="art in selectedNode.articles" :key="art" class="text-xs">
                  <NuxtLink :to="`/w/${encodeURIComponent(art.replace(/ /g, '_'))}`" class="text-forest-700 dark:text-forest-400 hover:underline flex items-center gap-1.5 font-medium">
                    📄 {{ art.replace(/_/g, ' ') }} →
                  </NuxtLink>
                </li>
              </ul>
            </div>

            <div class="pt-2 border-t border-slate-100 dark:border-slate-800">
              <NuxtLink
                :to="`/w/${encodeURIComponent(selectedNode.title.replace(/ /g, '_'))}`"
                class="block w-full py-2 px-3 text-center rounded-lg bg-forest-600 hover:bg-forest-500 text-white font-bold text-xs transition"
              >
                Read Article
              </NuxtLink>
            </div>
          </div>

          <div v-else class="text-xs text-slate-400 py-12 text-center">
            Click on any topic node on the left to inspect its prerequisites and bound encyclopedia articles.
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useApi } from '../../composables/useApi'

interface ApiTopicNode {
  id: string
  name: string
  description: string
  difficulty: number
  child_ids?: string[]
  parent_ids?: string[]
}

interface ApiTopicEdge {
  from: string
  to: string
}

interface ApiGraphExport {
  nodes: ApiTopicNode[]
  edges: ApiTopicEdge[]
}

interface DirectoryNodeItem {
  id: string
  title: string
  icon: string
  difficulty: number
  description: string
  children: string[]
  articles: string[]
}

const activeTab = ref<'graph' | 'tree' | 'directory'>('graph')
const api = useApi()
const router = useRouter()

const loading = ref(true)
const rawNodes = ref<ApiTopicNode[]>([])
const rawEdges = ref<ApiTopicEdge[]>([])

const icons: Record<string, string> = {
  cs: '💻',
  algo: '⚙️',
  ds: '🧱',
  graph: '🕸️',
  diff: '⚖️',
  ai: '🤖',
  math: '📐',
  physics: '⚛️',
}

const getIcon = (id: string) => icons[id.toLowerCase()] || '📄'

const loadTopicData = async () => {
  loading.value = true
  try {
    const res = await api.get<ApiGraphExport>('/topics/export')
    if (res && res.nodes) {
      rawNodes.value = res.nodes
      rawEdges.value = res.edges || []
    }
  } catch (err) {
    console.error('Failed to load topic taxonomy:', err)
  } finally {
    loading.value = false
    if (displayNodes.value.length > 0 && !selectedNode.value) {
      selectedNode.value = displayNodes.value[0]
    }
  }
}

const graphNodes = computed(() => {
  if (rawNodes.value.length === 0) return undefined
  return rawNodes.value.map(n => ({
    id: n.id,
    name: n.name,
    group: n.difficulty || 1,
    difficulty: n.difficulty || 1,
    description: n.description,
  }))
})

const graphEdges = computed(() => {
  if (rawEdges.value.length === 0) return undefined
  return rawEdges.value.map(e => ({
    source: e.from,
    target: e.to,
  }))
})

const displayNodes = computed<DirectoryNodeItem[]>(() => {
  if (rawNodes.value.length === 0) {
    return [
      {
        id: 'cs',
        title: 'Computer Science',
        icon: '💻',
        difficulty: 1,
        description: 'Study of algorithmic computation, hardware/software systems, and information architectures.',
        children: ['Algorithms', 'Data Structures'],
        articles: ['Computer Science'],
      },
    ]
  }

  return rawNodes.value.map(n => ({
    id: n.id,
    title: n.name,
    icon: getIcon(n.id),
    difficulty: n.difficulty || 1,
    description: n.description || 'Topic in the Treewiki knowledge taxonomy.',
    children: n.child_ids || [],
    articles: [n.name.replace(/ /g, '_')],
  }))
})

const selectedNode = ref<DirectoryNodeItem | null>(null)

const handleGraphSelect = (node: any) => {
  const match = displayNodes.value.find(n => n.id === node.id)
  if (match) {
    selectedNode.value = match
  }
}

onMounted(() => {
  loadTopicData()
})
</script>
