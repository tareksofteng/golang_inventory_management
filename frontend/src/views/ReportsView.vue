<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import api, { assetUrl } from '../lib/api'

const route = useRoute()
const money = (n) => '৳' + Number(n || 0).toLocaleString('en-IN')

// Each report is its own route: /reports/:type
const REPORTS = {
  sales: { label: 'Sales Report', endpoint: '/reports/sales', ranged: true },
  purchases: { label: 'Purchase Report', endpoint: '/reports/purchases', ranged: true },
  'customer-due': { label: 'Customer Due Report', endpoint: '/reports/customer-due', ranged: false },
  'supplier-due': { label: 'Supplier Due Report', endpoint: '/reports/supplier-due', ranged: false },
  stock: { label: 'Current Stock Report', endpoint: '/reports/stock', ranged: false },
}
const type = computed(() => (REPORTS[route.params.type] ? route.params.type : 'sales'))
const active = computed(() => REPORTS[type.value])

const data = ref(null)
const loading = ref(false)

const today = new Date()
const pad = (n) => String(n).padStart(2, '0')
const from = ref(`${today.getFullYear()}-${pad(today.getMonth() + 1)}-01`)
const to = ref(`${today.getFullYear()}-${pad(today.getMonth() + 1)}-${pad(today.getDate())}`)

// Stock report: category filter.
const categories = ref([])
const categoryId = ref('')

async function load() {
  loading.value = true
  data.value = null
  try {
    const params = active.value.ranged ? { from: from.value, to: to.value } : {}
    if (type.value === 'stock' && categoryId.value) params.category_id = categoryId.value
    const res = await api.get(active.value.endpoint, { params })
    data.value = res.data.data
  } finally {
    loading.value = false
  }
}

watch(type, load)
watch(categoryId, () => {
  if (type.value === 'stock') load()
})

const fmtDate = (d) => new Date(d).toLocaleDateString()
const printPage = () => window.print()

onMounted(async () => {
  try {
    const res = await api.get('/categories', { params: { per_page: 100 } })
    categories.value = res.data.data
  } catch (e) {
    /* categories are only needed for the stock filter */
  }
  load()
})
</script>

<template>
  <div>
    <div class="mb-6 flex flex-wrap items-center justify-between gap-3">
      <h1 class="text-2xl font-bold">{{ active.label }}</h1>
      <button class="btn-ghost print:hidden" @click="printPage">🖨️ Print</button>
    </div>

    <!-- Date range (sales/purchase only) -->
    <div v-if="active.ranged" class="card mb-4 flex flex-wrap items-end gap-3 p-4 print:hidden">
      <div>
        <label class="label">From</label>
        <input v-model="from" type="date" class="input" />
      </div>
      <div>
        <label class="label">To</label>
        <input v-model="to" type="date" class="input" />
      </div>
      <button class="btn-primary" @click="load">Apply</button>
    </div>

    <!-- Category filter (stock report only) -->
    <div v-if="type === 'stock'" class="card mb-4 flex flex-wrap items-center gap-3 p-4 print:hidden">
      <span class="text-sm font-medium text-slate-600 dark:text-slate-300">Filter by Category</span>
      <select v-model="categoryId" class="input max-w-xs">
        <option value="">All Categories</option>
        <option v-for="c in categories" :key="c.id" :value="c.id">{{ c.name }}</option>
      </select>
      <span v-if="data" class="ml-auto rounded-full bg-brand-50 px-3 py-1 text-xs font-medium text-brand-700 dark:bg-brand-600/20 dark:text-brand-200">
        {{ data.items.length }} product(s)
      </span>
    </div>

    <div v-if="loading" class="py-16 text-center text-slate-400">Loading…</div>

    <div v-else-if="data" class="space-y-4">
      <!-- Summary for ranged reports -->
      <div v-if="active.ranged" class="grid grid-cols-2 gap-4 lg:grid-cols-4">
        <div class="card p-4"><div class="text-sm text-slate-500">Invoices</div><div class="text-xl font-bold">{{ data.summary.count }}</div></div>
        <div class="card p-4"><div class="text-sm text-slate-500">Total</div><div class="text-xl font-bold">{{ money(data.summary.total) }}</div></div>
        <div class="card p-4"><div class="text-sm text-slate-500">Paid</div><div class="text-xl font-bold text-emerald-600">{{ money(data.summary.paid) }}</div></div>
        <div class="card p-4"><div class="text-sm text-slate-500">Due</div><div class="text-xl font-bold text-amber-600">{{ money(data.summary.due) }}</div></div>
      </div>

      <!-- Due / stock totals -->
      <div v-if="type === 'customer-due' || type === 'supplier-due'" class="card p-4">
        <span class="text-sm text-slate-500">Total Outstanding: </span>
        <span class="text-xl font-bold text-amber-600">{{ money(data.total_due) }}</span>
      </div>
      <div v-if="type === 'stock'" class="card p-4">
        <span class="text-sm text-slate-500">Total Stock Value: </span>
        <span class="text-xl font-bold text-brand-600">{{ money(data.total_value) }}</span>
      </div>

      <!-- Tables -->
      <div class="card overflow-hidden">
        <div class="overflow-x-auto">
          <!-- Sales -->
          <table v-if="type === 'sales'" class="w-full text-left text-sm">
            <thead class="bg-slate-50 text-xs uppercase text-slate-500 dark:bg-slate-700/40">
              <tr><th class="px-4 py-3">Invoice</th><th class="px-4 py-3">Customer</th><th class="px-4 py-3">Date</th><th class="px-4 py-3">Total</th><th class="px-4 py-3">Paid</th><th class="px-4 py-3">Due</th></tr>
            </thead>
            <tbody class="divide-y divide-slate-100 dark:divide-slate-700">
              <tr v-if="!data.items.length"><td colspan="6" class="px-4 py-8 text-center text-slate-400">No records</td></tr>
              <tr v-for="s in data.items" :key="s.id">
                <td class="px-4 py-3 font-medium">{{ s.invoice_no }}</td><td class="px-4 py-3">{{ s.customer?.name }}</td>
                <td class="px-4 py-3 text-slate-400">{{ fmtDate(s.created_at) }}</td><td class="px-4 py-3">{{ money(s.total_amount) }}</td>
                <td class="px-4 py-3 text-emerald-600">{{ money(s.paid_amount) }}</td><td class="px-4 py-3 text-amber-600">{{ money(s.due) }}</td>
              </tr>
            </tbody>
          </table>

          <!-- Purchases -->
          <table v-else-if="type === 'purchases'" class="w-full text-left text-sm">
            <thead class="bg-slate-50 text-xs uppercase text-slate-500 dark:bg-slate-700/40">
              <tr><th class="px-4 py-3">Invoice</th><th class="px-4 py-3">Supplier</th><th class="px-4 py-3">Date</th><th class="px-4 py-3">Total</th><th class="px-4 py-3">Paid</th><th class="px-4 py-3">Due</th></tr>
            </thead>
            <tbody class="divide-y divide-slate-100 dark:divide-slate-700">
              <tr v-if="!data.items.length"><td colspan="6" class="px-4 py-8 text-center text-slate-400">No records</td></tr>
              <tr v-for="p in data.items" :key="p.id">
                <td class="px-4 py-3 font-medium">{{ p.invoice_no }}</td><td class="px-4 py-3">{{ p.supplier?.name }}</td>
                <td class="px-4 py-3 text-slate-400">{{ fmtDate(p.created_at) }}</td><td class="px-4 py-3">{{ money(p.total_amount) }}</td>
                <td class="px-4 py-3 text-emerald-600">{{ money(p.paid_amount) }}</td><td class="px-4 py-3 text-amber-600">{{ money(p.due) }}</td>
              </tr>
            </tbody>
          </table>

          <!-- Customer / Supplier due -->
          <table v-else-if="type === 'customer-due' || type === 'supplier-due'" class="w-full text-left text-sm">
            <thead class="bg-slate-50 text-xs uppercase text-slate-500 dark:bg-slate-700/40">
              <tr><th class="px-4 py-3">Name</th><th class="px-4 py-3">Phone</th><th class="px-4 py-3">Due</th></tr>
            </thead>
            <tbody class="divide-y divide-slate-100 dark:divide-slate-700">
              <tr v-for="row in (data.customers || data.suppliers)" :key="row.id">
                <td class="px-4 py-3 font-medium">{{ row.name }}</td><td class="px-4 py-3 text-slate-400">{{ row.phone }}</td>
                <td class="px-4 py-3 font-semibold text-amber-600">{{ money(row.due) }}</td>
              </tr>
              <tr v-if="!(data.customers || data.suppliers || []).length"><td colspan="3" class="px-4 py-8 text-center text-slate-400">No outstanding dues 🎉</td></tr>
            </tbody>
          </table>

          <!-- Stock -->
          <table v-else-if="type === 'stock'" class="w-full text-left text-sm">
            <thead class="bg-slate-50 text-xs uppercase text-slate-500 dark:bg-slate-700/40">
              <tr><th class="px-4 py-3">Product</th><th class="px-4 py-3">SKU</th><th class="px-4 py-3">Category</th><th class="px-4 py-3">Qty</th><th class="px-4 py-3">Cost</th><th class="px-4 py-3">Value</th></tr>
            </thead>
            <tbody class="divide-y divide-slate-100 dark:divide-slate-700">
              <tr v-for="p in data.items" :key="p.id">
                <td class="px-4 py-3">
                  <div class="flex items-center gap-2">
                    <img v-if="p.image" :src="assetUrl(p.image)" class="h-8 w-8 rounded border border-slate-200 object-cover" />
                    <span v-else class="grid h-8 w-8 place-items-center rounded bg-slate-100 text-[9px] text-slate-400 dark:bg-slate-700">IMG</span>
                    <span class="font-medium">{{ p.name }}</span>
                  </div>
                </td>
                <td class="px-4 py-3 text-slate-400">{{ p.sku }}</td>
                <td class="px-4 py-3">{{ p.category?.name }}</td>
                <td class="px-4 py-3"><span :class="p.quantity <= 10 ? 'font-semibold text-red-600' : ''">{{ p.quantity }}</span></td>
                <td class="px-4 py-3">{{ money(p.cost_price) }}</td><td class="px-4 py-3 font-medium">{{ money(p.quantity * p.cost_price) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </div>
</template>
