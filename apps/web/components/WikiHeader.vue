<template>
  <header class="sticky top-0 z-50 bg-white/95 dark:bg-slate-900/95 backdrop-blur border-b border-slate-200 dark:border-slate-800 transition-colors">
    <div class="max-w-7xl mx-auto px-4 h-14 flex items-center justify-between gap-4">
      <!-- Logo & Brand -->
      <NuxtLink to="/" class="flex items-center gap-2.5 font-bold text-lg text-slate-800 dark:text-slate-100 hover:opacity-90">
        <img src="/logos/treewiki.svg" alt="Treewiki" class="w-8 h-8 rounded" />
        <span class="tracking-tight text-forest-700 dark:text-forest-500">Tree<span class="text-slate-700 dark:text-slate-200">wiki</span></span>
      </NuxtLink>

      <!-- Search Bar -->
      <div class="flex-1 max-w-lg relative">
        <div class="relative flex items-center">
          <input
            v-model="searchQuery"
            @keyup.enter="handleSearch"
            type="text"
            placeholder="Search Treewiki or enter title... (Ctrl+K)"
            class="w-full pl-9 pr-4 py-1.5 text-sm rounded-lg bg-slate-100 dark:bg-slate-800 border border-slate-300 dark:border-slate-700 text-slate-800 dark:text-slate-200 focus:outline-none focus:ring-2 focus:ring-forest-500 transition"
          />
          <span class="absolute left-2.5 text-slate-400 text-xs">🔍</span>
        </div>
      </div>

      <!-- Action Navigation -->
      <nav class="flex items-center gap-2">
        <NuxtLink to="/tree" class="px-3 py-1.5 text-xs font-semibold rounded-md text-forest-700 dark:text-forest-400 hover:bg-forest-50 dark:hover:bg-slate-800 transition">
          🌲 Topic Trees
        </NuxtLink>
        <NuxtLink to="/special/recent-changes" class="px-3 py-1.5 text-xs font-semibold rounded-md text-slate-600 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800 transition">
          ⚡ Changes
        </NuxtLink>
        <button
          @click="toggleTheme"
          class="p-2 rounded-lg text-slate-500 hover:bg-slate-100 dark:hover:bg-slate-800 transition"
          title="Toggle Dark Mode"
        >
          <span v-if="isDark">🌙</span>
          <span v-else>☀️</span>
        </button>
      </nav>
    </div>
  </header>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'

const searchQuery = ref('')
const isDark = ref(false)
const router = useRouter()

const handleSearch = () => {
  if (searchQuery.value.trim()) {
    router.push(`/w/${encodeURIComponent(searchQuery.value.trim().replace(/ /g, '_'))}`)
  }
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
  if (localStorage.getItem('treewiki_theme') === 'dark' || window.matchMedia('(prefers-color-scheme: dark)').matches) {
    isDark.value = true
    document.documentElement.classList.add('dark')
  }
})
</script>

