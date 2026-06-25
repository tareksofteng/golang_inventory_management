<script setup>
import { ref, computed, onMounted } from 'vue'
import api from '../lib/api'
import KpiCard from '../components/KpiCard.vue'

const data = ref(null)
const loading = ref(true)

const money = (n) => '৳' + Number(n || 0).toLocaleString('en-IN')

// Scale category bars relative to the largest count.
const maxCat = computed(() =>
  Math.max(1, ...(data.value?.products_by_category || []).map((c) => c.count)),
)

onMounted(async () => {
  try {
    const res = await api.get('/dashboard/summary')
    data.value = res.data.data
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div v-if="loading" class="py-20 text-center text-slate-400">Loading dashboard…</div>

  <div v-else-if="data" class="space-y-6">
    <h1 class="text-2xl font-bold">Dashboard</h1>

    <!-- KPI cards -->
    <div class="grid grid-cols-2 gap-4 lg:grid-cols-4">
      <KpiCard label="Products" :value="data.totals.products" icon="📦" accent="bg-blue-600" />
      <KpiCard label="Categories" :value="data.totals.categories" icon="🏷️" accent="bg-violet-600" />
      <KpiCard label="Suppliers" :value="data.totals.suppliers" icon="🚚" accent="bg-amber-600" />
      <KpiCard label="Customers" :value="data.totals.customers" icon="👤" accent="bg-emerald-600" />
      <KpiCard label="Stock Value" :value="money(data.stock_value)" icon="💰" accent="bg-brand-600" />
      <KpiCard label="Low Stock Items" :value="data.low_stock_count" icon="⚠️" accent="bg-red-600" />
    </div>

    <div class="grid gap-6 lg:grid-cols-2">
      <!-- Products by category (CSS bar chart) -->
      <div class="card p-6">
        <h2 class="mb-4 font-semibold">Products by Category</h2>
        <div v-if="!data.products_by_category.length" class="text-sm text-slate-400">No data</div>
        <div v-for="c in data.products_by_category" :key="c.category" class="mb-3">
          <div class="mb-1 flex justify-between text-sm">
            <span>{{ c.category }}</span><span class="text-slate-400">{{ c.count }}</span>
          </div>
          <div class="h-2.5 overflow-hidden rounded-full bg-slate-100 dark:bg-slate-700">
            <div class="h-full rounded-full bg-brand-600" :style="{ width: (c.count / maxCat) * 100 + '%' }" />
          </div>
        </div>
      </div>

      <!-- Low stock alerts -->
      <div class="card p-6">
        <h2 class="mb-4 flex items-center gap-2 font-semibold">⚠️ Low Stock Alerts</h2>
        <div v-if="!data.low_stock_products.length" class="text-sm text-slate-400">All products are well stocked 🎉</div>
        <ul class="divide-y divide-slate-100 dark:divide-slate-700">
          <li v-for="p in data.low_stock_products" :key="p.id" class="flex items-center justify-between py-2.5 text-sm">
            <div>
              <div class="font-medium">{{ p.name }}</div>
              <div class="text-xs text-slate-400">{{ p.sku }} · {{ p.category?.name }}</div>
            </div>
            <span class="badge bg-red-100 text-red-700 dark:bg-red-500/20 dark:text-red-300">{{ p.quantity }} {{ p.unit }}</span>
          </li>
        </ul>
      </div>
    </div>

    <!-- Recent products -->
    <div class="card overflow-hidden">
      <h2 class="border-b border-slate-200 px-6 py-4 font-semibold dark:border-slate-700">Recent Products</h2>
      <div class="overflow-x-auto">
        <table class="w-full text-left text-sm">
          <thead class="bg-slate-50 text-xs uppercase text-slate-500 dark:bg-slate-700/40">
            <tr>
              <th class="px-6 py-3">Name</th><th class="px-6 py-3">SKU</th>
              <th class="px-6 py-3">Category</th><th class="px-6 py-3">Price</th><th class="px-6 py-3">Qty</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 dark:divide-slate-700">
            <tr v-for="p in data.recent_products" :key="p.id" class="hover:bg-slate-50 dark:hover:bg-slate-700/30">
              <td class="px-6 py-3 font-medium">{{ p.name }}</td>
              <td class="px-6 py-3 text-slate-400">{{ p.sku }}</td>
              <td class="px-6 py-3">{{ p.category?.name }}</td>
              <td class="px-6 py-3">{{ money(p.price) }}</td>
              <td class="px-6 py-3">{{ p.quantity }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>
