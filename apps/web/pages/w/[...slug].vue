<template>
  <article class="bg-white dark:bg-slate-900 border border-slate-200 dark:border-slate-800 rounded-xl p-6 md:p-8 shadow-sm space-y-6">
    <!-- Article Header & Action Tab Bar -->
    <div class="border-b border-slate-200 dark:border-slate-800 pb-4">
      <div class="flex flex-wrap items-center justify-between gap-4 mb-2">
        <h1 class="text-3xl font-extrabold text-slate-900 dark:text-slate-100 tracking-tight">
          {{ displayTitle }}
        </h1>
        <!-- Tab Actions -->
        <div class="flex items-center gap-1 bg-slate-100 dark:bg-slate-800 p-1 rounded-lg text-xs font-semibold">
          <NuxtLink :to="`/w/${slug}`" class="px-3 py-1.5 rounded-md bg-white dark:bg-slate-700 text-forest-700 dark:text-forest-400 shadow-xs">
            Article
          </NuxtLink>
          <NuxtLink :to="`/talk/${slug}`" class="px-3 py-1.5 rounded-md text-slate-600 dark:text-slate-300 hover:text-slate-900">
            Talk (3)
          </NuxtLink>
          <NuxtLink :to="`/edit/${slug}`" class="px-3 py-1.5 rounded-md text-slate-600 dark:text-slate-300 hover:text-slate-900">
            ✏️ Edit
          </NuxtLink>
          <NuxtLink :to="`/history/${slug}`" class="px-3 py-1.5 rounded-md text-slate-600 dark:text-slate-300 hover:text-slate-900">
            🕒 History
          </NuxtLink>
        </div>
      </div>
      <div class="text-xs text-slate-500 flex items-center gap-2">
        <span>From Treewiki, the free knowledge platform</span>
        <span>•</span>
        <span>Last updated: 14 minutes ago by <strong class="text-slate-700 dark:text-slate-300">Curator_Alpha</strong></span>
      </div>
    </div>

    <!-- Article Content / Rendered Body -->
    <div class="prose dark:prose-invert max-w-none text-slate-800 dark:text-slate-200 leading-relaxed space-y-4">
      <div class="p-4 rounded-lg bg-slate-50 dark:bg-slate-800/50 border border-slate-200 dark:border-slate-700 text-sm">
        <p class="font-semibold text-slate-700 dark:text-slate-300 mb-1">📋 Table of Contents</p>
        <ul class="list-disc list-inside space-y-1 text-forest-700 dark:text-forest-400 text-xs">
          <li><a href="#overview" class="hover:underline">1. Overview & Foundational Principles</a></li>
          <li><a href="#taxonomy" class="hover:underline">2. Knowledge Taxonomy & Topic Trees</a></li>
          <li><a href="#references" class="hover:underline">3. References & Citations</a></li>
        </ul>
      </div>

      <section id="overview" class="space-y-2">
        <h2 class="text-xl font-bold text-slate-900 dark:text-slate-100 border-b border-slate-200 dark:border-slate-800 pb-1">
          1. Overview & Foundational Principles
        </h2>
        <p>
          <strong>{{ displayTitle }}</strong> represents a cornerstone discipline within modern scientific epistemology. It investigates systematic principles, structural models, and practical applications that govern empirical discovery and mathematical formulation<sup><a href="#references" class="text-forest-600">[1]</a></sup>.
        </p>
        <p>
          Under the collaborative governance of Treewiki, documentation regarding this subject adheres to strict <a href="/w/Treewiki:NPOV" class="text-sky-600 hover:underline">Neutral Point of View</a> and verifiable source standards<sup><a href="#references" class="text-forest-600">[2]</a></sup>.
        </p>
      </section>

      <section id="taxonomy" class="space-y-2">
        <h2 class="text-xl font-bold text-slate-900 dark:text-slate-100 border-b border-slate-200 dark:border-slate-800 pb-1">
          2. Knowledge Taxonomy & Topic Trees
        </h2>
        <p>
          In the Treewiki DAG (Directed Acyclic Graph) knowledge hierarchy, this article is positioned as an intermediate prerequisite node for higher-order specializations.
        </p>
        <div class="p-4 rounded-xl bg-forest-50/60 dark:bg-slate-800/60 border border-forest-200/80 dark:border-forest-900/50 flex items-center justify-between">
          <div class="text-xs">
            <span class="font-bold text-forest-900 dark:text-forest-300">Interactive Topic Tree Node</span>
            <p class="text-slate-600 dark:text-slate-400">Parent: Science → Formal Sciences → Computational Systems</p>
          </div>
          <NuxtLink to="/tree" class="px-3 py-1 rounded bg-forest-600 text-white font-medium text-xs hover:bg-forest-700">
            Inspect Tree Node →
          </NuxtLink>
        </div>
      </section>

      <section id="references" class="pt-6 border-t border-slate-200 dark:border-slate-800 space-y-2">
        <h2 class="text-lg font-bold text-slate-900 dark:text-slate-100">
          3. References & Citations
        </h2>
        <ol class="list-decimal list-inside text-xs text-slate-600 dark:text-slate-400 space-y-1">
          <li>International Journal of Foundational Sciences (2025). <em>Epistemological Foundations of Structured Knowledge</em>. DOI: 10.1016/j.ijfs.2025.04.</li>
          <li>Booking Network Knowledge Architecture Group (2026). <em>Treewiki Semantic Taxonomy Standards</em>.</li>
        </ol>
      </section>
    </div>

    <!-- Category Footnotes -->
    <div class="mt-8 pt-4 border-t border-slate-100 dark:border-slate-800 flex flex-wrap items-center gap-2 text-xs">
      <span class="font-bold text-slate-400">Categories:</span>
      <span class="px-2.5 py-1 rounded-full bg-slate-100 dark:bg-slate-800 text-forest-700 dark:text-forest-400 font-medium">Science</span>
      <span class="px-2.5 py-1 rounded-full bg-slate-100 dark:bg-slate-800 text-forest-700 dark:text-forest-400 font-medium">Technology</span>
      <span class="px-2.5 py-1 rounded-full bg-slate-100 dark:bg-slate-800 text-forest-700 dark:text-forest-400 font-medium">Academic Disciplines</span>
    </div>
  </article>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const slug = computed(() => {
  const p = route.params.slug
  return Array.isArray(p) ? p.join('/') : p || 'Main_Page'
})

const displayTitle = computed(() => {
  return slug.value.replace(/_/g, ' ')
})
</script>

