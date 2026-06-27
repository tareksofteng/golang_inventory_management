<script setup>
import { ref, computed, onMounted } from 'vue'
import api from '../lib/api'
import KpiCard from '../components/KpiCard.vue'

const data = ref(null)
const loading = ref(true)

const money = (n) => '৳' + Number(n || 0).toLocaleString('en-IN')

const maxCat = computed(() =>
  Math.max(1, ...(data.value?.products_by_category || []).map((c) => c.count)),
)
const maxTrend = computed(() =>
  Math.max(1, ...(data.value?.sales_trend || []).map((d) => d.total)),
)
const dayLabel = (d) =>
  new Date(d + 'T00:00:00').toLocaleDateString('en-US', { month: 'short', day: 'numeric' })

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

    <!-- Money KPIs -->
    <div class="grid grid-cols-2 gap-4 lg:grid-cols-4">
      <KpiCard label="Total Sales" :value="money(data.finance.total_sales)" icon="🛒" accent="bg-emerald-600" />
      <KpiCard label="Total Purchase" :value="money(data.finance.total_purchase)" icon="🧾" accent="bg-blue-600" />
      <KpiCard label="Receivable (Customer Due)" :value="money(data.finance.receivable)" icon="📥" accent="bg-amber-600" />
      <KpiCard label="Payable (Supplier Due)" :value="money(data.finance.payable)" icon="📤" accent="bg-red-600" />
    </div>

    <!-- Secondary KPIs -->
    <div class="grid grid-cols-2 gap-4 lg:grid-cols-4">
      <KpiCard label="Today's Sales" :value="money(data.finance.today_sales)" icon="📅" accent="bg-teal-600" />
      <KpiCard label="Monthly Revenue" :value="money(data.finance.month_sales)" icon="💹" accent="bg-violet-600" />
      <KpiCard label="Stock Value" :value="money(data.stock_value)" icon="💰" accent="bg-brand-600" />
      <KpiCard label="Low Stock Items" :value="data.low_stock_count" icon="⚠️" accent="bg-orange-600" />
    </div>

    <!-- Count chips -->
    <div class="grid grid-cols-2 gap-4 lg:grid-cols-4">
      <KpiCard label="Products" :value="data.totals.products" icon="📦" accent="bg-slate-500" />
      <KpiCard label="Categories" :value="data.totals.categories" icon="🏷️" accent="bg-slate-500" />
      <KpiCard label="Suppliers" :value="data.totals.suppliers" icon="🚚" accent="bg-slate-500" />
      <KpiCard label="Customers" :value="data.totals.customers" icon="👤" accent="bg-slate-500" />
    </div>

    <div class="grid gap-6 lg:grid-cols-2">
      <!-- 7-day sales trend (vertical bars) -->
      <div class="card p-6">
        <h2 class="mb-4 font-semibold">Sales — Last 7 Days</h2>
        <div class="flex h-44 items-end justify-between gap-2">
          <div v-for="d in data.sales_trend" :key="d.date" class="flex h-full flex-1 flex-col items-center justify-end gap-1">
            <span v-if="d.total > 0" class="text-[10px] font-medium text-slate-500">{{ money(d.total) }}</span>
            <div
              class="w-full rounded-t bg-brand-600 transition-all"
              :style="{ height: Math.max(4, (d.total / maxTrend) * 150) + 'px' }"
              :title="money(d.total)"
            />
            <span class="text-[10px] text-slate-400">{{ dayLabel(d.date) }}</span>
          </div>
        </div>
      </div>

      <!-- Top selling products -->
      <div class="card p-6">
        <h2 class="mb-4 font-semibold">🏆 Top Selling Products</h2>
        <div v-if="!data.top_selling_products.length" class="text-sm text-slate-400">No sales yet</div>
        <ul class="divide-y divide-slate-100 dark:divide-slate-700">
          <li v-for="(p, i) in data.top_selling_products" :key="p.product_id" class="flex items-center justify-between py-2.5 text-sm">
            <div class="flex items-center gap-3">
              <span class="grid h-6 w-6 place-items-center rounded-full bg-brand-100 text-xs font-bold text-brand-700 dark:bg-brand-600/20 dark:text-brand-200">{{ i + 1 }}</span>
              <span class="font-medium">{{ p.name }}</span>
            </div>
            <div class="text-right">
              <div class="font-semibold">{{ money(p.revenue) }}</div>
              <div class="text-xs text-slate-400">{{ p.quantity_sold }} sold</div>
            </div>
          </li>
        </ul>
      </div>
    </div>

    <div class="grid gap-6 lg:grid-cols-2">
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

      <!-- Recent sales -->
      <div class="card p-6">
        <h2 class="mb-4 font-semibold">🧾 Recent Sales</h2>
        <div v-if="!data.recent_sales.length" class="text-sm text-slate-400">No sales yet</div>
        <ul class="divide-y divide-slate-100 dark:divide-slate-700">
          <li v-for="s in data.recent_sales" :key="s.id" class="flex items-center justify-between py-2.5 text-sm">
            <div>
              <div class="font-medium">{{ s.invoice_no }}</div>
              <div class="text-xs text-slate-400">{{ s.customer?.name }}</div>
            </div>
            <div class="text-right">
              <div class="font-semibold">{{ money(s.total_amount) }}</div>
              <div class="text-xs" :class="s.due > 0 ? 'text-amber-500' : 'text-emerald-500'">
                {{ s.due > 0 ? money(s.due) + ' due' : 'Paid' }}
              </div>
            </div>
          </li>
        </ul>
      </div>
    </div>
  </div>
</template>
