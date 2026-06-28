<script setup>
import { ref, computed } from 'vue'

const props = defineProps({
  products: { type: Array, default: () => [] },
})
const emit = defineEmits(['select'])

const query = ref('')
const open = ref(false)

const filtered = computed(() => {
  const q = query.value.toLowerCase().trim()
  const list = q
    ? props.products.filter(
        (p) => p.name.toLowerCase().includes(q) || (p.sku || '').toLowerCase().includes(q),
      )
    : props.products
  return list.slice(0, 8)
})

function pick(p) {
  emit('select', p)
  query.value = ''
  open.value = false
}

// Small delay so a click on an option registers before blur closes the list.
const close = () => setTimeout(() => (open.value = false), 150)
</script>

<template>
  <div class="relative">
    <div class="relative">
      <span class="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-slate-400">
        <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="m21 21-4.3-4.3m1.8-4.45a6.25 6.25 0 1 1-12.5 0 6.25 6.25 0 0 1 12.5 0Z" /></svg>
      </span>
      <input
        v-model="query"
        class="input pl-9"
        placeholder="Search a product by name or SKU to add…"
        @focus="open = true"
        @blur="close"
      />
    </div>

    <div
      v-if="open && filtered.length"
      class="absolute z-30 mt-1 max-h-64 w-full overflow-y-auto rounded-xl border border-slate-200 bg-white p-1 shadow-xl dark:border-slate-600 dark:bg-slate-700"
    >
      <button
        v-for="p in filtered"
        :key="p.id"
        type="button"
        class="flex w-full items-center justify-between gap-3 rounded-lg px-3 py-2 text-left text-sm hover:bg-slate-100 dark:hover:bg-slate-600"
        @mousedown.prevent="pick(p)"
      >
        <span class="font-medium">{{ p.name }}</span>
        <span class="shrink-0 text-xs text-slate-400">{{ p.sku }} · stock {{ p.quantity }}</span>
      </button>
    </div>
  </div>
</template>
