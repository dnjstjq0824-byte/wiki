<!-- apps/web/components/ForceDirectedGraph.vue -->
<template>
  <div class="relative w-full h-[600px] bg-slate-900 rounded-xl overflow-hidden border border-slate-800 select-none">
    <!-- Toolbar Controls -->
    <div class="absolute top-4 left-4 z-10 flex items-center gap-2 bg-slate-800/80 backdrop-blur px-3 py-1.5 rounded-lg border border-slate-700 text-xs text-slate-200 shadow-md">
      <button @click="zoomIn" class="px-2 py-1 rounded bg-slate-700 hover:bg-slate-600 transition" title="Zoom In">+</button>
      <button @click="zoomOut" class="px-2 py-1 rounded bg-slate-700 hover:bg-slate-600 transition" title="Zoom Out">-</button>
      <button @click="resetTransform" class="px-2 py-1 rounded bg-slate-700 hover:bg-slate-600 transition" title="Reset View">Reset</button>
      <span class="text-slate-500">|</span>
      <button @click="toggleSimulation" class="px-2 py-1 rounded bg-forest-700 hover:bg-forest-600 text-white font-medium transition">
        {{ isSimulating ? 'Pause Physics' : 'Resume Physics' }}
      </button>
      <span class="text-slate-400 font-mono">Zoom: {{ Math.round(scale * 100) }}%</span>
    </div>

    <!-- Node Search / Filter -->
    <div class="absolute top-4 right-4 z-10">
      <input
        v-model="searchQuery"
        type="text"
        placeholder="Filter nodes..."
        class="px-3 py-1.5 text-xs rounded-lg bg-slate-800/80 backdrop-blur border border-slate-700 text-slate-200 placeholder-slate-500 focus:outline-none focus:ring-1 focus:ring-forest-500"
      />
    </div>

    <!-- Interactive SVG Canvas -->
    <svg
      ref="svgRef"
      class="w-full h-full cursor-grab active:cursor-grabbing"
      @mousedown="startPan"
      @mousemove="handleMouseMove"
      @mouseup="endPanOrDrag"
      @mouseleave="endPanOrDrag"
      @wheel.prevent="handleWheel"
    >
      <g :transform="`translate(${panX}, ${panY}) scale(${scale})`">
        <!-- Edges -->
        <g class="edges">
          <line
            v-for="(edge, idx) in simEdges"
            :key="idx"
            :x1="edge.sourceNode?.x ?? 0"
            :y1="edge.sourceNode?.y ?? 0"
            :x2="edge.targetNode?.x ?? 0"
            :y2="edge.targetNode?.y ?? 0"
            stroke="#334155"
            stroke-width="1.5"
            stroke-opacity="0.6"
          />
        </g>

        <!-- Nodes -->
        <g class="nodes">
          <g
            v-for="node in simNodes"
            :key="node.id"
            :transform="`translate(${node.x}, ${node.y})`"
            class="cursor-pointer transition-opacity"
            :class="{ 'opacity-30': isDimmed(node), 'opacity-100': !isDimmed(node) }"
            @mousedown.stop="startDragNode(node, $event)"
            @click.stop="selectNode(node)"
          >
            <!-- Outer Glow / Selection Ring -->
            <circle
              v-if="selectedNode?.id === node.id"
              r="24"
              fill="none"
              stroke="#10B981"
              stroke-width="2.5"
              stroke-dasharray="4,2"
              class="animate-spin-slow"
            />

            <!-- Main Node Circle -->
            <circle
              :r="getNodeRadius(node)"
              :fill="getNodeColor(node)"
              stroke="#0F172A"
              stroke-width="2"
              class="hover:brightness-125 transition"
            />

            <!-- Node Label -->
            <text
              y="22"
              text-anchor="middle"
              fill="#E2E8F0"
              font-size="11"
              font-weight="600"
              class="pointer-events-none drop-shadow-sm select-none"
            >
              {{ node.name }}
            </text>
          </g>
        </g>
      </g>
    </svg>

    <!-- Selected Node Detail Overlay Card -->
    <div
      v-if="selectedNode"
      class="absolute bottom-4 left-4 z-10 w-80 bg-slate-800/95 backdrop-blur p-4 rounded-xl border border-slate-700 text-slate-100 shadow-xl"
    >
      <div class="flex items-center justify-between pb-2 border-b border-slate-700 mb-2">
        <h3 class="font-bold text-sm text-forest-400 flex items-center gap-1.5">
          <span>🌿</span> {{ selectedNode.name }}
        </h3>
        <button @click="selectedNode = null" class="text-slate-400 hover:text-slate-200 text-xs">✕</button>
      </div>
      <p class="text-xs text-slate-300 leading-relaxed mb-3">
        {{ selectedNode.description || 'Topic node connected within the Treewiki knowledge graph.' }}
      </p>
      <div class="flex items-center justify-between text-[11px] text-slate-400 pt-2 border-t border-slate-700/60">
        <span>Connections: {{ getDegree(selectedNode.id) }} edges</span>
        <NuxtLink :to="`/w/${selectedNode.name.replace(/ /g, '_')}`" class="text-forest-400 hover:underline font-semibold">
          Open Article →
        </NuxtLink>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed, watch } from 'vue'

export interface GraphNode {
  id: string
  name: string
  description?: string
  group?: number
  difficulty?: number
}

export interface GraphEdge {
  source: string
  target: string
  weight?: number
}

interface SimNode extends GraphNode {
  x: number
  y: number
  vx: number
  vy: number
  isFixed?: boolean
}

interface SimEdge {
  source: string
  target: string
  sourceNode?: SimNode
  targetNode?: SimNode
}

const props = withDefaults(
  defineProps<{
    nodes?: GraphNode[]
    edges?: GraphEdge[]
  }>(),
  {
    nodes: () => [
      { id: 'cs', name: 'Computer Science', group: 1, difficulty: 1, description: 'The study of computation, information, and automation.' },
      { id: 'algo', name: 'Algorithms', group: 1, difficulty: 2, description: 'Finite sequences of rigorous instructions.' },
      { id: 'ds', name: 'Data Structures', group: 1, difficulty: 2, description: 'Data organization, management, and storage formats.' },
      { id: 'quantum', name: 'Quantum Computing', group: 2, difficulty: 4, description: 'Computation harnessing quantum mechanics (qubits, entanglement).' },
      { id: 'physics', name: 'Quantum Physics', group: 2, difficulty: 3, description: 'Fundamental physics branch describing nature at atomic scales.' },
      { id: 'crypto', name: 'Cryptography', group: 1, difficulty: 3, description: 'Practice and study of secure communication techniques.' },
      { id: 'math', name: 'Mathematics', group: 3, difficulty: 1, description: 'The science of structure, order, and relations.' },
      { id: 'linalg', name: 'Linear Algebra', group: 3, difficulty: 2, description: 'Branch concerning linear equations and matrices.' },
    ],
    edges: () => [
      { source: 'cs', target: 'algo' },
      { source: 'cs', target: 'ds' },
      { source: 'cs', target: 'crypto' },
      { source: 'algo', target: 'quantum' },
      { source: 'physics', target: 'quantum' },
      { source: 'math', target: 'cs' },
      { source: 'math', target: 'linalg' },
      { source: 'linalg', target: 'quantum' },
    ],
  }
)

const emit = defineEmits<{
  (e: 'select-node', node: GraphNode): void
}>()

const svgRef = ref<SVGSVGElement | null>(null)
const simNodes = ref<SimNode[]>([])
const simEdges = ref<SimEdge[]>([])
const selectedNode = ref<SimNode | null>(null)
const searchQuery = ref('')
const isSimulating = ref(true)

// Pan & Zoom
const panX = ref(400)
const panY = ref(300)
const scale = ref(1.0)
const isPanning = ref(false)
const dragStart = ref({ x: 0, y: 0 })
const draggedNode = ref<SimNode | null>(null)

let animId: number | null = null

const initSimulation = () => {
  const width = 800
  const height = 600

  // Initialize node positions in circle
  const n = props.nodes.length
  simNodes.value = props.nodes.map((node, i) => {
    const angle = (i / n) * 2 * Math.PI
    const radius = 180 + Math.random() * 40
    return {
      ...node,
      x: width / 2 + radius * Math.cos(angle),
      y: height / 2 + radius * Math.sin(angle),
      vx: 0,
      vy: 0,
    }
  })

  // Build edge references
  const nodeMap = new Map<string, SimNode>()
  simNodes.value.forEach((node) => nodeMap.set(node.id, node))

  simEdges.value = props.edges.map((edge) => ({
    source: edge.source,
    target: edge.target,
    sourceNode: nodeMap.get(edge.source),
    targetNode: nodeMap.get(edge.target),
  }))
}

const tickPhysics = () => {
  if (!isSimulating.value) return

  const width = 800
  const height = 600
  const nodes = simNodes.value
  const edges = simEdges.value

  const repulsionK = 3500
  const springK = 0.04
  const springLength = 120
  const centerGravity = 0.015
  const damping = 0.85

  // 1. Repulsion between all node pairs
  for (let i = 0; i < nodes.length; i++) {
    for (let j = i + 1; j < nodes.length; j++) {
      const n1 = nodes[i]
      const n2 = nodes[j]
      const dx = n2.x - n1.x
      const dy = n2.y - n1.y
      const distSq = dx * dx + dy * dy + 100 // prevent div by zero
      const dist = Math.sqrt(distSq)
      const force = repulsionK / distSq

      const fx = (dx / dist) * force
      const fy = (dy / dist) * force

      if (!n1.isFixed) {
        n1.vx -= fx
        n1.vy -= fy
      }
      if (!n2.isFixed) {
        n2.vx += fx
        n2.vy += fy
      }
    }
  }

  // 2. Spring attraction along edges
  for (const edge of edges) {
    if (!edge.sourceNode || !edge.targetNode) continue
    const n1 = edge.sourceNode
    const n2 = edge.targetNode
    const dx = n2.x - n1.x
    const dy = n2.y - n1.y
    const dist = Math.sqrt(dx * dx + dy * dy) || 1
    const displacement = dist - springLength
    const force = displacement * springK

    const fx = (dx / dist) * force
    const fy = (dy / dist) * force

    if (!n1.isFixed) {
      n1.vx += fx
      n1.vy += fy
    }
    if (!n2.isFixed) {
      n2.vx -= fx
      n2.vy -= fy
    }
  }

  // 3. Center gravity and position integration
  for (const node of nodes) {
    if (node.isFixed) continue

    // Gravity to center
    node.vx += (width / 2 - node.x) * centerGravity
    node.vy += (height / 2 - node.y) * centerGravity

    // Apply damping
    node.vx *= damping
    node.vy *= damping

    // Integrate
    node.x += node.vx
    node.y += node.vy
  }
}

const runLoop = () => {
  tickPhysics()
  animId = requestAnimationFrame(runLoop)
}

// Node Drag & Pan Events
const startPan = (e: MouseEvent) => {
  isPanning.value = true
  dragStart.value = { x: e.clientX - panX.value, y: e.clientY - panY.value }
}

const startDragNode = (node: SimNode, e: MouseEvent) => {
  draggedNode.value = node
  node.isFixed = true
}

const handleMouseMove = (e: MouseEvent) => {
  if (draggedNode.value && svgRef.value) {
    const rect = svgRef.value.getBoundingClientRect()
    draggedNode.value.x = (e.clientX - rect.left - panX.value) / scale.value
    draggedNode.value.y = (e.clientY - rect.top - panY.value) / scale.value
    draggedNode.value.vx = 0
    draggedNode.value.vy = 0
    return
  }

  if (isPanning.value) {
    panX.value = e.clientX - dragStart.value.x
    panY.value = e.clientY - dragStart.value.y
  }
}

const endPanOrDrag = () => {
  if (draggedNode.value) {
    draggedNode.value.isFixed = false
    draggedNode.value = null
  }
  isPanning.value = false
}

const handleWheel = (e: WheelEvent) => {
  const zoomFactor = e.deltaY < 0 ? 1.1 : 0.9
  const newScale = Math.min(Math.max(scale.value * zoomFactor, 0.4), 3.0)
  scale.value = newScale
}

const zoomIn = () => {
  scale.value = Math.min(scale.value * 1.2, 3.0)
}

const zoomOut = () => {
  scale.value = Math.max(scale.value * 0.8, 0.4)
}

const resetTransform = () => {
  panX.value = 100
  panY.value = 50
  scale.value = 1.0
}

const toggleSimulation = () => {
  isSimulating.value = !isSimulating.value
}

const selectNode = (node: SimNode) => {
  selectedNode.value = node
  emit('select-node', node)
}

const getNodeRadius = (node: SimNode) => {
  const deg = getDegree(node.id)
  return Math.min(14 + deg * 2, 24)
}

const getNodeColor = (node: SimNode) => {
  switch (node.group) {
    case 1:
      return '#0EA5E9' // Blue (Computer Science)
    case 2:
      return '#10B981' // Green (Quantum/Physics)
    case 3:
      return '#8B5CF6' // Purple (Math)
    default:
      return '#0F766E'
  }
}

const getDegree = (nodeId: string) => {
  return simEdges.value.filter((e) => e.source === nodeId || e.target === nodeId).length
}

const isDimmed = (node: SimNode) => {
  if (!searchQuery.value.trim()) return false
  return !node.name.toLowerCase().includes(searchQuery.value.toLowerCase())
}

watch(
  () => props.nodes,
  () => {
    initSimulation()
  },
  { deep: true }
)

onMounted(() => {
  initSimulation()
  resetTransform()
  runLoop()
})

onUnmounted(() => {
  if (animId !== null) {
    cancelAnimationFrame(animId)
  }
})
</script>
