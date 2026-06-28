<script setup>
defineProps({
  title: String,
  size: { type: String, default: 'max-w-lg' },
})
const emit = defineEmits(['close'])
</script>

<template>
  <div class="fixed inset-0 z-40 flex items-start justify-center overflow-y-auto p-4 sm:items-center">
    <div class="absolute inset-0 bg-black/40 backdrop-blur-sm" @click="emit('close')" />
    <div class="card relative z-10 my-4 max-h-[92vh] w-full overflow-y-auto p-6" :class="size">
      <div class="mb-4 flex items-center justify-between">
        <h3 class="text-lg font-semibold">{{ title }}</h3>
        <button class="btn-ghost !px-2 !py-1" @click="emit('close')">✕</button>
      </div>
      <slot />
    </div>
  </div>
</template>
