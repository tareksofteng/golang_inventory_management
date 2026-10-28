<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import api, { assetUrl } from '../lib/api'
import CrudPage from '../components/CrudPage.vue'
import Modal from '../components/Modal.vue'

const money = (n) => '৳' + Number(n || 0).toLocaleString('en-IN')

const categoryOptions = ref([])
const supplierOptions = ref([])

// Quick "low stock only" filter, sent to the API as ?low_stock=true.
const lowStock = ref(false)
const extraParams = computed(() => (lowStock.value ? { low_stock: true } : {}))

const stockBadge = (r) => {
  // "Low" is measured against each product's own reorder level (default 10).
  const reorder = r.reorder_level ?? 10
  const low = r.quantity <= reorder
  const cls = low
    ? 'bg-red-100 text-red-700 dark:bg-red-500/20 dark:text-red-300'
    : 'bg-slate-100 text-slate-600 dark:bg-slate-700 dark:text-slate-300'
  const title = low ? ` title="At or below reorder level (${reorder})"` : ''
  return `<span class="badge ${cls}"${title}>${r.quantity} ${r.unit || ''}</span>`
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
  { key: 'reorder_level', label: 'Reorder Level (low-stock alert)', type: 'number' },
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
  reorder_level: 10,
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

// ---- Stock adjustment -----------------------------------------------------
const crud = ref(null) // CrudPage instance, so we can refresh the list after an adjustment
const REASONS = ['damaged', 'lost', 'recount', 'found', 'other']
const adjustProduct = ref(null)
const adjustError = ref('')
const adjustSaving = ref(false)
const adjustForm = reactive({ direction: 'remove', quantity: 1, reason: 'damaged', note: '' })
const history = ref([])
const historyLoading = ref(false)
const fmtDate = (d) => new Date(d).toLocaleDateString()

async function loadHistory(productId) {
  historyLoading.value = true
  history.value = []
  try {
    const { data } = await api.get(`/products/${productId}/stock-history`, { params: { per_page: 5 } })
    history.value = data.data
  } finally {
    historyLoading.value = false
  }
}

function openAdjust(row) {
  adjustProduct.value = row
  adjustError.value = ''
  adjustForm.direction = 'remove'
  adjustForm.quantity = 1
  adjustForm.reason = 'damaged'
  adjustForm.note = ''
  loadHistory(row.id)
}

// Preview the resulting stock so the user sees the effect before saving.
const adjustResult = computed(() => {
  if (!adjustProduct.value) return 0
  const delta = (adjustForm.direction === 'add' ? 1 : -1) * Math.abs(Number(adjustForm.quantity) || 0)
  return adjustProduct.value.quantity + delta
})

async function submitAdjust() {
  const qty = Math.abs(Number(adjustForm.quantity) || 0)
  if (qty <= 0) {
    adjustError.value = 'Enter a quantity greater than zero'
    return
  }
  adjustSaving.value = true
  adjustError.value = ''
  try {
    const change = (adjustForm.direction === 'add' ? 1 : -1) * qty
    await api.post(`/products/${adjustProduct.value.id}/adjust-stock`, {
      change,
      reason: adjustForm.reason,
      note: adjustForm.note,
    })
    adjustProduct.value = null
    crud.value?.load()
  } catch (e) {
    adjustError.value = e.response?.data?.message || 'Failed to adjust stock'
  } finally {
    adjustSaving.value = false
  }
}
</script>

<template>
  <CrudPage ref="crud" title="Products" endpoint="/products" export-endpoint="/products/export" :columns="columns" :fields="fields" :new-item="newItem" :extra-params="extraParams">
    <template #filters>
      <label class="flex cursor-pointer items-center gap-2 whitespace-nowrap rounded-lg border border-slate-200 px-3 py-2 text-sm text-slate-600 transition hover:bg-slate-50 dark:border-slate-600 dark:text-slate-300 dark:hover:bg-slate-700/40">
        <input v-model="lowStock" type="checkbox" class="h-4 w-4 rounded text-brand-600" />
        Low stock only
      </label>
    </template>
    <template #row-actions="{ row }">
      <button class="btn-ghost !px-2 !py-1 text-xs" @click="openAdjust(row)">Adjust</button>
    </template>
  </CrudPage>

  <!-- Manual stock adjustment -->
  <Modal v-if="adjustProduct" :title="`Adjust Stock — ${adjustProduct.name}`" @close="adjustProduct = null">
    <form class="space-y-4" @submit.prevent="submitAdjust">
      <div v-if="adjustError" class="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600 dark:bg-red-500/10">{{ adjustError }}</div>

      <div class="rounded-lg bg-slate-50 px-3 py-2 text-sm text-slate-500 dark:bg-slate-700/40 dark:text-slate-300">
        Current stock: <span class="font-semibold text-slate-700 dark:text-slate-100">{{ adjustProduct.quantity }} {{ adjustProduct.unit }}</span>
      </div>

      <div class="grid grid-cols-2 gap-3">
        <div>
          <label class="label">Action</label>
          <select v-model="adjustForm.direction" class="input">
            <option value="remove">Remove (−)</option>
            <option value="add">Add (+)</option>
          </select>
        </div>
        <div>
          <label class="label">Quantity</label>
          <input v-model="adjustForm.quantity" type="number" min="1" class="input" />
        </div>
      </div>

      <div>
        <label class="label">Reason</label>
        <select v-model="adjustForm.reason" class="input">
          <option v-for="r in REASONS" :key="r" :value="r" class="capitalize">{{ r }}</option>
        </select>
      </div>

      <div>
        <label class="label">Note (optional)</label>
        <input v-model="adjustForm.note" class="input" placeholder="e.g. broken during handling" />
      </div>

      <div class="rounded-lg px-3 py-2 text-sm" :class="adjustResult < 0 ? 'bg-red-50 text-red-600 dark:bg-red-500/10' : 'bg-emerald-50 text-emerald-700 dark:bg-emerald-500/10 dark:text-emerald-300'">
        New stock will be: <span class="font-semibold">{{ adjustResult }} {{ adjustProduct.unit }}</span>
        <span v-if="adjustResult < 0"> — can't go below zero</span>
      </div>

      <!-- Recent adjustment history for this product -->
      <div class="border-t border-slate-100 pt-3 dark:border-slate-700">
        <div class="mb-2 text-xs font-semibold uppercase tracking-wide text-slate-400">Recent adjustments</div>
        <p v-if="historyLoading" class="text-sm text-slate-400">Loading…</p>
        <p v-else-if="!history.length" class="text-sm text-slate-400">No adjustments yet</p>
        <ul v-else class="space-y-1.5">
          <li v-for="h in history" :key="h.id" class="flex items-center justify-between gap-3 text-sm">
            <span class="flex items-center gap-2">
              <span class="font-semibold" :class="h.change < 0 ? 'text-red-600' : 'text-emerald-600'">{{ h.change > 0 ? '+' : '' }}{{ h.change }}</span>
              <span class="capitalize text-slate-500 dark:text-slate-300">{{ h.reason }}</span>
            </span>
            <span class="text-xs text-slate-400">→ {{ h.new_quantity }} · {{ fmtDate(h.created_at) }}</span>
          </li>
        </ul>
      </div>

      <div class="flex justify-end gap-2 pt-1">
        <button type="button" class="btn-ghost" @click="adjustProduct = null">Cancel</button>
        <button type="submit" class="btn-primary" :disabled="adjustSaving || adjustResult < 0">{{ adjustSaving ? 'Saving…' : 'Save adjustment' }}</button>
      </div>
    </form>
  </Modal>
</template>
