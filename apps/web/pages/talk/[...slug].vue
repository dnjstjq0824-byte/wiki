<!-- apps/web/pages/talk/[...slug].vue -->
<template>
  <div>
    <WikiTabs :slug="slug" active="talk" />

    <h1 class="text-3xl font-bold text-slate-900 dark:text-slate-100 border-b-2 border-namu-600 dark:border-namu-500 pb-1">
      Talk: {{ displayTitle }}
    </h1>
    <p class="text-sm text-slate-500 mt-1">Discussion, RFC proposals, and editorial consensus for this article.</p>

    <div class="mt-3 flex justify-end">
      <button
        @click="showNewThreadModal = true"
        class="px-3 py-1 text-sm border border-wiki-border dark:border-slate-700 bg-slate-100 dark:bg-slate-800 hover:bg-slate-200 dark:hover:bg-slate-700"
      >
        + New section
      </button>
    </div>

    <div class="mt-3 border border-wiki-border dark:border-slate-700 bg-sky-50 dark:bg-sky-950/30 px-3 py-2 text-[13px] text-slate-700 dark:text-slate-300">
      <strong>Talk page guidelines:</strong> sign your comments, stay neutral (NPOV), and be civil. Structural proposals should go through RFC with a clear quorum.
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="py-12 text-center text-sm text-slate-500">
      <p>Loading discussion threads…</p>
    </div>

    <!-- Empty Discussion State -->
    <div v-else-if="threads.length === 0" class="py-10 text-center">
      <h2 class="text-xl font-bold text-slate-900 dark:text-slate-100">No discussion topics yet</h2>
      <p class="text-sm text-slate-500 mt-1">
        Be the first to open a discussion or propose an RFC regarding “{{ displayTitle }}”.
      </p>
      <button
        @click="showNewThreadModal = true"
        class="mt-4 px-4 py-1.5 text-sm bg-namu-600 hover:bg-namu-700 text-white"
      >
        Start the first topic
      </button>
    </div>

    <!-- Threads List -->
    <div v-else class="mt-4">
      <section
        v-for="thread in threads"
        :key="thread.id"
        class="border-t border-wiki-border dark:border-slate-700 py-3"
      >
        <div class="flex flex-wrap items-baseline justify-between gap-2 cursor-pointer" @click="toggleThread(thread.id)">
          <h2 class="text-xl font-bold text-slate-900 dark:text-slate-100">
            <span class="text-slate-400 text-base mr-1">{{ expandedThreads[thread.id] ? '▾' : '▸' }}</span>
            {{ thread.title }}
          </h2>
          <p class="text-xs text-slate-500">
            <span
              class="font-mono text-[11px] px-1.5 py-0.5 border uppercase mr-2"
              :class="getStatusBadgeClass(thread.status)"
            >
              {{ thread.status }}
            </span>
            {{ (threadComments[thread.id] || []).length }} comments · {{ thread.author_id }} · {{ formatTime(thread.created_at) }}
          </p>
        </div>

        <div v-if="expandedThreads[thread.id]" class="mt-2 ml-1 space-y-4">
          <!-- Optional RFC Section -->
          <div
            v-if="threadRFC[thread.id]"
            class="border border-wiki-border dark:border-slate-700 bg-wiki-notice dark:bg-amber-950/20 px-4 py-3"
          >
            <div class="flex items-center justify-between text-sm">
              <span class="font-bold">Active Community RFC</span>
              <span class="text-xs text-slate-500 font-mono">Quorum: {{ threadRFC[thread.id].quorum }} votes</span>
            </div>
            <p class="text-sm mt-1">“{{ threadRFC[thread.id].question }}”</p>

            <div class="h-2 w-full bg-slate-200 dark:bg-slate-700 mt-2 flex">
              <div :style="{ width: `${getRFCVotePct(threadRFC[thread.id], 'SUPPORT')}%` }" class="bg-emerald-600" title="Support"></div>
              <div :style="{ width: `${getRFCVotePct(threadRFC[thread.id], 'NEUTRAL')}%` }" class="bg-amber-500" title="Neutral"></div>
              <div :style="{ width: `${getRFCVotePct(threadRFC[thread.id], 'OPPOSE')}%` }" class="bg-rose-600" title="Oppose"></div>
            </div>
            <div class="flex items-center justify-between text-xs text-slate-600 dark:text-slate-400 mt-1 font-mono">
              <span>Support: {{ getVoteCount(threadRFC[thread.id], 'SUPPORT') }}</span>
              <span>Neutral: {{ getVoteCount(threadRFC[thread.id], 'NEUTRAL') }}</span>
              <span>Oppose: {{ getVoteCount(threadRFC[thread.id], 'OPPOSE') }}</span>
            </div>

            <div class="flex items-center gap-2 mt-2 pt-2 border-t border-wiki-border dark:border-slate-700 text-xs">
              <span class="text-slate-600 dark:text-slate-400">Vote:</span>
              <button @click="castRFCVote(threadRFC[thread.id].id, thread.id, 'SUPPORT')" class="px-2.5 py-0.5 border border-wiki-border dark:border-slate-700 bg-white dark:bg-slate-900 text-emerald-700 dark:text-emerald-400 hover:bg-emerald-50 font-semibold">Support</button>
              <button @click="castRFCVote(threadRFC[thread.id].id, thread.id, 'NEUTRAL')" class="px-2.5 py-0.5 border border-wiki-border dark:border-slate-700 bg-white dark:bg-slate-900 text-amber-700 dark:text-amber-400 hover:bg-amber-50 font-semibold">Neutral</button>
              <button @click="castRFCVote(threadRFC[thread.id].id, thread.id, 'OPPOSE')" class="px-2.5 py-0.5 border border-wiki-border dark:border-slate-700 bg-white dark:bg-slate-900 text-rose-700 dark:text-rose-400 hover:bg-rose-50 font-semibold">Oppose</button>
            </div>
          </div>

          <!-- Comments Thread -->
          <div>
            <div
              v-for="comment in (threadComments[thread.id] || [])"
              :key="comment.id"
              class="py-2 border-b border-dotted border-slate-200 dark:border-slate-800 text-sm"
            >
              <p class="text-slate-800 dark:text-slate-200 whitespace-pre-wrap">{{ comment.content }}</p>
              <p class="text-xs text-slate-500 mt-1">— {{ comment.author_id }}, {{ formatTime(comment.created_at) }}</p>
            </div>
            <div v-if="(threadComments[thread.id] || []).length === 0" class="text-sm text-slate-400 italic py-2">
              No replies yet on this thread.
            </div>
          </div>

          <!-- Reply Box -->
          <form @submit.prevent="submitReply(thread.id)" class="space-y-2">
            <textarea
              v-model="replyText[thread.id]"
              rows="2"
              placeholder="Write a constructive response or counterpoint..."
              class="w-full p-2 text-sm border border-wiki-border dark:border-slate-700 bg-white dark:bg-slate-950 text-slate-900 dark:text-slate-100 focus:outline-none focus:border-namu-600 resize-y"
              required
            ></textarea>
            <div class="flex justify-end">
              <button
                type="submit"
                :disabled="submittingReply[thread.id]"
                class="px-4 py-1 bg-namu-600 hover:bg-namu-700 text-white text-sm disabled:opacity-50"
              >
                {{ submittingReply[thread.id] ? 'Posting…' : 'Post reply' }}
              </button>
            </div>
          </form>
        </div>
      </section>
    </div>

    <!-- Start New Thread Modal -->
    <div
      v-if="showNewThreadModal"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4"
    >
      <div class="bg-white dark:bg-slate-900 border border-wiki-border dark:border-slate-700 max-w-lg w-full p-5 shadow-xl">
        <div class="flex items-center justify-between border-b border-wiki-border dark:border-slate-700 pb-2">
          <h2 class="text-xl font-bold text-slate-900 dark:text-slate-100">New discussion topic</h2>
          <button @click="showNewThreadModal = false" class="text-slate-400 hover:text-slate-600 px-1">✕</button>
        </div>

        <form @submit.prevent="createThread" class="mt-3 space-y-3 text-sm">
          <div>
            <label class="block font-semibold text-slate-700 dark:text-slate-300 mb-1">Subject</label>
            <input
              v-model="newThreadTitle"
              type="text"
              placeholder="e.g. Propose restructuring the Asymptotic Complexity section"
              class="w-full px-2.5 py-1.5 border border-wiki-border dark:border-slate-700 bg-white dark:bg-slate-950 text-slate-900 dark:text-slate-100 focus:outline-none focus:border-namu-600"
              required
            />
          </div>

          <div>
            <label class="block font-semibold text-slate-700 dark:text-slate-300 mb-1">Opening statement</label>
            <textarea
              v-model="newThreadContent"
              rows="4"
              placeholder="Describe your suggestion, rationale, or query..."
              class="w-full p-2.5 border border-wiki-border dark:border-slate-700 bg-white dark:bg-slate-950 text-slate-900 dark:text-slate-100 focus:outline-none focus:border-namu-600"
              required
            ></textarea>
          </div>

          <div class="border border-wiki-border dark:border-slate-700 bg-wiki-pagebg dark:bg-slate-800/40 p-3 space-y-2">
            <label class="flex items-center gap-2 cursor-pointer font-semibold">
              <input type="checkbox" v-model="attachRFC" />
              <span>Attach a formal Request for Comments (RFC)</span>
            </label>

            <div v-if="attachRFC" class="space-y-2 pt-2 border-t border-wiki-border dark:border-slate-700">
              <div>
                <label class="block text-slate-500 mb-0.5 text-[13px]">RFC question</label>
                <input
                  v-model="rfcQuestion"
                  type="text"
                  placeholder="Should we split this article into two independent pages?"
                  class="w-full px-2.5 py-1 border border-wiki-border dark:border-slate-700 bg-white dark:bg-slate-950"
                  required
                />
              </div>
              <div>
                <label class="block text-slate-500 mb-0.5 text-[13px]">Quorum (minimum votes)</label>
                <input
                  v-model.number="rfcQuorum"
                  type="number"
                  min="1"
                  max="100"
                  class="w-24 px-2.5 py-1 border border-wiki-border dark:border-slate-700 bg-white dark:bg-slate-950 font-mono"
                />
              </div>
            </div>
          </div>

          <div class="flex items-center justify-end gap-2 pt-1">
            <button
              type="button"
              @click="showNewThreadModal = false"
              class="px-4 py-1.5 border border-wiki-border dark:border-slate-700 text-slate-600 dark:text-slate-300"
            >
              Cancel
            </button>
            <button
              type="submit"
              :disabled="submittingNewThread"
              class="px-5 py-1.5 bg-namu-600 hover:bg-namu-700 text-white disabled:opacity-50"
            >
              {{ submittingNewThread ? 'Publishing…' : 'Publish topic' }}
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
  pinned: boolean
  locked: boolean
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

interface RFCVote {
  user_id: string
  choice: string // 'SUPPORT' | 'OPPOSE' | 'NEUTRAL'
}

interface RFCItem {
  id: string
  thread_id: string
  question: string
  quorum: number
  closed: boolean
  votes: RFCVote[]
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
    const rfc = await api.get<RFCItem>(`/talk/threads/${threadId}/rfc`)
    if (rfc && rfc.question) {
      if (!Array.isArray(rfc.votes)) rfc.votes = []
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
      if (!Array.isArray(updated.votes)) updated.votes = []
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
    case 'RFC':
      return 'bg-forest-50 text-forest-700 border-forest-500 dark:bg-emerald-950 dark:text-emerald-400 dark:border-emerald-800'
    case 'UNDER_DISCUSSION':
      return 'bg-amber-50 text-amber-700 border-amber-300 dark:bg-amber-950 dark:text-amber-400 dark:border-amber-800'
    case 'CLOSED':
    case 'ARCHIVED':
      return 'bg-slate-100 text-slate-600 border-slate-300 dark:bg-slate-800 dark:text-slate-400 dark:border-slate-700'
    default:
      return 'bg-slate-100 text-slate-700 border-slate-200'
  }
}

const getVoteCount = (rfc: RFCItem, voteType: string) => {
  if (!rfc || !Array.isArray(rfc.votes)) return 0
  return rfc.votes.filter(v => v && v.choice === voteType).length
}

const getRFCVotePct = (rfc: RFCItem, voteType: string) => {
  if (!rfc || !Array.isArray(rfc.votes)) return 0
  const total = rfc.votes.length
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

useSeoMeta({
  title: () => `Talk: ${displayTitle.value} — Treewiki`,
  description: () => `Community discussion, RFC proposals, and editorial consensus for ${displayTitle.value} on Treewiki.`,
  ogTitle: () => `Talk: ${displayTitle.value} — Treewiki`,
  ogDescription: () => `Join the consensus discussion for ${displayTitle.value}.`,
  ogType: 'article',
})

onMounted(() => {
  loadThreads()
})
</script>
