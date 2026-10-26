<script setup>
import { ref, computed, onMounted } from 'vue'
import api from '../lib/api'
import KpiCard from '../components/KpiCard.vue'
import Icon from '../components/Icon.vue'
import SalesTrendChart from '../components/SalesTrendChart.vue'

const data = ref(null)
const loading = ref(true)

const money = (n) => '৳' + Number(n || 0).toLocaleString('en-IN')

const hour = new Date().getHours()
const greeting = hour < 12 ? 'Good morning' : hour < 17 ? 'Good afternoon' : 'Good evening'
const todayLabel = new Date().toLocaleDateString('en-US', {
  weekday: 'long', day: 'numeric', month: 'long', year: 'numeric',
})

// All-time profit margin — a derived signal the API doesn't send directly.
const profitMargin = computed(() => {
  const sales = data.value?.finance.total_sales || 0
  return sales > 0 ? (data.value.finance.total_profit / sales) * 100 : 0
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
  <!-- Skeleton while the summary loads -->
  <div v-if="loading" class="space-y-5">
    <div class="h-28 animate-pulse rounded-2xl bg-slate-200/70 dark:bg-slate-700/40" />
    <div class="grid grid-cols-2 gap-4 xl:grid-cols-4">
      <div v-for="n in 4" :key="n" class="h-24 animate-pulse rounded-xl bg-slate-200/70 dark:bg-slate-700/40" />
    </div>
    <div class="grid gap-5 xl:grid-cols-5">
      <div class="h-64 animate-pulse rounded-xl bg-slate-200/70 dark:bg-slate-700/40 xl:col-span-3" />
      <div class="h-64 animate-pulse rounded-xl bg-slate-200/70 dark:bg-slate-700/40 xl:col-span-2" />
    </div>
  </div>

  <div v-else-if="data" class="space-y-5">
    <header class="flex flex-wrap items-end justify-between gap-2">
      <div>
        <h1 class="text-2xl font-bold tracking-tight">{{ greeting }} 👋</h1>
        <p class="mt-0.5 text-sm text-slate-500 dark:text-slate-400">Here's your business overview · {{ todayLabel }}</p>
      </div>
    </header>

    <!-- Hero: the three numbers that matter most, all-time -->
    <section class="relative overflow-hidden rounded-2xl bg-gradient-to-br from-brand-600 to-indigo-700 p-6 text-white shadow-lg shadow-brand-600/20 sm:p-7">
      <div class="relative z-10 grid grid-cols-1 gap-6 sm:grid-cols-3 sm:divide-x sm:divide-white/15">
        <div class="sm:pr-6">
          <div class="flex items-center gap-2 text-sm font-medium text-white/70">
            <Icon name="cart" :size="16" /> Total sales
          </div>
          <div class="mt-1.5 text-3xl font-bold tracking-tight">{{ money(data.finance.total_sales) }}</div>
        </div>
        <div class="sm:px-6">
          <div class="flex items-center gap-2 text-sm font-medium text-white/70">
            <Icon name="trending" :size="16" /> Gross profit
          </div>
          <div class="mt-1.5 text-3xl font-bold tracking-tight">{{ money(data.finance.total_profit) }}</div>
        </div>
        <div class="sm:pl-6">
          <div class="flex items-center gap-2 text-sm font-medium text-white/70">
            <Icon name="wallet" :size="16" /> Profit margin
          </div>
          <div class="mt-1.5 text-3xl font-bold tracking-tight">
            {{ data.finance.total_sales > 0 ? profitMargin.toFixed(1) + '%' : '—' }}
          </div>
        </div>
      </div>
      <!-- Soft decorative bubbles -->
      <div class="pointer-events-none absolute -right-12 -top-12 h-44 w-44 rounded-full bg-white/10" />
      <div class="pointer-events-none absolute -bottom-20 right-28 h-52 w-52 rounded-full bg-white/5" />
    </section>

    <!-- Balances & stock -->
    <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
      <KpiCard label="Total purchase" :value="money(data.finance.total_purchase)" icon="bag" tone="blue" to="/reports/purchases" />
      <KpiCard label="Receivable" :value="money(data.finance.receivable)" icon="inbound" tone="amber" hint="Due from customers" to="/reports/customer-due" />
      <KpiCard label="Payable" :value="money(data.finance.payable)" icon="outbound" tone="red" hint="Due to suppliers" to="/reports/supplier-due" />
      <KpiCard label="Stock value" :value="money(data.stock_value)" icon="wallet" tone="slate" to="/reports/stock" />
    </div>

    <!-- Today & this month at a glance -->
    <div class="grid grid-cols-2 gap-4 xl:grid-cols-4">
      <KpiCard compact label="Today's sales" :value="money(data.finance.today_sales)" icon="calendar" tone="slate" />
      <KpiCard compact label="Today's profit" :value="money(data.finance.today_profit)" icon="trending" :tone="data.finance.today_profit < 0 ? 'red' : 'emerald'" />
      <KpiCard compact label="This month's sales" :value="money(data.finance.month_sales)" icon="calendar" tone="slate" />
      <KpiCard compact label="This month's profit" :value="money(data.finance.month_profit)" icon="trending" :tone="data.finance.month_profit < 0 ? 'red' : 'emerald'" />
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
        <h2 class="mb-4 flex items-center justify-between gap-2 font-semibold">
          <span class="flex items-center gap-2">
            <span class="text-amber-500"><Icon name="alert" :size="18" /></span> Low stock alerts
          </span>
          <span v-if="data.low_stock_count" class="badge bg-amber-100 text-amber-700 dark:bg-amber-500/20 dark:text-amber-300">
            {{ data.low_stock_count }}
          </span>
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
