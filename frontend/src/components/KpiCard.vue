<script setup>
defineProps({
  label: String,
  value: [String, Number],
  icon: String,
  accent: { type: String, default: 'bg-brand-600' },
})
</script>

<template>
  <div class="card flex items-center gap-4 p-5">
    <span :class="accent" class="grid h-12 w-12 shrink-0 place-items-center rounded-xl text-xl text-white">
      {{ icon }}
    </span>
    <div class="min-w-0">
      <div class="truncate text-2xl font-bold">{{ value }}</div>
      <div class="text-sm text-slate-500 dark:text-slate-400">{{ label }}</div>
    </div>
  </div>
</template>
