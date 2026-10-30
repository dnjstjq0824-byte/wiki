<!-- apps/web/pages/talk/[...slug].vue -->
<template>
  <div class="space-y-6 max-w-5xl mx-auto">
    <!-- Header Banner -->
    <div class="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl p-6 shadow-sm space-y-4">
      <div class="flex flex-wrap items-center justify-between gap-4 border-b border-slate-100 dark:border-slate-800 pb-4">
        <div>
          <div class="flex items-center gap-2">
            <span class="text-2xl">💬</span>
            <h1 class="text-2xl font-bold text-slate-900 dark:text-slate-100">
              Talk: <span class="text-forest-700 dark:text-forest-400">{{ displayTitle }}</span>
            </h1>
          </div>
          <p class="text-xs text-slate-500 mt-1">
            Collaborative consensus, RFC proposals, and editorial discussion for this article.
          </p>
        </div>

        <div class="flex items-center gap-2 text-xs font-semibold">
          <NuxtLink
            :to="`/w/${slug}`"
            class="px-3.5 py-1.5 rounded-lg border border-slate-200 dark:border-slate-700 text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-800 transition"
          >
            ← Back to Article
          </NuxtLink>
          <button
            @click="showNewThreadModal = true"
            class="px-4 py-1.5 rounded-lg bg-forest-600 hover:bg-forest-500 text-white shadow-xs transition flex items-center gap-1.5 font-bold"
          >
            <span>+</span> Start New Topic
          </button>
        </div>
      </div>

      <!-- Quick Guidance Notice -->
      <div class="p-3 bg-sky-50 dark:bg-sky-950/40 border border-sky-100 dark:border-sky-900 rounded-lg text-xs text-sky-800 dark:text-sky-300 flex items-start gap-2">
        <span class="text-base leading-none">ℹ️</span>
        <div>
          <strong>Talk Page Guidelines:</strong> Sign your comments, maintain a neutral point of view (NPOV), and be civil. Proposals affecting structure should be initiated via RFC with clear quorum goals.
        </div>
      </div>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="py-16 text-center text-slate-400 space-y-2">
      <div class="inline-block animate-spin text-3xl">⌛</div>
      <p class="text-xs">Loading discussion threads...</p>
    </div>

    <!-- Empty Discussion State -->
    <div
      v-else-if="threads.length === 0"
      class="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl p-12 text-center shadow-sm space-y-4"
    >
      <div class="inline-flex p-4 rounded-full bg-slate-100 dark:bg-slate-800 text-slate-400 text-3xl">
        💭
      </div>
      <div class="space-y-1">
        <h2 class="text-lg font-bold text-slate-800 dark:text-slate-200">No discussion topics yet</h2>
        <p class="text-xs text-slate-500 max-w-sm mx-auto">
          Be the first to open a discussion, propose an RFC improvement, or ask a question regarding "{{ displayTitle }}".
        </p>
      </div>
      <div>
        <button
          @click="showNewThreadModal = true"
          class="px-5 py-2.5 rounded-lg bg-forest-600 hover:bg-forest-500 text-white font-bold text-xs shadow-xs transition"
        >
          Start the First Topic
        </button>
      </div>
    </div>

    <!-- Threads List -->
    <div v-else class="space-y-4">
      <div
        v-for="thread in threads"
        :key="thread.id"
        class="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl overflow-hidden shadow-sm transition"
      >
        <!-- Thread Header Bar -->
        <div
          class="p-4 md:p-5 flex flex-wrap items-center justify-between gap-3 border-b border-slate-100 dark:border-slate-800 bg-slate-50/50 dark:bg-slate-800/40 cursor-pointer hover:bg-slate-100/60 dark:hover:bg-slate-800/70 transition"
          @click="toggleThread(thread.id)"
        >
          <div class="flex items-center gap-3">
            <span class="text-sm text-slate-400">
              {{ expandedThreads[thread.id] ? '▼' : '▶' }}
            </span>
            <div>
              <div class="flex items-center gap-2">
                <span v-if="thread.is_pinned" class="text-xs" title="Pinned Topic">📌</span>
                <h3 class="text-base font-bold text-slate-900 dark:text-slate-100">
                  {{ thread.title }}
                </h3>
              </div>
              <p class="text-[11px] text-slate-400 mt-0.5">
                Opened by <strong class="text-slate-600 dark:text-slate-300">{{ thread.author_id }}</strong>
                • {{ formatTime(thread.created_at) }}
              </p>
            </div>
          </div>

          <!-- Status Badge -->
          <div class="flex items-center gap-2">
            <span
              class="text-[10px] font-mono font-bold px-2 py-0.5 rounded border uppercase"
              :class="getStatusBadgeClass(thread.status)"
            >
              {{ thread.status }}
            </span>
            <span class="text-xs text-slate-400 font-mono">
              {{ (threadComments[thread.id] || []).length }} replies
            </span>
          </div>
        </div>

        <!-- Thread Body (When Expanded) -->
        <div v-if="expandedThreads[thread.id]" class="p-5 space-y-6">
          <!-- Optional RFC Section -->
          <div
            v-if="threadRFC[thread.id]"
            class="p-4 rounded-xl border border-forest-200 dark:border-forest-900/60 bg-forest-50/50 dark:bg-forest-950/20 space-y-3"
          >
            <div class="flex items-center justify-between">
              <span class="text-xs font-bold text-forest-700 dark:text-forest-400 uppercase tracking-wider flex items-center gap-1.5">
                <span>🗳️</span> Active Community RFC
              </span>
              <span class="text-[11px] font-mono text-slate-500">
                Quorum: {{ threadRFC[thread.id].quorum }} votes
              </span>
            </div>

            <p class="text-sm font-semibold text-slate-800 dark:text-slate-200">
              "{{ threadRFC[thread.id].question }}"
            </p>

            <!-- Vote Tally Visual Bar -->
            <div class="space-y-1.5 pt-1">
              <div class="h-2.5 w-full bg-slate-200 dark:bg-slate-700 rounded-full overflow-hidden flex">
                <div
                  :style="{ width: `${getRFCVotePct(threadRFC[thread.id], 'SUPPORT')}%` }"
                  class="bg-emerald-500"
                  title="Support"
                ></div>
                <div
                  :style="{ width: `${getRFCVotePct(threadRFC[thread.id], 'NEUTRAL')}%` }"
                  class="bg-amber-400"
                  title="Neutral"
                ></div>
                <div
                  :style="{ width: `${getRFCVotePct(threadRFC[thread.id], 'OPPOSE')}%` }"
                  class="bg-rose-500"
                  title="Oppose"
                ></div>
              </div>

              <div class="flex items-center justify-between text-[11px] text-slate-500 font-mono">
                <span class="text-emerald-600 dark:text-emerald-400 font-bold">
                  Support: {{ getVoteCount(threadRFC[thread.id], 'SUPPORT') }}
                </span>
                <span class="text-amber-600 dark:text-amber-400 font-bold">
                  Neutral: {{ getVoteCount(threadRFC[thread.id], 'NEUTRAL') }}
                </span>
                <span class="text-rose-600 dark:text-rose-400 font-bold">
                  Oppose: {{ getVoteCount(threadRFC[thread.id], 'OPPOSE') }}
                </span>
              </div>
            </div>

            <!-- Cast Vote Action -->
            <div class="flex items-center gap-2 pt-2 border-t border-forest-100 dark:border-forest-900/40">
              <span class="text-xs font-semibold text-slate-600 dark:text-slate-400">Cast your vote:</span>
              <button
                @click="castRFCVote(threadRFC[thread.id].id, thread.id, 'SUPPORT')"
                class="px-3 py-1 rounded-md text-xs font-bold bg-emerald-100 dark:bg-emerald-950 text-emerald-700 dark:text-emerald-300 hover:bg-emerald-200 transition"
              >
                ✓ Support
              </button>
              <button
                @click="castRFCVote(threadRFC[thread.id].id, thread.id, 'NEUTRAL')"
                class="px-3 py-1 rounded-md text-xs font-bold bg-amber-100 dark:bg-amber-950 text-amber-700 dark:text-amber-300 hover:bg-amber-200 transition"
              >
                ~ Neutral
              </button>
              <button
                @click="castRFCVote(threadRFC[thread.id].id, thread.id, 'OPPOSE')"
                class="px-3 py-1 rounded-md text-xs font-bold bg-rose-100 dark:bg-rose-950 text-rose-700 dark:text-rose-300 hover:bg-rose-200 transition"
              >
                ✕ Oppose
              </button>
            </div>
          </div>

          <!-- Comments Thread -->
          <div class="space-y-3">
            <div
              v-for="comment in (threadComments[thread.id] || [])"
              :key="comment.id"
              class="p-4 rounded-xl border border-slate-100 dark:border-slate-800 bg-slate-50/70 dark:bg-slate-800/50 space-y-1.5"
            >
              <div class="flex items-center justify-between text-xs">
                <span class="font-bold text-slate-800 dark:text-slate-200 flex items-center gap-1.5">
                  <span class="w-5 h-5 rounded-full bg-forest-600 text-white flex items-center justify-center text-[10px]">
                    {{ comment.author_id.charAt(0).toUpperCase() }}
                  </span>
                  {{ comment.author_id }}
                </span>
                <span class="text-[11px] text-slate-400 font-mono">
                  {{ formatTime(comment.created_at) }}
                </span>
              </div>
              <p class="text-xs text-slate-700 dark:text-slate-300 whitespace-pre-wrap leading-relaxed">
                {{ comment.content }}
              </p>
            </div>

            <div v-if="(threadComments[thread.id] || []).length === 0" class="text-xs text-slate-400 italic py-2">
              No replies yet on this thread.
            </div>
          </div>

          <!-- Reply Box -->
          <form @submit.prevent="submitReply(thread.id)" class="space-y-2 pt-2 border-t border-slate-100 dark:border-slate-800">
            <textarea
              v-model="replyText[thread.id]"
              rows="2"
              placeholder="Write a constructive response or counterpoint..."
              class="w-full p-3 text-xs rounded-lg bg-slate-50 dark:bg-slate-800 border border-slate-300 dark:border-slate-700 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-forest-500 transition resize-y"
              required
            ></textarea>
            <div class="flex justify-end">
              <button
                type="submit"
                :disabled="submittingReply[thread.id]"
                class="px-4 py-1.5 rounded-lg bg-forest-600 hover:bg-forest-500 text-white font-bold text-xs shadow-xs transition disabled:opacity-50"
              >
                {{ submittingReply[thread.id] ? 'Posting...' : 'Post Reply' }}
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>

    <!-- Start New Thread Modal -->
    <div
      v-if="showNewThreadModal"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-xs p-4"
    >
      <div class="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-2xl max-w-lg w-full p-6 shadow-2xl space-y-4">
        <div class="flex items-center justify-between border-b border-slate-100 dark:border-slate-800 pb-3">
          <h2 class="text-lg font-bold text-slate-900 dark:text-slate-100 flex items-center gap-2">
            <span>💬</span> New Discussion Topic
          </h2>
          <button @click="showNewThreadModal = false" class="text-slate-400 hover:text-slate-600">✕</button>
        </div>

        <form @submit.prevent="createThread" class="space-y-4 text-xs">
          <div>
            <label class="block font-semibold text-slate-700 dark:text-slate-300 mb-1">Topic Subject / Title</label>
            <input
              v-model="newThreadTitle"
              type="text"
              placeholder="e.g. Propose restructuring the Asymptotic Complexity section"
              class="w-full px-3 py-2 rounded-lg bg-slate-50 dark:bg-slate-800 border border-slate-300 dark:border-slate-700 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-forest-500"
              required
            />
          </div>

          <div>
            <label class="block font-semibold text-slate-700 dark:text-slate-300 mb-1">Initial Opening Statement</label>
            <textarea
              v-model="newThreadContent"
              rows="4"
              placeholder="Describe your suggestion, rationale, or query..."
              class="w-full p-3 rounded-lg bg-slate-50 dark:bg-slate-800 border border-slate-300 dark:border-slate-700 text-slate-900 dark:text-slate-100 focus:outline-none focus:ring-2 focus:ring-forest-500"
              required
            ></textarea>
          </div>

          <!-- Attach RFC Checkbox -->
          <div class="p-3 rounded-lg bg-slate-50 dark:bg-slate-800/50 border border-slate-200 dark:border-slate-700 space-y-3">
            <label class="flex items-center gap-2 cursor-pointer">
              <input type="checkbox" v-model="attachRFC" class="rounded text-forest-600" />
              <span class="font-bold text-slate-800 dark:text-slate-200">Formal Request For Comments (RFC)</span>
            </label>

            <div v-if="attachRFC" class="space-y-2 pt-1 border-t border-slate-200 dark:border-slate-700">
              <div>
                <label class="block text-slate-500 mb-0.5">RFC Proposal Question</label>
                <input
                  v-model="rfcQuestion"
                  type="text"
                  placeholder="Should we split this article into two independent pages?"
                  class="w-full px-3 py-1.5 rounded bg-white dark:bg-slate-900 border border-slate-300 dark:border-slate-700"
                  required
                />
              </div>
              <div>
                <label class="block text-slate-500 mb-0.5">Quorum Threshold (minimum votes)</label>
                <input
                  v-model.number="rfcQuorum"
                  type="number"
                  min="1"
                  max="100"
                  class="w-24 px-3 py-1.5 rounded bg-white dark:bg-slate-900 border border-slate-300 dark:border-slate-700 font-mono"
                />
              </div>
            </div>
          </div>

          <div class="flex items-center justify-end gap-2 pt-2">
            <button
              type="button"
              @click="showNewThreadModal = false"
              class="px-4 py-2 rounded-lg border border-slate-200 dark:border-slate-700 text-slate-600 dark:text-slate-300 font-semibold"
            >
              Cancel
            </button>
            <button
              type="submit"
              :disabled="submittingNewThread"
              class="px-5 py-2 rounded-lg bg-forest-600 hover:bg-forest-500 text-white font-bold transition disabled:opacity-50"
            >
              {{ submittingNewThread ? 'Creating...' : 'Publish Topic' }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useApi } from '../../composables/useApi'

interface TalkThread {
  id: string
  page_key: string
  title: string
  author_id: string
  status: string
  is_pinned: boolean
  is_locked: boolean
  created_at: string
}

interface TalkComment {
  id: string
  thread_id: string
  parent_id?: string
  author_id: string
  content: string
  created_at: string
}

interface RFCItem {
  id: string
  thread_id: string
  question: string
  quorum: number
  is_closed: boolean
  votes: Record<string, string> // username/author -> 'SUPPORT' | 'OPPOSE' | 'NEUTRAL'
}

const route = useRoute()
const api = useApi()

const slug = computed(() => {
  const p = route.params.slug
  return Array.isArray(p) ? p.join('/') : p || 'Treewiki'
})

const displayTitle = computed(() => slug.value.replace(/_/g, ' '))
const pageKey = computed(() => `0:${slug.value}`)

const loading = ref(true)
const threads = ref<TalkThread[]>([])
const expandedThreads = ref<Record<string, boolean>>({})
const threadComments = ref<Record<string, TalkComment[]>>({})
const threadRFC = ref<Record<string, RFCItem>>({})
const replyText = ref<Record<string, string>>({})
const submittingReply = ref<Record<string, boolean>>({})

// New Thread Form Modal
const showNewThreadModal = ref(false)
const newThreadTitle = ref('')
const newThreadContent = ref('')
const attachRFC = ref(false)
const rfcQuestion = ref('')
const rfcQuorum = ref(3)
const submittingNewThread = ref(false)

const loadThreads = async () => {
  loading.value = true
  try {
    const list = await api.get<TalkThread[]>('/talk/threads', { page_key: pageKey.value })
    threads.value = list || []

    // Automatically expand the first thread if available
    if (threads.value.length > 0) {
      const firstId = threads.value[0].id
      expandedThreads.value[firstId] = true
      await fetchThreadDetails(firstId)
    }
  } catch (err) {
    console.error('Failed to load talk threads:', err)
  } finally {
    loading.value = false
  }
}

const toggleThread = async (threadId: string) => {
  expandedThreads.value[threadId] = !expandedThreads.value[threadId]
  if (expandedThreads.value[threadId] && !threadComments.value[threadId]) {
    await fetchThreadDetails(threadId)
  }
}

const fetchThreadDetails = async (threadId: string) => {
  try {
    const comments = await api.get<TalkComment[]>(`/talk/threads/${threadId}/comments`)
    threadComments.value[threadId] = comments || []
  } catch (err) {
    console.error(`Failed to load comments for ${threadId}:`, err)
  }

  // Attempt to check if an RFC exists for this thread
  try {
    const rfc = await api.get<RFCItem>(`/talk/rfc/${threadId}`)
    if (rfc && rfc.question) {
      threadRFC.value[threadId] = rfc
    }
  } catch {
    // No RFC attached, normal thread
  }
}

const submitReply = async (threadId: string) => {
  const content = (replyText.value[threadId] || '').trim()
  if (!content) return

  submittingReply.value[threadId] = true
  try {
    const newComment = await api.post<TalkComment>(`/talk/threads/${threadId}/comments`, {
      content,
    })
    if (!threadComments.value[threadId]) {
      threadComments.value[threadId] = []
    }
    threadComments.value[threadId].push(newComment)
    replyText.value[threadId] = ''
  } catch (err) {
    console.error('Failed to submit comment:', err)
  } finally {
    submittingReply.value[threadId] = false
  }
}

const castRFCVote = async (rfcId: string, threadId: string, choice: string) => {
  try {
    await api.post(`/talk/rfc/${rfcId}/vote`, { choice })
    // Refresh RFC state
    const updated = await api.get<RFCItem>(`/talk/rfc/${rfcId}`)
    if (updated) {
      threadRFC.value[threadId] = updated
    }
  } catch (err) {
    console.error('Failed to cast vote:', err)
  }
}

const createThread = async () => {
  if (!newThreadTitle.value.trim() || !newThreadContent.value.trim()) return

  submittingNewThread.value = true
  try {
    const created = await api.post<TalkThread>('/talk/threads', {
      page_key: pageKey.value,
      title: newThreadTitle.value.trim(),
    })

    // Post opening comment
    await api.post(`/talk/threads/${created.id}/comments`, {
      content: newThreadContent.value.trim(),
    })

    // If RFC is selected
    if (attachRFC.value && rfcQuestion.value.trim()) {
      await api.post('/talk/rfc', {
        thread_id: created.id,
        question: rfcQuestion.value.trim(),
        quorum: rfcQuorum.value || 3,
      })
    }

    showNewThreadModal.value = false
    newThreadTitle.value = ''
    newThreadContent.value = ''
    attachRFC.value = false
    rfcQuestion.value = ''

    await loadThreads()
  } catch (err) {
    console.error('Failed to create thread:', err)
  } finally {
    submittingNewThread.value = false
  }
}

const getStatusBadgeClass = (status: string) => {
  switch (status?.toUpperCase()) {
    case 'OPEN':
      return 'bg-emerald-50 text-emerald-700 border-emerald-300 dark:bg-emerald-950 dark:text-emerald-400 dark:border-emerald-800'
    case 'RESOLVED':
      return 'bg-blue-50 text-blue-700 border-blue-300 dark:bg-blue-950 dark:text-blue-400 dark:border-blue-800'
    case 'FROZEN':
      return 'bg-purple-50 text-purple-700 border-purple-300 dark:bg-purple-950 dark:text-purple-400 dark:border-purple-800'
    case 'ARCHIVED':
      return 'bg-slate-100 text-slate-600 border-slate-300 dark:bg-slate-800 dark:text-slate-400 dark:border-slate-700'
    default:
      return 'bg-slate-100 text-slate-700 border-slate-200'
  }
}

const getVoteCount = (rfc: RFCItem, voteType: string) => {
  if (!rfc?.votes) return 0
  return Object.values(rfc.votes).filter(v => v === voteType).length
}

const getRFCVotePct = (rfc: RFCItem, voteType: string) => {
  if (!rfc?.votes) return 0
  const total = Object.values(rfc.votes).length
  if (total === 0) return 0
  const count = getVoteCount(rfc, voteType)
  return Math.round((count / total) * 100)
}

const formatTime = (ts: string) => {
  if (!ts) return ''
  try {
    const d = new Date(ts)
    if (isNaN(d.getTime())) return ts
    return d.toLocaleString()
  } catch {
    return ts
  }
}

onMounted(() => {
  loadThreads()
})
</script>

