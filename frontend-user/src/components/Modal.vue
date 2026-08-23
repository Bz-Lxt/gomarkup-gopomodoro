<script setup lang="ts">
defineProps<{ title: string }>()
const emit = defineEmits<{ close: [] }>()
</script>

<template>
  <div class="fixed inset-0 z-40 bg-black/60 flex items-center justify-center p-4" @click.self="emit('close')">
    <div class="card w-full max-w-md p-6">
      <h3 class="font-display text-xl mb-4">{{ title }}</h3>
      <slot />
    </div>
  </div>
</template>
