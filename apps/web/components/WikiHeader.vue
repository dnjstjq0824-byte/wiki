<!-- apps/web/components/WikiHeader.vue -->
<template>
  <header class="bg-white dark:bg-slate-900 border-b border-wiki-border dark:border-slate-800">
    <!-- Personal tools bar -->
    <div class="max-w-[72rem] mx-auto px-3 md:px-5 flex items-center justify-end gap-3 h-8 text-xs">
      <button
        @click="toggleTheme"
        :aria-pressed="isDark"
        aria-label="Toggle dark mode"
        class="text-slate-500 hover:text-slate-800 dark:hover:text-slate-200"
        title="Toggle appearance"
      >
        {{ isDark ? 'Dark' : 'Light' }}
      </button>
      <template v-if="authStore.isAuthenticated">
        <span class="font-semibold text-slate-700 dark:text-slate-200">{{ authStore.user?.username }}</span>
        <span class="font-mono text-[10px] px-1 border border-wiki-border dark:border-slate-700 text-slate-500">
          {{ authStore.userRole }}
        </span>
        <button @click="authStore.clearAuth" class="text-wiki-link dark:text-sky-400 hover:underline" title="Sign Out">
          Log out
        </button>
      </template>
      <template v-else>
        <button @click="isAuthModalOpen = true" class="text-wiki-link dark:text-sky-400 hover:underline">
          Log in
        </button>
        <button @click="isAuthModalOpen = true" class="text-wiki-link dark:text-sky-400 hover:underline">
          Create account
        </button>
      </template>
    </div>

    <!-- Main bar: logo, wordmark, search -->
    <div class="max-w-[72rem] mx-auto px-3 md:px-5 pb-2 flex items-end justify-between gap-4">
      <NuxtLink to="/" class="flex items-center gap-2.5 shrink-0" aria-label="Treewiki home">
        <img src="/logos/treewiki.svg" alt="" class="w-10 h-10" aria-hidden="true" />
        <span class="leading-tight">
          <span class="block font-wiki text-xl text-slate-900 dark:text-slate-100">Treewiki</span>
          <span class="block text-[11px] text-slate-500">The free knowledge platform</span>
        </span>
      </NuxtLink>

      <div class="flex-1 max-w-md relative" ref="searchContainerRef">
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
            class="w-full px-2.5 py-1 text-sm border border-wiki-border dark:border-slate-700 bg-white dark:bg-slate-950 text-slate-900 dark:text-slate-100 focus:outline-none focus:border-wiki-link"
          />
          <button
            @click="handleSearch"
            class="px-3 text-sm border border-l-0 border-wiki-border dark:border-slate-700 bg-slate-100 dark:bg-slate-800 text-slate-700 dark:text-slate-200 hover:bg-slate-200 dark:hover:bg-slate-700"
          >
            Go
          </button>
        </div>

        <div
          v-if="showSuggestions && suggestions.length > 0"
          id="search-suggestions"
          role="listbox"
          aria-label="Search suggestions"
          class="absolute top-full left-0 right-0 mt-0 bg-white dark:bg-slate-900 border border-t-0 border-wiki-border dark:border-slate-700 shadow-md z-50 text-sm"
        >
          <div
            v-for="(item, idx) in suggestions"
            :key="idx"
            role="option"
            :aria-selected="false"
            @click="selectSuggestion(item)"
            class="px-2.5 py-1.5 hover:bg-sky-50 dark:hover:bg-slate-800 cursor-pointer text-wiki-link dark:text-sky-400"
          >
            {{ item }}
          </div>
        </div>
      </div>
    </div>

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
