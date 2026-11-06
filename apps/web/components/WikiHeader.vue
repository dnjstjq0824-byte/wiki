<!-- apps/web/components/WikiHeader.vue -->
<template>
  <header class="bg-namu-700 dark:bg-namu-900 text-white">
    <div class="max-w-5xl mx-auto px-3 md:px-4 py-2.5 flex items-center gap-3">
      <NuxtLink to="/" class="flex items-center gap-2 shrink-0" aria-label="Treewiki home">
        <img src="/logos/treewiki.svg" alt="" class="w-9 h-9 bg-white/10 p-0.5" aria-hidden="true" />
        <span class="leading-tight">
          <span class="block font-bold text-lg tracking-tight">Treewiki</span>
        </span>
      </NuxtLink>

      <div class="flex-1 max-w-md ml-auto relative" ref="searchContainerRef">
        <div class="flex items-stretch" role="search">
          <label for="site-search" class="sr-only">Search Treewiki</label>
          <input
            id="site-search"
            v-model="searchQuery"
            @input="onSearchInput"
            @keyup.enter="handleSearch"
            @focus="showSuggestions = true"
            type="text"
            role="combobox"
            :aria-expanded="showSuggestions && suggestions.length > 0"
            aria-controls="search-suggestions"
            aria-autocomplete="list"
            placeholder="Search Treewiki"
            class="w-full px-2.5 py-1.5 text-sm text-slate-900 bg-white focus:outline-none"
          />
          <button
            @click="handleSearch"
            aria-label="Search"
            class="px-3.5 bg-namu-900 dark:bg-black/40 hover:bg-black/30 text-white text-sm font-bold"
          >
            Go
          </button>
        </div>

        <div
          v-if="showSuggestions && suggestions.length > 0"
          id="search-suggestions"
          role="listbox"
          aria-label="Search suggestions"
          class="absolute top-full left-0 right-0 bg-white text-slate-900 shadow-lg z-50 text-sm border border-slate-200"
        >
          <div
            v-for="(item, idx) in suggestions"
            :key="idx"
            role="option"
            :aria-selected="false"
            @click="selectSuggestion(item)"
            class="px-2.5 py-1.5 hover:bg-namu-50 cursor-pointer text-wiki-link"
          >
            {{ item }}
          </div>
        </div>
      </div>

      <div class="hidden sm:flex items-center gap-2.5 text-[13px] shrink-0">
        <button
          @click="toggleTheme"
          :aria-pressed="isDark"
          aria-label="Toggle dark mode"
          class="text-white/80 hover:text-white"
          title="Toggle appearance"
        >
          {{ isDark ? 'Dark' : 'Light' }}
        </button>
        <template v-if="authStore.isAuthenticated">
          <span class="font-semibold">{{ authStore.user?.username }}</span>
          <span class="font-mono text-[10px] px-1 bg-white/20">{{ authStore.userRole }}</span>
          <button @click="authStore.clearAuth" class="text-white/80 hover:text-white hover:underline" title="Sign Out">
            Log out
          </button>
        </template>
        <template v-else>
          <button @click="isAuthModalOpen = true" class="hover:underline">Log in</button>
          <button @click="isAuthModalOpen = true" class="px-2 py-0.5 bg-white text-namu-700 font-semibold hover:bg-namu-50">
            Sign up
          </button>
        </template>
      </div>
    </div>

    <!-- Mobile user tools -->
    <div class="sm:hidden max-w-5xl mx-auto px-3 pb-2 flex items-center gap-3 text-[13px]">
      <template v-if="authStore.isAuthenticated">
        <span class="font-semibold">{{ authStore.user?.username }}</span>
        <button @click="authStore.clearAuth" class="text-white/80 hover:underline">Log out</button>
      </template>
      <template v-else>
        <button @click="isAuthModalOpen = true" class="hover:underline">Log in</button>
        <button @click="isAuthModalOpen = true" class="hover:underline">Sign up</button>
      </template>
    </div>

    <!-- Menu strip -->
    <nav aria-label="Site" class="bg-namu-800 dark:bg-black/40">
      <ul class="max-w-5xl mx-auto px-3 md:px-4 flex flex-wrap items-center gap-x-4 gap-y-1 py-1.5 text-[13px]">
        <li><NuxtLink to="/" class="text-white/90 hover:text-white hover:underline">Main page</NuxtLink></li>
        <li><NuxtLink to="/tree" class="text-white/90 hover:text-white hover:underline">Topic trees</NuxtLink></li>
        <li><NuxtLink to="/special/random" class="text-white/90 hover:text-white hover:underline">Random article</NuxtLink></li>
        <li><NuxtLink to="/special/recent-changes" class="text-white/90 hover:text-white hover:underline">Recent changes</NuxtLink></li>
        <li><NuxtLink to="/w/Treewiki:Sandbox" class="text-white/90 hover:text-white hover:underline">Sandbox</NuxtLink></li>
        <li v-if="isAdmin"><NuxtLink to="/admin" class="text-white/90 hover:text-white hover:underline font-semibold">Admin</NuxtLink></li>
      </ul>
    </nav>

    <AuthModal :is-open="isAuthModalOpen" @close="isAuthModalOpen = false" />
  </header>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useApi } from '../composables/useApi'

const searchQuery = ref('')
const suggestions = ref<string[]>([])
const showSuggestions = ref(false)
const searchContainerRef = ref<HTMLElement | null>(null)
const isDark = ref(false)
const isAuthModalOpen = ref(false)

const router = useRouter()
const authStore = useAuthStore()
const api = useApi()

const isAdmin = computed(() =>
  ['SYSOP', 'OVERSIGHT', 'BUREAUCRAT', 'SYSTEM_ADMIN'].includes(authStore.userRole)
)

let debounceTimer: any = null

const onSearchInput = () => {
  clearTimeout(debounceTimer)
  const q = searchQuery.value.trim()
  if (!q) {
    suggestions.value = []
    return
  }

  debounceTimer = setTimeout(async () => {
    try {
      const res = await api.get<string[]>('/search/autocomplete', { prefix: q, limit: 6 })
      suggestions.value = res || []
    } catch {
      suggestions.value = []
    }
  }, 200)
}

const handleSearch = () => {
  showSuggestions.value = false
  const q = searchQuery.value.trim()
  if (q) {
    router.push(`/search?q=${encodeURIComponent(q)}`)
  }
}

const selectSuggestion = (title: string) => {
  showSuggestions.value = false
  searchQuery.value = title
  router.push(`/w/${encodeURIComponent(title.replace(/ /g, '_'))}`)
}

const toggleTheme = () => {
  isDark.value = !isDark.value
  if (isDark.value) {
    document.documentElement.classList.add('dark')
    localStorage.setItem('treewiki_theme', 'dark')
  } else {
    document.documentElement.classList.remove('dark')
    localStorage.setItem('treewiki_theme', 'light')
  }
}

onMounted(() => {
  authStore.initAuth()

  if (localStorage.getItem('treewiki_theme') === 'dark' || window.matchMedia('(prefers-color-scheme: dark)').matches) {
    isDark.value = true
    document.documentElement.classList.add('dark')
  }

  // Close suggestions on outside click
  document.addEventListener('click', (e) => {
    if (searchContainerRef.value && !searchContainerRef.value.contains(e.target as Node)) {
      showSuggestions.value = false
    }
  })
})
</script>
