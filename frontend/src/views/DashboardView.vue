<script setup>
import { ref, computed, onMounted } from 'vue'
import api from '../lib/api'
import KpiCard from '../components/KpiCard.vue'
import Icon from '../components/Icon.vue'
import SalesTrendChart from '../components/SalesTrendChart.vue'

const data = ref(null)
const loading = ref(true)

const money = (n) => '৳' + Number(n || 0).toLocaleString('en-IN')

const todayLabel = new Date().toLocaleDateString('en-US', {
  weekday: 'long', day: 'numeric', month: 'long', year: 'numeric',
})

// Top sellers are a magnitude comparison — one hue, bar length carries the value.
const maxRevenue = computed(() =>
  Math.max(1, ...(data.value?.top_selling_products || []).map((p) => p.revenue)),
)

const counts = computed(() => [
  { label: 'Products', value: data.value?.totals.products, icon: 'cube' },
  { label: 'Categories', value: data.value?.totals.categories, icon: 'tag' },
  { label: 'Suppliers', value: data.value?.totals.suppliers, icon: 'truck' },
  { label: 'Customers', value: data.value?.totals.customers, icon: 'user' },
])

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

  <div v-else-if="data" class="space-y-5">
    <header>
      <h1 class="text-2xl font-bold tracking-tight">Dashboard</h1>
      <p class="mt-0.5 text-sm text-slate-500 dark:text-slate-400">{{ todayLabel }}</p>
    </header>

    <!-- Primary money KPIs -->
    <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
      <KpiCard label="Total sales" :value="money(data.finance.total_sales)" icon="cart" tone="emerald" />
      <KpiCard label="Total purchase" :value="money(data.finance.total_purchase)" icon="bag" tone="blue" />
      <KpiCard label="Receivable" :value="money(data.finance.receivable)" icon="inbound" tone="amber" hint="Due from customers" />
      <KpiCard label="Payable" :value="money(data.finance.payable)" icon="outbound" tone="red" hint="Due to suppliers" />
    </div>

    <!-- Secondary metrics — deliberately lighter than the row above -->
    <div class="grid grid-cols-2 gap-4 xl:grid-cols-4">
      <KpiCard compact label="Today's sales" :value="money(data.finance.today_sales)" icon="calendar" tone="slate" />
      <KpiCard compact label="This month" :value="money(data.finance.month_sales)" icon="trending" tone="slate" />
      <KpiCard compact label="Stock value" :value="money(data.stock_value)" icon="wallet" tone="slate" />
      <KpiCard
        compact label="Low stock items" :value="data.low_stock_count" icon="alert"
        :tone="data.low_stock_count > 0 ? 'amber' : 'slate'"
      />
    </div>

    <!-- Catalogue counts: least important data, least ink — one compact strip -->
    <div class="card grid grid-cols-2 divide-y divide-slate-100 sm:grid-cols-4 sm:divide-y-0 sm:divide-x dark:divide-slate-700">
      <div v-for="c in counts" :key="c.label" class="flex items-center gap-3 px-5 py-3.5">
        <span class="text-slate-400"><Icon :name="c.icon" :size="18" /></span>
        <div>
          <div class="text-lg font-semibold leading-tight">{{ c.value }}</div>
          <div class="text-xs text-slate-500 dark:text-slate-400">{{ c.label }}</div>
        </div>
      </div>
    </div>

    <div class="grid gap-5 xl:grid-cols-5">
      <!-- Sales trend -->
      <section class="card p-5 xl:col-span-3">
        <div class="mb-1 flex items-baseline justify-between">
          <h2 class="font-semibold">Sales</h2>
          <span class="text-xs text-slate-400">Last 7 days</span>
        </div>
        <SalesTrendChart :points="data.sales_trend" />
      </section>

      <!-- Top selling: ranked magnitude bars -->
      <section class="card p-5 xl:col-span-2">
        <h2 class="mb-4 flex items-center gap-2 font-semibold">
          <span class="text-slate-400"><Icon name="trophy" :size="18" /></span> Top selling products
        </h2>
        <p v-if="!data.top_selling_products.length" class="text-sm text-slate-400">No sales yet</p>
        <ul v-else class="space-y-3.5">
          <li v-for="(p, i) in data.top_selling_products" :key="p.product_id">
            <div class="flex items-baseline justify-between gap-3 text-sm">
              <span class="flex min-w-0 items-baseline gap-2">
                <span class="w-4 shrink-0 text-xs font-medium text-slate-400">{{ i + 1 }}</span>
                <span class="truncate font-medium">{{ p.name }}</span>
              </span>
              <span class="shrink-0 font-semibold">{{ money(p.revenue) }}</span>
            </div>
            <div class="mt-1.5 flex items-center gap-2 pl-6">
              <div class="h-2 flex-1 overflow-hidden rounded-full bg-slate-100/80 dark:bg-slate-700/60">
                <div
                  class="h-full rounded-full bg-brand-600 dark:bg-brand-500"
                  :style="{ width: Math.max(2, (p.revenue / maxRevenue) * 100) + '%' }"
                />
              </div>
              <span class="shrink-0 text-xs text-slate-400">{{ p.quantity_sold }} sold</span>
            </div>
          </li>
        </ul>
      </section>
    </div>

    <div class="grid items-start gap-5 lg:grid-cols-2">
      <!-- Low stock: icon + label, never colour alone -->
      <section class="card p-5">
        <h2 class="mb-4 flex items-center gap-2 font-semibold">
          <span class="text-amber-500"><Icon name="alert" :size="18" /></span> Low stock alerts
        </h2>
        <div v-if="!data.low_stock_products.length" class="flex items-center gap-2 text-sm text-emerald-600 dark:text-emerald-400">
          <Icon name="check" :size="16" /> All products are well stocked
        </div>
        <ul v-else class="divide-y divide-slate-100 dark:divide-slate-700">
          <li v-for="p in data.low_stock_products" :key="p.id" class="flex items-center justify-between gap-3 py-2.5 text-sm">
            <div class="min-w-0">
              <div class="truncate font-medium">{{ p.name }}</div>
              <div class="truncate text-xs text-slate-400">{{ p.sku }} · {{ p.category?.name }}</div>
            </div>
            <span class="badge shrink-0 items-center gap-1 bg-amber-100 text-amber-700 dark:bg-amber-500/20 dark:text-amber-300">
              <Icon name="alert" :size="12" /> {{ p.quantity }} {{ p.unit }}
            </span>
          </li>
        </ul>
      </section>

      <!-- Recent sales -->
      <section class="card p-5">
        <h2 class="mb-4 flex items-center gap-2 font-semibold">
          <span class="text-slate-400"><Icon name="receipt" :size="18" /></span> Recent sales
        </h2>
        <p v-if="!data.recent_sales.length" class="text-sm text-slate-400">No sales yet</p>
        <ul v-else class="divide-y divide-slate-100 dark:divide-slate-700">
          <li v-for="s in data.recent_sales" :key="s.id" class="flex items-center justify-between gap-3 py-2.5 text-sm">
            <div class="min-w-0">
              <div class="truncate font-medium">{{ s.invoice_no }}</div>
              <div class="truncate text-xs text-slate-400">{{ s.customer?.name }}</div>
            </div>
            <div class="shrink-0 text-right">
              <div class="font-semibold">{{ money(s.total_amount) }}</div>
              <div class="text-xs" :class="s.due > 0 ? 'text-amber-600 dark:text-amber-400' : 'text-emerald-600 dark:text-emerald-400'">
                {{ s.due > 0 ? money(s.due) + ' due' : 'Paid' }}
              </div>
            </div>
          </li>
        </ul>
      </section>
    </div>
  </div>
</template>
