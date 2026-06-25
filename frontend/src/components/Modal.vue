<script setup>
defineProps({ title: String })
const emit = defineEmits(['close'])
</script>

<template>
  <div class="fixed inset-0 z-40 flex items-center justify-center p-4">
    <div class="absolute inset-0 bg-black/40" @click="emit('close')" />
    <div class="card relative z-10 w-full max-w-lg p-6">
      <div class="mb-4 flex items-center justify-between">
        <h3 class="text-lg font-semibold">{{ title }}</h3>
        <button class="btn-ghost !px-2 !py-1" @click="emit('close')">✕</button>
      </div>
      <slot />
    </div>
  </div>
</template>
