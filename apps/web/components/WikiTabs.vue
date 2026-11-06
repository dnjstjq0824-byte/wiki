<!-- apps/web/components/WikiTabs.vue — classic content tabs for a page -->
<template>
  <nav :aria-label="`Page tabs for ${slug}`" class="flex items-end gap-0 text-sm border-b border-wiki-border dark:border-slate-700 -mb-px">
    <NuxtLink
      v-for="t in tabs"
      :key="t.id"
      :to="t.href"
      :aria-current="t.id === active ? 'page' : undefined"
      class="px-3 py-1 border border-b-0 -mb-px mr-1"
      :class="t.id === active
        ? 'bg-white dark:bg-slate-900 border-wiki-border dark:border-slate-700 text-slate-900 dark:text-slate-100 font-semibold'
        : 'bg-slate-100 dark:bg-slate-800 border-slate-200 dark:border-slate-700 text-wiki-link dark:text-sky-400 hover:underline'"
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
