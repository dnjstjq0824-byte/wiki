<!-- apps/web/components/D3TreeViewer.vue -->
<template>
  <div class="relative w-full h-[600px] bg-slate-900 rounded-xl overflow-hidden border border-slate-800 flex flex-col">
    <!-- Header Bar -->
    <div class="px-4 py-3 bg-slate-800/80 backdrop-blur border-b border-slate-700 flex items-center justify-between">
      <div class="flex items-center gap-2">
        <span class="text-forest-400 font-bold text-sm">🌳 Hierarchical Topic Tree Explorer</span>
        <span class="text-xs text-slate-400">| Total Nodes: {{ flattenedNodes.length }}</span>
      </div>
      <div class="flex items-center gap-2">
        <button
          @click="expandAll"
          class="px-2.5 py-1 text-xs rounded bg-slate-700 hover:bg-slate-600 text-slate-200 transition"
        >
          Expand All
        </button>
        <button
          @click="collapseAll"
          class="px-2.5 py-1 text-xs rounded bg-slate-700 hover:bg-slate-600 text-slate-200 transition"
        >
          Collapse All
        </button>
      </div>
    </div>

    <!-- Tree Tree Canvas / Hierarchy View -->
    <div class="flex-1 overflow-auto p-6 space-y-2">
      <div
        v-for="item in visibleNodes"
        :key="item.node.id"
        :style="{ paddingLeft: `${item.depth * 28}px` }"
        class="flex items-center gap-2 py-1.5 px-3 rounded-lg transition hover:bg-slate-800/60 cursor-pointer group"
        :class="{ 'bg-slate-800/90 border-l-4 border-forest-500': selectedId === item.node.id }"
        @click="selectNode(item.node)"
      >
        <!-- Toggle Expand/Collapse Icon -->
        <button
          v-if="item.node.children && item.node.children.length > 0"
          @click.stop="toggleCollapse(item.node.id)"
          class="w-5 h-5 flex items-center justify-center rounded text-slate-400 hover:text-white hover:bg-slate-700 text-xs transition"
        >
          {{ isCollapsed(item.node.id) ? '▶' : '▼' }}
        </button>
        <span v-else class="w-5 text-center text-slate-600 text-xs">•</span>

        <!-- Node Icon & Title -->
        <span class="text-sm font-semibold text-slate-200 group-hover:text-forest-400 transition flex items-center gap-2">
          <span>{{ item.node.icon || '📄' }}</span>
          {{ item.node.name }}
        </span>

        <!-- Difficulty Badge -->
        <span
          v-if="item.node.difficulty"
          class="text-[10px] px-2 py-0.5 rounded font-mono font-medium ml-2"
          :class="difficultyBadgeClass(item.node.difficulty)"
        >
          Level {{ item.node.difficulty }}
        </span>

        <!-- Child Count -->
        <span
          v-if="item.node.children && item.node.children.length > 0"
          class="text-[11px] text-slate-500 ml-auto"
        >
          {{ item.node.children.length }} subtopics
        </span>
      </div>
    </div>

    <!-- Bottom Selected Node Bar -->
    <div
      v-if="selectedNode"
      class="p-4 bg-slate-800 border-t border-slate-700 flex items-center justify-between text-xs text-slate-300"
    >
      <div>
        <span class="font-bold text-slate-100 text-sm mr-2">{{ selectedNode.name }}</span>
        <span class="text-slate-400">{{ selectedNode.description }}</span>
      </div>
      <NuxtLink
        :to="`/w/${selectedNode.name.replace(/ /g, '_')}`"
        class="px-3 py-1.5 rounded-lg bg-forest-600 hover:bg-forest-500 text-white font-semibold transition"
      >
        Read Full Article →
      </NuxtLink>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue'

export interface TreeNode {
  id: string
  name: string
  icon?: string
  description?: string
  difficulty?: number
  children?: TreeNode[]
}

const props = withDefaults(
  defineProps<{
    root?: TreeNode
  }>(),
  {
    // Empty fallback: real taxonomy arrives via the root prop from /topics/export.
    root: () => ({
      id: 'empty',
      name: 'No topics yet',
      icon: '🌱',
      description: 'Topics appear here once curators add them to the taxonomy.',
      difficulty: 1,
      children: [],
    }),
  }
)

const emit = defineEmits<{
  (e: 'select-node', node: TreeNode): void
}>()

const collapsedMap = ref<Record<string, boolean>>({})
const selectedId = ref<string>(props.root?.id || 'empty')

// Reselect the root when live taxonomy arrives asynchronously.
watch(() => props.root?.id, (id) => {
  if (id) selectedId.value = id
})

const isCollapsed = (id: string) => !!collapsedMap.value[id]

const toggleCollapse = (id: string) => {
  collapsedMap.value[id] = !collapsedMap.value[id]
}

const expandAll = () => {
  collapsedMap.value = {}
}

const collapseAll = () => {
  const map: Record<string, boolean> = {}
  const visit = (node: TreeNode) => {
    if (node.children && node.children.length > 0) {
      map[node.id] = true
      node.children.forEach(visit)
    }
  }
  visit(props.root)
  collapsedMap.value = map
}

interface FlattenedItem {
  node: TreeNode
  depth: number
}

const flatten = (node: TreeNode, depth: number, result: FlattenedItem[]) => {
  result.push({ node, depth })
  if (node.children && !isCollapsed(node.id)) {
    node.children.forEach((child) => flatten(child, depth + 1, result))
  }
}

const visibleNodes = computed(() => {
  const result: FlattenedItem[] = []
  flatten(props.root, 0, result)
  return result
})

const flattenedNodes = computed(() => {
  const result: TreeNode[] = []
  const visit = (n: TreeNode) => {
    result.push(n)
    n.children?.forEach(visit)
  }
  visit(props.root)
  return result
})

const selectedNode = computed(() => {
  return flattenedNodes.value.find((n) => n.id === selectedId.value) || null
})

const selectNode = (node: TreeNode) => {
  selectedId.value = node.id
  emit('select-node', node)
}

const difficultyBadgeClass = (lvl: number) => {
  switch (lvl) {
    case 1:
      return 'bg-emerald-950 text-emerald-400 border border-emerald-800'
    case 2:
      return 'bg-sky-950 text-sky-400 border border-sky-800'
    case 3:
      return 'bg-amber-950 text-amber-400 border border-amber-800'
    case 4:
    case 5:
      return 'bg-rose-950 text-rose-400 border border-rose-800'
    default:
      return 'bg-slate-800 text-slate-400'
  }
}
</script>
