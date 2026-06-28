<script setup>
import { ref, computed } from 'vue'
import { assetUrl } from '../lib/api'

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
  return list.slice(0, 12)
})

function pick(p) {
  emit('select', p)
  query.value = ''
  // Keep the list open + input focused so several products can be added fast.
  open.value = true
}

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
        @input="open = true"
        @blur="close"
      />
    </div>

    <div
      v-if="open && filtered.length"
      class="absolute z-30 mt-1 max-h-72 w-full overflow-y-auto rounded-xl border border-slate-200 bg-white p-1 shadow-xl dark:border-slate-600 dark:bg-slate-700"
    >
      <button
        v-for="p in filtered"
        :key="p.id"
        type="button"
        class="flex w-full items-center gap-3 rounded-lg px-2 py-1.5 text-left text-sm hover:bg-slate-100 dark:hover:bg-slate-600"
        @mousedown.prevent="pick(p)"
      >
        <img v-if="p.image" :src="assetUrl(p.image)" class="h-9 w-9 shrink-0 rounded-lg border border-slate-200 object-cover dark:border-slate-600" />
        <span v-else class="grid h-9 w-9 shrink-0 place-items-center rounded-lg bg-slate-100 text-[10px] text-slate-400 dark:bg-slate-600">IMG</span>
        <span class="min-w-0 flex-1">
          <span class="block truncate font-medium">{{ p.name }}</span>
          <span class="block truncate text-xs text-slate-400">{{ p.sku }}</span>
        </span>
        <span class="shrink-0 text-xs text-slate-400">stock {{ p.quantity }}</span>
      </button>
    </div>
  </div>
</template>
