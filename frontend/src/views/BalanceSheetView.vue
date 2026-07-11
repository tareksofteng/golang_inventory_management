<script setup>
import { ref, onMounted } from 'vue'
import api from '../lib/api'

const money = (n) => '৳' + Number(n || 0).toLocaleString('en-IN', { minimumFractionDigits: 2 })
const data = ref(null)
const loading = ref(true)
const printPage = () => window.print()

const today = new Date()
const pad = (n) => String(n).padStart(2, '0')
const asOf = ref(`${today.getFullYear()}-${pad(today.getMonth() + 1)}-${pad(today.getDate())}`)

async function load() {
  loading.value = true
  try {
    const res = await api.get('/accounting/balance-sheet', { params: { as_of: asOf.value } })
    data.value = res.data.data
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<template>
  <div>
    <div class="mb-6 flex items-center justify-between">
      <h1 class="text-2xl font-bold">Balance Sheet</h1>
      <button class="btn-ghost print:hidden" @click="printPage">🖨️ Print</button>
    </div>

    <div class="card mb-4 flex flex-wrap items-end gap-3 p-4 print:hidden">
      <div><label class="label">As of</label><input v-model="asOf" type="date" class="input" /></div>
      <button class="btn-primary" @click="load">Apply</button>
    </div>

    <div v-if="loading" class="py-16 text-center text-slate-400">Loading…</div>

    <div v-else-if="data" class="space-y-4">
      <p class="text-center text-sm text-slate-400">As of {{ data.as_of }}</p>

      <div class="grid gap-4 lg:grid-cols-2">
        <!-- Assets -->
        <div class="card overflow-hidden">
          <div class="border-b border-slate-200 bg-blue-50 px-5 py-3 font-semibold text-blue-700 dark:border-slate-700 dark:bg-blue-500/10 dark:text-blue-300">Assets</div>
          <table class="w-full text-left text-sm">
            <tbody class="divide-y divide-slate-100 dark:divide-slate-700">
              <tr v-if="!(data.assets || []).length"><td class="px-5 py-3 text-slate-400">—</td><td></td></tr>
              <tr v-for="a in data.assets" :key="a.code">
                <td class="px-5 py-2.5"><span class="font-mono text-slate-400">{{ a.code }}</span> {{ a.name }}</td>
                <td class="px-5 py-2.5 text-right">{{ money(a.amount) }}</td>
              </tr>
            </tbody>
            <tfoot><tr class="border-t border-slate-200 bg-slate-50 font-bold dark:border-slate-700 dark:bg-slate-700/40"><td class="px-5 py-2.5">Total Assets</td><td class="px-5 py-2.5 text-right text-blue-700 dark:text-blue-300">{{ money(data.total_assets) }}</td></tr></tfoot>
          </table>
        </div>

        <!-- Liabilities + Equity -->
        <div class="card overflow-hidden">
          <div class="border-b border-slate-200 bg-amber-50 px-5 py-3 font-semibold text-amber-700 dark:border-slate-700 dark:bg-amber-500/10 dark:text-amber-300">Liabilities &amp; Equity</div>
          <table class="w-full text-left text-sm">
            <tbody class="divide-y divide-slate-100 dark:divide-slate-700">
              <tr class="bg-slate-50/60 dark:bg-slate-700/20"><td class="px-5 py-1.5 text-xs font-semibold uppercase text-slate-400" colspan="2">Liabilities</td></tr>
              <tr v-if="!(data.liabilities || []).length"><td class="px-5 py-2 text-slate-400">None</td><td></td></tr>
              <tr v-for="l in data.liabilities" :key="l.code">
                <td class="px-5 py-2.5"><span class="font-mono text-slate-400">{{ l.code }}</span> {{ l.name }}</td>
                <td class="px-5 py-2.5 text-right">{{ money(l.amount) }}</td>
              </tr>
              <tr class="bg-slate-50/60 dark:bg-slate-700/20"><td class="px-5 py-1.5 text-xs font-semibold uppercase text-slate-400" colspan="2">Equity</td></tr>
              <tr v-for="e in data.equity" :key="e.code">
                <td class="px-5 py-2.5"><span class="font-mono text-slate-400">{{ e.code }}</span> {{ e.name }}</td>
                <td class="px-5 py-2.5 text-right">{{ money(e.amount) }}</td>
              </tr>
              <tr>
                <td class="px-5 py-2.5 text-slate-500">Retained Earnings (Net Profit)</td>
                <td class="px-5 py-2.5 text-right">{{ money(data.net_profit) }}</td>
              </tr>
            </tbody>
            <tfoot><tr class="border-t border-slate-200 bg-slate-50 font-bold dark:border-slate-700 dark:bg-slate-700/40"><td class="px-5 py-2.5">Total Liabilities &amp; Equity</td><td class="px-5 py-2.5 text-right text-amber-700 dark:text-amber-300">{{ money(data.total_liabilities + data.total_equity) }}</td></tr></tfoot>
          </table>
        </div>
      </div>

      <div class="card py-3 text-center text-sm font-medium" :class="data.balanced ? 'text-emerald-600' : 'text-red-600'">
        {{ data.balanced ? '✓ Assets = Liabilities + Equity — the balance sheet balances' : '⚠ Out of balance' }}
      </div>
    </div>
  </div>
</template>
