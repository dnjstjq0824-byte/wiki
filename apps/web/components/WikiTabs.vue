<!-- apps/web/components/WikiTabs.vue — document action tabs -->
<template>
  <nav :aria-label="`Page tabs for ${slug}`" class="flex items-center justify-end gap-1 text-sm mb-2">
    <NuxtLink
      v-for="t in tabs"
      :key="t.id"
      :to="t.href"
      :aria-current="t.id === active ? 'page' : undefined"
      class="px-3 py-1 border"
      :class="t.id === active
        ? 'bg-namu-600 border-namu-600 text-white font-semibold'
        : 'bg-white dark:bg-slate-900 border-wiki-border dark:border-slate-700 text-slate-600 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-800'"
    >
      {{ t.label }}
    </NuxtLink>
  </nav>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  slug: string
  active: 'article' | 'talk' | 'edit' | 'history'
}>()

const tabs = computed(() => [
  { id: 'article', label: 'Article', href: `/w/${props.slug}` },
  { id: 'talk', label: 'Talk', href: `/talk/${props.slug}` },
  { id: 'edit', label: 'Edit', href: `/edit/${props.slug}` },
  { id: 'history', label: 'History', href: `/history/${props.slug}` },
])
</script>
