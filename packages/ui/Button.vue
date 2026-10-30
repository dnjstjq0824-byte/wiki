<!-- packages/ui/Button.vue -->
<template>
  <button class="tw-btn" :class="variant" @click="$emit('click')">
    <slot />
  </button>
</template>

<script setup lang="ts">
defineProps<{ variant?: string }>();
defineEmits(['click']);
</script>

<style scoped>
.tw-btn { padding: 0.5rem 1rem; border-radius: 0.375rem; }
</style>