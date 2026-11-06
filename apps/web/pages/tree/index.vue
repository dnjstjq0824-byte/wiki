<!-- apps/web/pages/tree/index.vue -->
<template>
  <div>
    <h1 class="text-3xl font-bold text-slate-900 dark:text-slate-100 border-b-2 border-namu-600 dark:border-namu-500 pb-1">
      Topic Tree Knowledge Taxonomy
    </h1>
    <p class="text-sm text-slate-500 mt-1">Prerequisite relationships and hierarchies across the knowledge graph (DAG).</p>

    <!-- View Switcher -->
    <nav aria-label="Taxonomy views" class="flex items-end gap-0 text-sm border-b border-wiki-border dark:border-slate-700 mt-4 -mb-px">
      <button
        @click="activeTab = 'graph'"
        class="px-3 py-1 border border-b-0 -mb-px mr-1"
        :class="activeTab === 'graph' ? 'bg-white dark:bg-slate-900 border-wiki-border dark:border-slate-700 font-semibold' : 'bg-slate-100 dark:bg-slate-800 border-slate-200 dark:border-slate-700 text-wiki-link dark:text-sky-400'"
      >
        Network Graph
      </button>
      <button
        @click="activeTab = 'tree'"
        class="px-3 py-1 border border-b-0 -mb-px mr-1"
        :class="activeTab === 'tree' ? 'bg-white dark:bg-slate-900 border-wiki-border dark:border-slate-700 font-semibold' : 'bg-slate-100 dark:bg-slate-800 border-slate-200 dark:border-slate-700 text-wiki-link dark:text-sky-400'"
      >
        Hierarchy Tree
      </button>
      <button
        @click="activeTab = 'directory'"
        class="px-3 py-1 border border-b-0 -mb-px mr-1"
        :class="activeTab === 'directory' ? 'bg-white dark:bg-slate-900 border-wiki-border dark:border-slate-700 font-semibold' : 'bg-slate-100 dark:bg-slate-800 border-slate-200 dark:border-slate-700 text-wiki-link dark:text-sky-400'"
      >
        Topic Directory
      </button>
    </nav>

    <!-- Tab 1: Force-Directed Knowledge Graph -->
    <div v-if="activeTab === 'graph'" class="mt-3">
      <div v-if="loading" class="h-[400px] flex items-center justify-center border border-wiki-border dark:border-slate-700 bg-slate-900 text-slate-400 text-sm">
        Loading taxonomy…
      </div>
      <ForceDirectedGraph
        v-else
        :nodes="graphNodes"
        :edges="graphEdges"
        @select-node="handleGraphSelect"
      />
    </div>

    <!-- Tab 2: D3 Hierarchy Tree Viewer -->
    <div v-else-if="activeTab === 'tree'" class="mt-3">
      <div v-if="loading" class="h-[400px] flex items-center justify-center border border-wiki-border dark:border-slate-700 bg-slate-900 text-slate-400 text-sm">
        Loading hierarchy…
      </div>
      <div v-else-if="loadError" class="py-10 text-center text-sm text-slate-500 space-y-3">
        <p>{{ loadError }}</p>
        <button @click="loadTopicData" class="px-4 py-1 bg-namu-600 hover:bg-namu-700 text-white">Retry</button>
      </div>
      <D3TreeViewer v-else :root="hierarchyRoot" />
    </div>

    <!-- Tab 3: Topic Directory -->
    <div v-else class="mt-3 grid grid-cols-1 md:grid-cols-3 gap-4">
      <div class="md:col-span-2">
        <div v-if="loading" class="py-10 text-center text-sm text-slate-500">
          <p>Loading topic taxonomy…</p>
        </div>

        <div v-else-if="loadError" class="py-10 text-center text-sm text-slate-500 space-y-3">
          <p>{{ loadError }}</p>
          <button @click="loadTopicData" class="px-4 py-1 bg-namu-600 hover:bg-namu-700 text-white">Retry</button>
        </div>

        <div v-else-if="displayNodes.length === 0" class="py-10 text-center text-sm text-slate-500">
          <p>No topic nodes yet. Topics appear here once curators add them.</p>
        </div>

        <table v-else class="w-full text-sm border-collapse">
          <thead>
            <tr class="border-b border-wiki-border dark:border-slate-700 text-left text-slate-500">
              <th class="py-1.5 pr-3 font-semibold">Topic</th>
              <th class="py-1.5 pr-3 font-semibold">Difficulty</th>
              <th class="py-1.5 font-semibold">Subtopics</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="node in displayNodes"
              :key="node.id"
              class="border-b border-slate-200 dark:border-slate-800 cursor-pointer hover:bg-slate-50 dark:hover:bg-slate-800/50"
              :class="{ 'bg-sky-50 dark:bg-sky-950/30': selectedNode?.id === node.id }"
              @click="selectedNode = node"
            >
              <td class="py-2 pr-3">
                <span class="font-semibold">{{ node.title }}</span>
                <span class="block text-[13px] text-slate-500">{{ node.description }}</span>
              </td>
              <td class="py-2 pr-3 font-mono text-[13px]">{{ node.difficulty }}</td>
              <td class="py-2 text-[13px] font-mono text-slate-500">{{ node.children.join(', ') || '—' }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Selected Node Inspector -->
      <aside aria-label="Topic details" class="border border-wiki-border dark:border-slate-700 h-fit">
        <h2 class="px-3 py-1.5 bg-slate-100 dark:bg-slate-800 border-b border-wiki-border dark:border-slate-700 font-bold text-sm">
          Topic details
        </h2>
        <div v-if="selectedNode" class="p-3 text-sm space-y-2">
          <p class="text-lg font-bold">{{ selectedNode.title }}</p>
          <p class="text-slate-600 dark:text-slate-400 text-[13px]">{{ selectedNode.description }}</p>
          <ul class="space-y-1">
            <li v-for="art in selectedNode.articles" :key="art">
              <NuxtLink :to="`/w/${encodeURIComponent(art.replace(/ /g, '_'))}`" class="text-wiki-link dark:text-sky-400 hover:underline">
                {{ art.replace(/_/g, ' ') }}
              </NuxtLink>
            </li>
          </ul>
          <p>
            <NuxtLink
              :to="`/w/${encodeURIComponent(selectedNode.title.replace(/ /g, '_'))}`"
              class="inline-block px-4 py-1 bg-namu-600 hover:bg-namu-700 text-white text-[13px]"
            >
              Read article
            </NuxtLink>
          </p>
        </div>
        <div v-else class="p-3 text-[13px] text-slate-500">
          Select a topic on the left to inspect its prerequisites and bound articles.
        </div>
      </aside>
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
  difficulty: number
  description: string
  children: string[]
  articles: string[]
}

const activeTab = ref<'graph' | 'tree' | 'directory'>('graph')
const api = useApi()
const router = useRouter()

const loading = ref(true)
const loadError = ref('')
const rawNodes = ref<ApiTopicNode[]>([])
const rawEdges = ref<ApiTopicEdge[]>([])

const loadTopicData = async () => {
  loading.value = true
  loadError.value = ''
  try {
    const res = await api.get<ApiGraphExport>('/topics/export')
    if (res && res.nodes) {
      rawNodes.value = res.nodes
      rawEdges.value = res.edges || []
    }
  } catch (err) {
    console.error('Failed to load topic taxonomy:', err)
    loadError.value = 'The topic service is unreachable. Check your connection and try again.'
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
  return rawNodes.value.map(n => ({
    id: n.id,
    title: n.name,
    difficulty: n.difficulty || 1,
    description: n.description || 'Topic in the Treewiki knowledge taxonomy.',
    children: n.child_ids || [],
    articles: [n.name.replace(/ /g, '_')],
  }))
})

const selectedNode = ref<DirectoryNodeItem | null>(null)

// Nest the DAG into a display tree (first-parent wins, cycles guarded).
const hierarchyRoot = computed(() => {
  if (rawNodes.value.length === 0) return undefined
  const byId = new Map(rawNodes.value.map(n => [n.id, n]))
  const childIds = new Set(rawEdges.value.map(e => e.to))
  const roots = rawNodes.value.filter(n => !childIds.has(n.id))
  const seeds = roots.length > 0 ? roots : rawNodes.value.slice(0, 1)
  const build = (id: string, seen: Set<string>): any => {
    const n = byId.get(id)
    if (!n) return null
    if (seen.has(id)) return { id: n.id, name: n.name, description: n.description, difficulty: n.difficulty || 1 }
    const next = new Set(seen)
    next.add(id)
    return {
      id: n.id,
      name: n.name,
      description: n.description || 'Topic in the Treewiki knowledge taxonomy.',
      difficulty: n.difficulty || 1,
      children: rawEdges.value
        .filter(e => e.from === id)
        .map(e => build(e.to, next))
        .filter(Boolean),
    }
  }
  const trees = seeds.map(s => build(s.id, new Set()))
  if (trees.length === 1) return trees[0]
  return {
    id: 'root',
    name: 'Knowledge Graph',
    description: 'Root of the Treewiki topic taxonomy.',
    difficulty: 1,
    children: trees,
  }
})

const handleGraphSelect = (node: any) => {
  const match = displayNodes.value.find(n => n.id === node.id)
  if (match) {
    selectedNode.value = match
  }
}

useSeoMeta({
  title: 'Topic Tree Knowledge Taxonomy — Treewiki',
  description: 'Explore the Treewiki topic DAG: prerequisite relationships and hierarchies across the knowledge graph.',
  ogTitle: 'Topic Tree Knowledge Taxonomy — Treewiki',
  ogDescription: 'A directed acyclic graph of human knowledge: prerequisites, hierarchies, and learning paths.',
  ogType: 'website',
})

onMounted(() => {
  loadTopicData()
})
</script>
