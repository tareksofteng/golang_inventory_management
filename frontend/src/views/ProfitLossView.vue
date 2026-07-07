<script setup>
import { ref, onMounted } from 'vue'
import api from '../lib/api'

const money = (n) => '৳' + Number(n || 0).toLocaleString('en-IN', { minimumFractionDigits: 2 })
const data = ref(null)
const loading = ref(true)
const printPage = () => window.print()

onMounted(async () => {
  try {
    const res = await api.get('/accounting/profit-loss')
    data.value = res.data.data
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div>
    <div class="mb-6 flex items-center justify-between">
      <h1 class="text-2xl font-bold">Profit &amp; Loss</h1>
      <button class="btn-ghost print:hidden" @click="printPage">🖨️ Print</button>
    </div>

    <div v-if="loading" class="py-16 text-center text-slate-400">Loading…</div>

    <div v-else-if="data" class="mx-auto max-w-2xl space-y-4">
      <!-- Income -->
      <div class="card overflow-hidden">
        <div class="border-b border-slate-200 bg-emerald-50 px-5 py-3 font-semibold text-emerald-700 dark:border-slate-700 dark:bg-emerald-500/10 dark:text-emerald-300">Income</div>
        <table class="w-full text-left text-sm">
          <tbody class="divide-y divide-slate-100 dark:divide-slate-700">
            <tr v-if="!(data.income || []).length"><td class="px-5 py-3 text-slate-400">No income yet</td><td></td></tr>
            <tr v-for="i in data.income" :key="i.code">
              <td class="px-5 py-2.5"><span class="font-mono text-slate-400">{{ i.code }}</span> {{ i.name }}</td>
              <td class="px-5 py-2.5 text-right">{{ money(i.amount) }}</td>
            </tr>
          </tbody>
          <tfoot><tr class="border-t border-slate-200 bg-slate-50 font-semibold dark:border-slate-700 dark:bg-slate-700/40"><td class="px-5 py-2.5">Total Income</td><td class="px-5 py-2.5 text-right">{{ money(data.total_income) }}</td></tr></tfoot>
        </table>
      </div>

      <!-- Expenses -->
      <div class="card overflow-hidden">
        <div class="border-b border-slate-200 bg-red-50 px-5 py-3 font-semibold text-red-700 dark:border-slate-700 dark:bg-red-500/10 dark:text-red-300">Expenses</div>
        <table class="w-full text-left text-sm">
          <tbody class="divide-y divide-slate-100 dark:divide-slate-700">
            <tr v-if="!(data.expenses || []).length"><td class="px-5 py-3 text-slate-400">No expenses yet</td><td></td></tr>
            <tr v-for="e in data.expenses" :key="e.code">
              <td class="px-5 py-2.5"><span class="font-mono text-slate-400">{{ e.code }}</span> {{ e.name }}</td>
              <td class="px-5 py-2.5 text-right">{{ money(e.amount) }}</td>
            </tr>
          </tbody>
          <tfoot><tr class="border-t border-slate-200 bg-slate-50 font-semibold dark:border-slate-700 dark:bg-slate-700/40"><td class="px-5 py-2.5">Total Expense</td><td class="px-5 py-2.5 text-right">{{ money(data.total_expense) }}</td></tr></tfoot>
        </table>
      </div>

      <!-- Net profit -->
      <div class="card flex items-center justify-between p-5" :class="data.net_profit >= 0 ? 'bg-gradient-to-r from-emerald-600 to-teal-600' : 'bg-gradient-to-r from-red-600 to-rose-600'">
        <span class="text-lg font-semibold text-white">{{ data.net_profit >= 0 ? 'Net Profit' : 'Net Loss' }}</span>
        <span class="text-2xl font-extrabold text-white">{{ money(Math.abs(data.net_profit)) }}</span>
      </div>
    </div>
  </div>
</template>
