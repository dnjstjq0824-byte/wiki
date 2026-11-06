<!-- apps/web/pages/admin/index.vue -->
<template>
  <div>
    <h1 class="font-wiki text-3xl text-slate-900 dark:text-slate-100 border-b border-wiki-border dark:border-slate-700 pb-1">
      Administration
    </h1>
    <p class="text-sm text-slate-500 mt-1">Moderation queue, abuse filters, blocks, takedowns, and oversight ledger.</p>

    <!-- Access / Error States -->
    <div v-if="!canAdmin" class="py-10 text-center">
      <h2 class="font-wiki text-xl text-slate-900 dark:text-slate-100">Restricted area</h2>
      <p class="text-sm text-slate-500 mt-1">Sign in with a SYSOP, OVERSIGHT, BUREAUCRAT, or SYSTEM_ADMIN account to manage moderation.</p>
      <p class="text-xs text-slate-400 font-mono mt-1">Current role: {{ authStore.userRole }}</p>
    </div>

    <div v-else-if="loading" class="py-12 text-center text-sm text-slate-500">
      <p>Loading moderation data…</p>
    </div>

    <div v-else-if="loadError" class="py-10 text-center space-y-3">
      <p class="text-sm text-slate-500">{{ loadError }}</p>
      <button @click="loadAll" class="px-4 py-1 bg-wiki-link hover:bg-sky-700 text-white text-sm">Retry</button>
    </div>

    <div v-else class="mt-4">
      <div class="flex justify-end">
        <button
          @click="loadAll"
          class="px-3 py-1 text-sm border border-wiki-border dark:border-slate-700 bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700"
        >
          Refresh
        </button>
      </div>

      <!-- Stat summary -->
      <p class="text-sm text-slate-600 dark:text-slate-400 mt-2">
        {{ stats?.pages ?? '—' }} articles ·
        {{ stats?.topics ?? '—' }} topics ·
        {{ blocks.length }} blocks ·
        {{ pendingTakedowns }} pending takedowns ·
        {{ filters.length }} filters
      </p>

      <!-- Tabs -->
      <nav aria-label="Moderation sections" class="flex items-end gap-0 text-sm border-b border-wiki-border dark:border-slate-700 mt-3 -mb-px">
        <button
          v-for="t in tabs"
          :key="t.id"
          @click="activeTab = t.id"
          class="px-3 py-1 border border-b-0 -mb-px mr-1"
          :class="activeTab === t.id ? 'bg-white dark:bg-slate-900 border-wiki-border dark:border-slate-700 font-semibold' : 'bg-slate-100 dark:bg-slate-800 border-slate-200 dark:border-slate-700 text-wiki-link dark:text-sky-400'"
        >
          {{ t.label }}
        </button>
      </nav>

      <div v-if="actionError" class="mt-3 px-3 py-2 text-sm text-red-700 dark:text-red-300 bg-red-50 dark:bg-red-950/40 border border-red-300 dark:border-red-900">
        {{ actionError }}
      </div>
      <div v-if="actionOk" class="mt-3 px-3 py-2 text-sm text-emerald-700 dark:text-emerald-300 bg-emerald-50 dark:bg-emerald-950/40 border border-emerald-300 dark:border-emerald-900">
        {{ actionOk }}
      </div>

      <!-- Blocks Tab -->
      <section v-if="activeTab === 'blocks'" class="mt-3 space-y-3 text-sm">
        <form @submit.prevent="addBlock" class="flex flex-wrap gap-2 border border-wiki-border dark:border-slate-700 bg-wiki-pagebg dark:bg-slate-800/40 p-3">
          <input v-model="blockForm.target" required placeholder="Username or IP to block" class="flex-1 min-w-[10rem] px-2 py-1 border border-wiki-border dark:border-slate-700 bg-white dark:bg-slate-950 focus:outline-none focus:border-wiki-link" />
          <input v-model="blockForm.reason" required placeholder="Reason (public log)" class="flex-[2] min-w-[12rem] px-2 py-1 border border-wiki-border dark:border-slate-700 bg-white dark:bg-slate-950 focus:outline-none focus:border-wiki-link" />
          <button type="submit" :disabled="acting" class="px-4 py-1 bg-red-700 hover:bg-red-600 text-white disabled:opacity-50">Block</button>
        </form>
        <p v-if="blocks.length === 0" class="text-slate-500">No active blocks.</p>
        <ul v-else class="divide-y divide-slate-200 dark:divide-slate-800">
          <li v-for="b in blocks" :key="b.id" class="py-2 flex flex-wrap items-baseline gap-x-3 gap-y-1">
            <span class="font-mono font-semibold">{{ b.target_user_id || b.target_ip }}</span>
            <span v-if="b.target_ip && b.target_user_id" class="text-slate-400 font-mono text-[13px]">{{ b.target_ip }}</span>
            <span class="text-slate-600 dark:text-slate-300">{{ b.reason }}</span>
            <span class="text-slate-400 text-[13px]">by {{ b.blocked_by }} · {{ formatTime(b.created_at) }}</span>
          </li>
        </ul>
      </section>

      <!-- Filters Tab -->
      <section v-if="activeTab === 'filters'" class="mt-3 space-y-3 text-sm">
        <form @submit.prevent="addFilter" class="flex flex-wrap gap-2 border border-wiki-border dark:border-slate-700 bg-wiki-pagebg dark:bg-slate-800/40 p-3">
          <input v-model="filterForm.name" required placeholder="Filter name" class="flex-1 min-w-[8rem] px-2 py-1 border border-wiki-border dark:border-slate-700 bg-white dark:bg-slate-950 focus:outline-none focus:border-wiki-link" />
          <input v-model="filterForm.pattern" required placeholder="Regex, e.g. spam-site\.com" class="flex-[2] min-w-[10rem] px-2 py-1 border border-wiki-border dark:border-slate-700 bg-white dark:bg-slate-950 font-mono focus:outline-none focus:border-wiki-link" />
          <select v-model="filterForm.action" class="px-2 py-1 border border-wiki-border dark:border-slate-700 bg-white dark:bg-slate-950">
            <option>TAG</option>
            <option>WARN</option>
            <option>DISALLOW</option>
            <option>BLOCK</option>
          </select>
          <button type="submit" :disabled="acting" class="px-4 py-1 bg-wiki-link hover:bg-sky-700 text-white disabled:opacity-50">Add filter</button>
        </form>
        <p v-if="filters.length === 0" class="text-slate-500">No abuse filters configured.</p>
        <ul v-else class="divide-y divide-slate-200 dark:divide-slate-800">
          <li v-for="f in filters" :key="f.id" class="py-2 flex flex-wrap items-baseline gap-x-3 gap-y-1">
            <span class="font-semibold">{{ f.name }}</span>
            <span class="font-mono text-[13px] text-slate-500">{{ f.pattern }}</span>
            <span class="font-mono text-xs px-1 border" :class="f.action === 'BLOCK' || f.action === 'DISALLOW' ? 'text-red-700 border-red-400' : 'text-amber-700 border-amber-400'">{{ f.action }}</span>
          </li>
        </ul>
      </section>

      <!-- Takedowns Tab -->
      <section v-if="activeTab === 'takedowns'" class="mt-3 text-sm">
        <p v-if="takedowns.length === 0" class="text-slate-500">No takedown requests.</p>
        <ul v-else class="divide-y divide-slate-200 dark:divide-slate-800">
          <li v-for="t in takedowns" :key="t.id" class="py-2 flex flex-wrap items-center gap-x-3 gap-y-1">
            <NuxtLink :to="`/w/${encodeURIComponent(t.page_title.replace(/ /g, '_'))}`" class="font-semibold text-wiki-link dark:text-sky-400 hover:underline">{{ t.page_title }}</NuxtLink>
            <span class="text-slate-600 dark:text-slate-300">{{ t.reason }}</span>
            <span class="text-slate-400 text-[13px]">filed by {{ t.requester }}</span>
            <span class="font-mono text-xs px-1 border border-slate-300 dark:border-slate-700 text-slate-600 dark:text-slate-300">{{ t.status }}</span>
            <select
              :value="t.status"
              @change="updateTakedown(t.id, ($event.target as HTMLSelectElement).value)"
              class="px-2 py-0.5 border border-wiki-border dark:border-slate-700 bg-white dark:bg-slate-950 text-[13px]"
            >
              <option>PENDING</option>
              <option>APPROVED</option>
              <option>REJECTED</option>
              <option>ESCALATED</option>
            </select>
          </li>
        </ul>
      </section>

      <!-- Oversight Tab -->
      <section v-if="activeTab === 'oversight'" class="mt-3 space-y-3 text-sm">
        <form @submit.prevent="addOversight" class="flex flex-wrap gap-2 border border-wiki-border dark:border-slate-700 bg-wiki-pagebg dark:bg-slate-800/40 p-3">
          <input v-model="oversightForm.action" required placeholder="Action, e.g. suppress" class="flex-1 min-w-[8rem] px-2 py-1 border border-wiki-border dark:border-slate-700 bg-white dark:bg-slate-950 focus:outline-none focus:border-wiki-link" />
          <input v-model="oversightForm.target" required placeholder="Target, e.g. revision id" class="flex-1 min-w-[8rem] px-2 py-1 border border-wiki-border dark:border-slate-700 bg-white dark:bg-slate-950 focus:outline-none focus:border-wiki-link" />
          <input v-model="oversightForm.details" placeholder="Details (private log)" class="flex-[2] min-w-[10rem] px-2 py-1 border border-wiki-border dark:border-slate-700 bg-white dark:bg-slate-950 focus:outline-none focus:border-wiki-link" />
          <button type="submit" :disabled="acting" class="px-4 py-1 bg-slate-700 hover:bg-slate-600 text-white disabled:opacity-50">Log entry</button>
        </form>
        <p v-if="oversight.length === 0" class="text-slate-500">Oversight ledger is empty.</p>
        <ul v-else class="divide-y divide-slate-200 dark:divide-slate-800">
          <li v-for="o in oversight" :key="o.id" class="py-2 flex flex-wrap items-baseline gap-x-3 gap-y-1">
            <span class="font-semibold">{{ o.action }}</span>
            <span class="font-mono text-[13px] text-slate-500">{{ o.target }}</span>
            <span class="text-slate-600 dark:text-slate-300">{{ o.details }}</span>
            <span class="text-slate-400 text-[13px]">by {{ o.actor_id }} · {{ formatTime(o.created_at) }}</span>
          </li>
        </ul>
      </section>

      <!-- Recent Tab -->
      <section v-if="activeTab === 'recent'" class="mt-3 text-sm">
        <p v-if="recent.length === 0" class="text-slate-500">No recent activity in this window.</p>
        <ul v-else class="divide-y divide-slate-200 dark:divide-slate-800 font-mono text-[13px]">
          <li v-for="(r, idx) in recent.slice(0, 30)" :key="idx" class="py-1.5 text-slate-600 dark:text-slate-400 break-all">
            {{ r }}
          </li>
        </ul>
        <p class="mt-3">
          <NuxtLink to="/special/recent-changes" class="text-wiki-link dark:text-sky-400 hover:underline">Full Recent Changes feed</NuxtLink>
        </p>
      </section>
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
