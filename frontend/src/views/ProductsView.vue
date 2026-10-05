<script setup>
import { ref, computed, onMounted } from 'vue'
import api, { assetUrl } from '../lib/api'
import CrudPage from '../components/CrudPage.vue'

const money = (n) => '৳' + Number(n || 0).toLocaleString('en-IN')

const categoryOptions = ref([])
const supplierOptions = ref([])

// Quick "low stock only" filter, sent to the API as ?low_stock=true.
const lowStock = ref(false)
const extraParams = computed(() => (lowStock.value ? { low_stock: true } : {}))

const stockBadge = (r) => {
  const low = r.quantity <= 10
  const cls = low
    ? 'bg-red-100 text-red-700 dark:bg-red-500/20 dark:text-red-300'
    : 'bg-slate-100 text-slate-600 dark:bg-slate-700 dark:text-slate-300'
  return `<span class="badge ${cls}">${r.quantity} ${r.unit || ''}</span>`
}

// Gross margin on the selling price: (price - cost) / price. Green = healthy,
// amber = thin (<15%), red = selling at or below cost (a loss).
const marginBadge = (r) => {
  const price = Number(r.price) || 0
  const cost = Number(r.cost_price) || 0
  if (price <= 0) return '<span class="text-slate-400">—</span>'
  const pct = ((price - cost) / price) * 100
  const cls =
    pct <= 0
      ? 'bg-red-100 text-red-700 dark:bg-red-500/20 dark:text-red-300'
      : pct < 15
        ? 'bg-amber-100 text-amber-700 dark:bg-amber-500/20 dark:text-amber-300'
        : 'bg-emerald-100 text-emerald-700 dark:bg-emerald-500/20 dark:text-emerald-300'
  return `<span class="badge ${cls}">${pct.toFixed(1)}%</span>`
}

const imgCell = (r) =>
  r.image
    ? `<img src="${assetUrl(r.image)}" class="h-10 w-10 rounded-lg object-cover border border-slate-200" />`
    : `<div class="h-10 w-10 rounded-lg bg-slate-100 grid place-items-center text-[9px] text-slate-400 dark:bg-slate-700">IMG</div>`

const columns = [
  { key: 'image', label: '', render: imgCell },
  { key: 'name', label: 'Name' },
  { key: 'sku', label: 'SKU' },
  { key: 'category', label: 'Category', render: (r) => r.category?.name || '—' },
  { key: 'price', label: 'Price', render: (r) => money(r.price) },
  { key: 'margin', label: 'Margin', render: marginBadge },
  { key: 'quantity', label: 'Stock', render: stockBadge },
]

// fields is computed so the select options appear once categories/suppliers load.
const fields = computed(() => [
  { key: 'image', label: 'Product Image', type: 'image' },
  { key: 'name', label: 'Name', type: 'text', required: true },
  { key: 'sku', label: 'SKU', type: 'text', required: true },
  { key: 'category_id', label: 'Category', type: 'select', options: categoryOptions.value },
  { key: 'supplier_id', label: 'Supplier', type: 'select', options: supplierOptions.value },
  { key: 'price', label: 'Selling Price', type: 'number' },
  { key: 'cost_price', label: 'Cost Price', type: 'number' },
  { key: 'quantity', label: 'Opening Quantity', type: 'number' },
  { key: 'unit', label: 'Unit (pcs, kg...)', type: 'text' },
  { key: 'is_active', label: 'Status', type: 'checkbox' },
])

const newItem = () => ({
  image: '',
  name: '',
  sku: '',
  category_id: categoryOptions.value[0]?.value || '',
  supplier_id: supplierOptions.value[0]?.value || '',
  price: 0,
  cost_price: 0,
  quantity: 0,
  unit: 'pcs',
  is_active: true,
})

onMounted(async () => {
  const [cats, sups] = await Promise.all([
    api.get('/categories', { params: { per_page: 100 } }),
    api.get('/suppliers', { params: { per_page: 100 } }),
  ])
  categoryOptions.value = cats.data.data.map((c) => ({ value: c.id, label: c.name }))
  supplierOptions.value = sups.data.data.map((s) => ({ value: s.id, label: s.name }))
})
</script>

<template>
  <CrudPage title="Products" endpoint="/products" export-endpoint="/products/export" :columns="columns" :fields="fields" :new-item="newItem" :extra-params="extraParams">
    <template #filters>
      <label class="flex cursor-pointer items-center gap-2 whitespace-nowrap rounded-lg border border-slate-200 px-3 py-2 text-sm text-slate-600 transition hover:bg-slate-50 dark:border-slate-600 dark:text-slate-300 dark:hover:bg-slate-700/40">
        <input v-model="lowStock" type="checkbox" class="h-4 w-4 rounded text-brand-600" />
        Low stock only
      </label>
    </template>
  </CrudPage>
</template>
