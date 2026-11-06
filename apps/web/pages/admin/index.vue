<!-- apps/web/pages/admin/index.vue -->
<template>
  <div class="space-y-6 max-w-6xl mx-auto">
    <div class="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl p-6 shadow-sm">
      <div class="flex flex-wrap items-center justify-between gap-4 border-b border-slate-200 dark:border-slate-800 pb-4">
        <div>
          <h1 class="text-2xl font-bold text-slate-900 dark:text-slate-100 flex items-center gap-2">
            <span>🛡️</span> Administration Dashboard
          </h1>
          <p class="text-xs text-slate-500">Moderation queue, abuse filters, blocks, takedowns, and oversight ledger.</p>
        </div>
        <button
          @click="loadAll"
          class="px-3.5 py-1.5 rounded-lg border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-800 hover:bg-slate-100 text-xs font-semibold transition"
        >
          <span :class="{ 'animate-spin': loading }">🔄</span> Refresh
        </button>
      </div>

      <!-- Access / Error States -->
      <div v-if="!canAdmin" class="py-12 text-center space-y-3">
        <div class="text-3xl">🔒</div>
        <h2 class="text-lg font-bold text-slate-800 dark:text-slate-200">Restricted area</h2>
        <p class="text-xs text-slate-500 max-w-sm mx-auto">Sign in with a SYSOP, OVERSIGHT, BUREAUCRAT, or SYSTEM_ADMIN account to manage moderation.</p>
        <p class="text-[11px] text-slate-400 font-mono">Current role: {{ authStore.userRole }}</p>
      </div>

      <div v-else-if="loading" class="py-16 text-center text-slate-400 text-xs space-y-2">
        <div class="inline-block animate-spin text-3xl">⌛</div>
        <p>Loading moderation data...</p>
      </div>

      <div v-else-if="loadError" class="py-12 text-center space-y-3">
        <div class="text-3xl">📡</div>
        <p class="text-xs text-slate-500">{{ loadError }}</p>
        <button @click="loadAll" class="px-4 py-1.5 rounded-lg bg-forest-600 hover:bg-forest-500 text-white font-bold text-xs">Retry</button>
      </div>

      <div v-else class="pt-5 space-y-6">
        <!-- Stat Cards -->
        <div class="grid grid-cols-2 md:grid-cols-5 gap-3 text-center">
          <div class="p-3 rounded-lg bg-slate-50 dark:bg-slate-800/60 border border-slate-200 dark:border-slate-800">
            <div class="text-xl font-extrabold font-mono text-slate-900 dark:text-slate-100">{{ stats?.pages ?? '—' }}</div>
            <div class="text-[10px] text-slate-500 uppercase tracking-wide font-semibold">Articles</div>
          </div>
          <div class="p-3 rounded-lg bg-slate-50 dark:bg-slate-800/60 border border-slate-200 dark:border-slate-800">
            <div class="text-xl font-extrabold font-mono text-slate-900 dark:text-slate-100">{{ stats?.topics ?? '—' }}</div>
            <div class="text-[10px] text-slate-500 uppercase tracking-wide font-semibold">Topics</div>
          </div>
          <div class="p-3 rounded-lg bg-slate-50 dark:bg-slate-800/60 border border-slate-200 dark:border-slate-800">
            <div class="text-xl font-extrabold font-mono text-rose-600">{{ blocks.length }}</div>
            <div class="text-[10px] text-slate-500 uppercase tracking-wide font-semibold">Blocks</div>
          </div>
          <div class="p-3 rounded-lg bg-slate-50 dark:bg-slate-800/60 border border-slate-200 dark:border-slate-800">
            <div class="text-xl font-extrabold font-mono text-amber-600">{{ pendingTakedowns }}</div>
            <div class="text-[10px] text-slate-500 uppercase tracking-wide font-semibold">Pending Takedowns</div>
          </div>
          <div class="p-3 rounded-lg bg-slate-50 dark:bg-slate-800/60 border border-slate-200 dark:border-slate-800">
            <div class="text-xl font-extrabold font-mono text-forest-600">{{ filters.length }}</div>
            <div class="text-[10px] text-slate-500 uppercase tracking-wide font-semibold">Filters</div>
          </div>
        </div>

        <!-- Tabs -->
        <div class="flex items-center gap-1 bg-slate-100 dark:bg-slate-800 p-1 rounded-lg text-xs font-semibold w-fit">
          <button
            v-for="t in tabs"
            :key="t.id"
            @click="activeTab = t.id"
            :class="activeTab === t.id ? 'bg-white dark:bg-slate-700 text-forest-700 dark:text-forest-400 shadow-xs' : 'text-slate-600 dark:text-slate-400'"
            class="px-3 py-1.5 rounded-md transition"
          >
            {{ t.label }}
          </button>
        </div>

        <div v-if="actionError" class="p-3 rounded-lg bg-rose-50 dark:bg-rose-950/40 border border-rose-200 dark:border-rose-900 text-xs text-rose-700 dark:text-rose-300">
          {{ actionError }}
        </div>
        <div v-if="actionOk" class="p-3 rounded-lg bg-emerald-50 dark:bg-emerald-950/40 border border-emerald-200 dark:border-emerald-900 text-xs text-emerald-700 dark:text-emerald-300">
          {{ actionOk }}
        </div>

        <!-- Blocks Tab -->
        <section v-if="activeTab === 'blocks'" class="space-y-4">
          <form @submit.prevent="addBlock" class="grid grid-cols-1 md:grid-cols-4 gap-2 p-4 rounded-lg bg-slate-50 dark:bg-slate-800/50 border border-slate-200 dark:border-slate-700 text-xs">
            <input v-model="blockForm.target" required placeholder="Username or IP to block" class="px-3 py-2 rounded-lg bg-white dark:bg-slate-900 border border-slate-300 dark:border-slate-700 focus:outline-none focus:ring-2 focus:ring-forest-500" />
            <input v-model="blockForm.reason" required placeholder="Reason (public log)" class="px-3 py-2 rounded-lg bg-white dark:bg-slate-900 border border-slate-300 dark:border-slate-700 focus:outline-none focus:ring-2 focus:ring-forest-500 md:col-span-2" />
            <button type="submit" :disabled="acting" class="px-4 py-2 rounded-lg bg-rose-600 hover:bg-rose-500 text-white font-bold transition disabled:opacity-50">Issue Block</button>
          </form>
          <div class="border border-slate-200 dark:border-slate-800 rounded-xl overflow-hidden divide-y divide-slate-100 dark:divide-slate-800 text-xs">
            <div v-if="blocks.length === 0" class="p-6 text-center text-slate-400">No active blocks.</div>
            <div v-for="b in blocks" :key="b.id" class="p-3.5 flex flex-wrap items-center justify-between gap-3">
              <div>
                <span class="font-bold text-slate-800 dark:text-slate-200 font-mono">{{ b.target_user_id || b.target_ip }}</span>
                <span v-if="b.target_ip && b.target_user_id" class="text-slate-400 font-mono ml-2">{{ b.target_ip }}</span>
                <p class="text-slate-500 mt-0.5">{{ b.reason }} <span class="text-slate-400">· by {{ b.blocked_by }}</span></p>
              </div>
              <span class="text-[10px] font-mono text-slate-400">{{ formatTime(b.created_at) }}</span>
            </div>
          </div>
        </section>

        <!-- Filters Tab -->
        <section v-if="activeTab === 'filters'" class="space-y-4">
          <form @submit.prevent="addFilter" class="grid grid-cols-1 md:grid-cols-4 gap-2 p-4 rounded-lg bg-slate-50 dark:bg-slate-800/50 border border-slate-200 dark:border-slate-700 text-xs">
            <input v-model="filterForm.name" required placeholder="Filter name" class="px-3 py-2 rounded-lg bg-white dark:bg-slate-900 border border-slate-300 dark:border-slate-700 focus:outline-none focus:ring-2 focus:ring-forest-500" />
            <input v-model="filterForm.pattern" required placeholder="Regex pattern, e.g. spam-site\.com" class="px-3 py-2 rounded-lg bg-white dark:bg-slate-900 border border-slate-300 dark:border-slate-700 font-mono focus:outline-none focus:ring-2 focus:ring-forest-500" />
            <select v-model="filterForm.action" class="px-3 py-2 rounded-lg bg-white dark:bg-slate-900 border border-slate-300 dark:border-slate-700 focus:outline-none focus:ring-2 focus:ring-forest-500">
              <option>TAG</option>
              <option>WARN</option>
              <option>DISALLOW</option>
              <option>BLOCK</option>
            </select>
            <button type="submit" :disabled="acting" class="px-4 py-2 rounded-lg bg-forest-600 hover:bg-forest-500 text-white font-bold transition disabled:opacity-50">Add Filter</button>
          </form>
          <div class="border border-slate-200 dark:border-slate-800 rounded-xl overflow-hidden divide-y divide-slate-100 dark:divide-slate-800 text-xs">
            <div v-if="filters.length === 0" class="p-6 text-center text-slate-400">No abuse filters configured.</div>
            <div v-for="f in filters" :key="f.id" class="p-3.5 flex flex-wrap items-center justify-between gap-3">
              <div>
                <span class="font-bold text-slate-800 dark:text-slate-200">{{ f.name }}</span>
                <p class="font-mono text-slate-500 mt-0.5">{{ f.pattern }}</p>
              </div>
              <span class="text-[10px] font-mono font-bold px-2 py-0.5 rounded border" :class="f.action === 'BLOCK' || f.action === 'DISALLOW' ? 'bg-rose-50 text-rose-700 border-rose-300' : 'bg-amber-50 text-amber-700 border-amber-300'">{{ f.action }}</span>
            </div>
          </div>
        </section>

        <!-- Takedowns Tab -->
        <section v-if="activeTab === 'takedowns'" class="space-y-4">
          <div class="border border-slate-200 dark:border-slate-800 rounded-xl overflow-hidden divide-y divide-slate-100 dark:divide-slate-800 text-xs">
            <div v-if="takedowns.length === 0" class="p-6 text-center text-slate-400">No takedown requests.</div>
            <div v-for="t in takedowns" :key="t.id" class="p-3.5 flex flex-wrap items-center justify-between gap-3">
              <div>
                <NuxtLink :to="`/w/${encodeURIComponent(t.page_title.replace(/ /g, '_'))}`" class="font-bold text-forest-700 dark:text-forest-400 hover:underline">{{ t.page_title }}</NuxtLink>
                <p class="text-slate-500 mt-0.5">{{ t.reason }} <span class="text-slate-400">· filed by {{ t.requester }}</span></p>
              </div>
              <div class="flex items-center gap-2">
                <span class="text-[10px] font-mono font-bold px-2 py-0.5 rounded border bg-slate-100 dark:bg-slate-800 text-slate-600 dark:text-slate-300">{{ t.status }}</span>
                <select
                  :value="t.status"
                  @change="updateTakedown(t.id, ($event.target as HTMLSelectElement).value)"
                  class="px-2 py-1 rounded bg-white dark:bg-slate-900 border border-slate-300 dark:border-slate-700 text-xs"
                >
                  <option>PENDING</option>
                  <option>APPROVED</option>
                  <option>REJECTED</option>
                  <option>ESCALATED</option>
                </select>
              </div>
            </div>
          </div>
        </section>

        <!-- Oversight Tab -->
        <section v-if="activeTab === 'oversight'" class="space-y-4">
          <form @submit.prevent="addOversight" class="grid grid-cols-1 md:grid-cols-4 gap-2 p-4 rounded-lg bg-slate-50 dark:bg-slate-800/50 border border-slate-200 dark:border-slate-700 text-xs">
            <input v-model="oversightForm.action" required placeholder="Action, e.g. suppress" class="px-3 py-2 rounded-lg bg-white dark:bg-slate-900 border border-slate-300 dark:border-slate-700 focus:outline-none focus:ring-2 focus:ring-forest-500" />
            <input v-model="oversightForm.target" required placeholder="Target, e.g. revision id" class="px-3 py-2 rounded-lg bg-white dark:bg-slate-900 border border-slate-300 dark:border-slate-700 focus:outline-none focus:ring-2 focus:ring-forest-500" />
            <input v-model="oversightForm.details" placeholder="Details (private log)" class="px-3 py-2 rounded-lg bg-white dark:bg-slate-900 border border-slate-300 dark:border-slate-700 focus:outline-none focus:ring-2 focus:ring-forest-500" />
            <button type="submit" :disabled="acting" class="px-4 py-2 rounded-lg bg-slate-800 hover:bg-slate-700 text-white font-bold transition disabled:opacity-50">Log Entry</button>
          </form>
          <div class="border border-slate-200 dark:border-slate-800 rounded-xl overflow-hidden divide-y divide-slate-100 dark:divide-slate-800 text-xs">
            <div v-if="oversight.length === 0" class="p-6 text-center text-slate-400">Oversight ledger is empty.</div>
            <div v-for="o in oversight" :key="o.id" class="p-3.5 flex flex-wrap items-center justify-between gap-3">
              <div>
                <span class="font-bold text-slate-800 dark:text-slate-200">{{ o.action }}</span>
                <span class="font-mono text-slate-500 ml-2">{{ o.target }}</span>
                <p class="text-slate-500 mt-0.5">{{ o.details }} <span class="text-slate-400">· by {{ o.actor_id }}</span></p>
              </div>
              <span class="text-[10px] font-mono text-slate-400">{{ formatTime(o.created_at) }}</span>
            </div>
          </div>
        </section>

        <!-- Recent Tab -->
        <section v-if="activeTab === 'recent'" class="space-y-3">
          <div class="border border-slate-200 dark:border-slate-800 rounded-xl overflow-hidden divide-y divide-slate-100 dark:divide-slate-800 text-xs">
            <div v-if="recent.length === 0" class="p-6 text-center text-slate-400">No recent activity in this window.</div>
            <div v-for="(r, idx) in recent.slice(0, 30)" :key="idx" class="p-3 font-mono text-slate-600 dark:text-slate-400 break-all">
              {{ r }}
            </div>
          </div>
          <NuxtLink to="/special/recent-changes" class="text-xs font-semibold text-forest-600 dark:text-forest-400 hover:underline">Open full Recent Changes feed →</NuxtLink>
        </section>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useApi } from '../../composables/useApi'
import { useAuthStore } from '../../stores/auth'

interface BlockItem { id: string; target_user_id: string; target_ip: string; blocked_by: string; reason: string; created_at: string }
interface FilterItem { id: string; name: string; pattern: string; action: string; is_enabled: boolean }
interface TakedownItem { id: string; page_title: string; reason: string; requester: string; status: string; created_at: string }
interface OversightItem { id: string; actor_id: string; action: string; target: string; details: string; created_at: string }
interface GovDump { filters: FilterItem[]; blocks: BlockItem[]; oversight: OversightItem[]; takedowns: TakedownItem[]; recent: string[] }
interface PlatformStats { version: string; database: string; pages: number; topics: number; timestamp: string }

const api = useApi()
const authStore = useAuthStore()

const ADMIN_ROLES = ['SYSOP', 'OVERSIGHT', 'BUREAUCRAT', 'SYSTEM_ADMIN']
const canAdmin = computed(() => ADMIN_ROLES.includes(authStore.userRole))

const tabs = [
  { id: 'blocks', label: 'Blocks' },
  { id: 'filters', label: 'Abuse Filters' },
  { id: 'takedowns', label: 'Takedowns' },
  { id: 'oversight', label: 'Oversight' },
  { id: 'recent', label: 'Recent' },
]
const activeTab = ref('blocks')

const loading = ref(true)
const acting = ref(false)
const loadError = ref('')
const actionError = ref('')
const actionOk = ref('')

const stats = ref<PlatformStats | null>(null)
const blocks = ref<BlockItem[]>([])
const filters = ref<FilterItem[]>([])
const takedowns = ref<TakedownItem[]>([])
const oversight = ref<OversightItem[]>([])
const recent = ref<string[]>([])

const pendingTakedowns = computed(() => takedowns.value.filter(t => t.status === 'PENDING').length)

const blockForm = ref({ target: '', reason: '' })
const filterForm = ref({ name: '', pattern: '', action: 'TAG' })
const oversightForm = ref({ action: '', target: '', details: '' })

const formatTime = (ts: string) => {
  if (!ts) return ''
  try {
    const d = new Date(ts)
    return isNaN(d.getTime()) ? ts : d.toLocaleString()
  } catch {
    return ts
  }
}

const flashOk = (msg: string) => {
  actionError.value = ''
  actionOk.value = msg
  setTimeout(() => { actionOk.value = '' }, 4000)
}

const flashErr = (err: any, fallback: string) => {
  actionOk.value = ''
  actionError.value = err?.data?.error || err?.message || fallback
}

const loadAll = async () => {
  if (!canAdmin.value) {
    loading.value = false
    return
  }
  loading.value = true
  loadError.value = ''
  try {
    const [dump, st] = await Promise.all([
      api.get<GovDump>('/governance/dump'),
      api.get<PlatformStats>('/stats'),
    ])
    blocks.value = dump?.blocks || []
    filters.value = dump?.filters || []
    takedowns.value = dump?.takedowns || []
    oversight.value = dump?.oversight || []
    recent.value = (dump?.recent || []).slice().reverse()
    stats.value = st
  } catch (err: any) {
    console.error('Failed to load admin data:', err)
    loadError.value = err?.data?.error || 'Moderation data is unreachable. Your session may have expired — sign in again.'
  } finally {
    loading.value = false
  }
}

const looksLikeIP = (s: string) => /^[0-9a-fA-F.:]+$/.test(s) && (s.includes('.') || s.includes(':'))

const addBlock = async () => {
  acting.value = true
  try {
    const target = blockForm.value.target.trim()
    const payload: Record<string, string> = { reason: blockForm.value.reason.trim() }
    if (looksLikeIP(target)) payload.target_ip = target
    else payload.target_user_id = target
    await api.post('/governance/blocks', payload)
    blockForm.value = { target: '', reason: '' }
    flashOk('Block issued.')
    await loadAll()
  } catch (err) {
    flashErr(err, 'Failed to issue block.')
  } finally {
    acting.value = false
  }
}

const addFilter = async () => {
  acting.value = true
  try {
    await api.post('/governance/filters', {
      name: filterForm.value.name.trim(),
      pattern: filterForm.value.pattern,
      action: filterForm.value.action,
    })
    filterForm.value = { name: '', pattern: '', action: 'TAG' }
    flashOk('Abuse filter added.')
    await loadAll()
  } catch (err) {
    flashErr(err, 'Failed to add filter. Check the regex syntax.')
  } finally {
    acting.value = false
  }
}

const updateTakedown = async (id: string, status: string) => {
  try {
    await api.put(`/governance/takedowns/${id}`, { status })
    flashOk(`Takedown marked ${status}.`)
    await loadAll()
  } catch (err) {
    flashErr(err, 'Failed to update takedown.')
  }
}

const addOversight = async () => {
  acting.value = true
  try {
    await api.post('/governance/oversight', {
      actor_id: authStore.user?.username || '',
      action: oversightForm.value.action.trim(),
      target: oversightForm.value.target.trim(),
      details: oversightForm.value.details.trim(),
    })
    oversightForm.value = { action: '', target: '', details: '' }
    flashOk('Oversight entry logged.')
    await loadAll()
  } catch (err) {
    flashErr(err, 'Failed to log oversight entry. OVERSIGHT role required.')
  } finally {
    acting.value = false
  }
}

useSeoMeta({
  title: 'Administration — Treewiki',
  robots: 'noindex',
})

onMounted(() => {
  loadAll()
})
</script>
