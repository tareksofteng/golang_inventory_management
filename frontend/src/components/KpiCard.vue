<script setup>
import Icon from './Icon.vue'

// Stat tile: label (sentence case) + value + optional hint. The icon rides a
// soft tinted chip — saturated fills are reserved for small marks, not blocks.
const tones = {
  brand: 'bg-brand-50 text-brand-600 dark:bg-brand-600/15 dark:text-brand-200',
  emerald: 'bg-emerald-50 text-emerald-600 dark:bg-emerald-500/15 dark:text-emerald-300',
  blue: 'bg-blue-50 text-blue-600 dark:bg-blue-500/15 dark:text-blue-300',
  amber: 'bg-amber-50 text-amber-600 dark:bg-amber-500/15 dark:text-amber-300',
  red: 'bg-red-50 text-red-600 dark:bg-red-500/15 dark:text-red-300',
  slate: 'bg-slate-100 text-slate-500 dark:bg-slate-700 dark:text-slate-300',
}

defineProps({
  label: String,
  value: [String, Number],
  icon: String,
  hint: String,
  tone: { type: String, default: 'brand' },
  compact: Boolean,
})
</script>

<template>
  <div class="card p-4 transition hover:shadow-md sm:p-5">
    <div class="flex items-start justify-between gap-3">
      <div class="min-w-0">
        <div class="truncate text-sm text-slate-500 dark:text-slate-400">{{ label }}</div>
        <!-- Proportional figures: tabular-nums makes large values look loose. -->
        <div class="mt-1.5 truncate font-semibold tracking-tight" :class="compact ? 'text-xl' : 'text-2xl'">
          {{ value }}
        </div>
        <div v-if="hint" class="mt-1 truncate text-xs text-slate-400">{{ hint }}</div>
      </div>
      <span :class="tones[tone]" class="grid h-10 w-10 shrink-0 place-items-center rounded-xl">
        <Icon :name="icon" :size="20" />
      </span>
    </div>
  </div>
</template>
