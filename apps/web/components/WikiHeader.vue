<!-- apps/web/components/WikiHeader.vue -->
<template>
  <header class="sticky top-0 z-50 bg-white/95 dark:bg-slate-900/95 backdrop-blur border-b border-slate-200 dark:border-slate-800 transition-colors">
    <div class="max-w-7xl mx-auto px-4 h-14 flex items-center justify-between gap-4">
      <!-- Logo & Brand -->
      <NuxtLink to="/" class="flex items-center gap-2.5 font-bold text-lg text-slate-800 dark:text-slate-100 hover:opacity-90">
        <img src="/logos/treewiki.svg" alt="Treewiki" class="w-8 h-8 rounded" />
        <span class="tracking-tight text-forest-700 dark:text-forest-500">Tree<span class="text-slate-700 dark:text-slate-200">wiki</span></span>
      </NuxtLink>

      <!-- Search Bar with Live Autocomplete -->
      <div class="flex-1 max-w-lg relative" ref="searchContainerRef">
        <div class="relative flex items-center" role="search">
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
            placeholder="Search Treewiki or enter title... (Press Enter)"
            class="w-full pl-9 pr-4 py-1.5 text-sm rounded-lg bg-slate-100 dark:bg-slate-800 border border-slate-300 dark:border-slate-700 text-slate-800 dark:text-slate-200 focus:outline-none focus:ring-2 focus:ring-forest-500 transition"
          />
          <span class="absolute left-2.5 text-slate-400 text-xs" aria-hidden="true">🔍</span>
        </div>

        <!-- Autocomplete Dropdown Menu -->
        <div
          v-if="showSuggestions && suggestions.length > 0"
          id="search-suggestions"
          role="listbox"
          aria-label="Search suggestions"
          class="absolute top-full left-0 right-0 mt-1 bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl shadow-xl overflow-hidden z-50 text-xs"
        >
          <div
            v-for="(item, idx) in suggestions"
            :key="idx"
            role="option"
            :aria-selected="false"
            @click="selectSuggestion(item)"
            class="px-4 py-2 hover:bg-slate-100 dark:hover:bg-slate-800/80 cursor-pointer flex items-center justify-between transition"
          >
            <span class="font-medium text-slate-800 dark:text-slate-200 flex items-center gap-2">
              <span>📄</span> {{ item }}
            </span>
            <span class="text-[10px] text-slate-400">Jump to article →</span>
          </div>
        </div>
      </div>

      <!-- Action Navigation & Auth Controls -->
      <nav aria-label="Primary" class="flex items-center gap-2">
        <NuxtLink to="/tree" class="px-3 py-1.5 text-xs font-semibold rounded-md text-forest-700 dark:text-forest-400 hover:bg-forest-50 dark:hover:bg-slate-800 transition">
          🌲 Topic Trees
        </NuxtLink>
        <NuxtLink to="/special/recent-changes" class="px-3 py-1.5 text-xs font-semibold rounded-md text-slate-600 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 transition">
          ⚡ Changes
        </NuxtLink>

        <!-- Dark/Light Theme Toggle -->
        <button
          @click="toggleTheme"
          :aria-pressed="isDark"
          aria-label="Toggle dark mode"
          class="p-2 rounded-lg text-slate-500 hover:bg-slate-100 dark:hover:bg-slate-800 transition"
          title="Toggle Dark Mode"
        >
          <span v-if="isDark">🌙</span>
          <span v-else>☀️</span>
        </button>

        <!-- User Authentication State -->
        <div v-if="authStore.isAuthenticated" class="flex items-center gap-2 pl-2 border-l border-slate-200 dark:border-slate-800">
          <div class="flex items-center gap-1.5">
            <span class="w-6 h-6 rounded-full bg-forest-600 text-white flex items-center justify-center font-bold text-[10px]">
              {{ (authStore.user?.username || 'U').charAt(0).toUpperCase() }}
            </span>
            <span class="text-xs font-bold text-slate-700 dark:text-slate-200">
              {{ authStore.user?.username }}
            </span>
            <span class="text-[9px] px-1.5 py-0.5 rounded font-mono font-bold bg-slate-100 dark:bg-slate-800 text-forest-600 dark:text-forest-400 border border-slate-200 dark:border-slate-700">
              {{ authStore.userRole }}
            </span>
          </div>
          <button
            @click="authStore.clearAuth"
            class="text-xs text-slate-400 hover:text-rose-600 transition ml-1"
            title="Sign Out"
          >
            Logout
          </button>
        </div>

        <div v-else class="flex items-center gap-1.5 pl-2 border-l border-slate-200 dark:border-slate-800">
          <button
            @click="isAuthModalOpen = true"
            class="px-3 py-1.5 rounded-lg text-xs font-semibold text-slate-700 dark:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800 transition"
          >
            Sign In
          </button>
          <button
            @click="isAuthModalOpen = true"
            class="px-3 py-1.5 rounded-lg text-xs font-bold bg-forest-600 hover:bg-forest-500 text-white shadow-2xs transition"
          >
            Sign Up
          </button>
        </div>
      </nav>
    </div>

    <!-- Auth Modal Dialog -->
    <AuthModal :is-open="isAuthModalOpen" @close="isAuthModalOpen = false" />
  </header>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
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
