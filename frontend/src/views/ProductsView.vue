<script setup>
import { ref, computed, onMounted } from 'vue'
import api from '../lib/api'
import CrudPage from '../components/CrudPage.vue'

const money = (n) => '৳' + Number(n || 0).toLocaleString('en-IN')

const categoryOptions = ref([])
const supplierOptions = ref([])

const stockBadge = (r) => {
  const low = r.quantity <= 10
  const cls = low
    ? 'bg-red-100 text-red-700 dark:bg-red-500/20 dark:text-red-300'
    : 'bg-slate-100 text-slate-600 dark:bg-slate-700 dark:text-slate-300'
  return `<span class="badge ${cls}">${r.quantity} ${r.unit || ''}</span>`
}

const columns = [
  { key: 'name', label: 'Name' },
  { key: 'sku', label: 'SKU' },
  { key: 'category', label: 'Category', render: (r) => r.category?.name || '—' },
  { key: 'price', label: 'Price', render: (r) => money(r.price) },
  { key: 'quantity', label: 'Stock', render: stockBadge },
]

// fields is computed so the select options appear once categories/suppliers load.
const fields = computed(() => [
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
  <CrudPage title="Products" endpoint="/products" :columns="columns" :fields="fields" :new-item="newItem" />
</template>
