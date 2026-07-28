<script setup>
import { ref, onMounted } from 'vue'
import api from '../lib/api'
import { downloadCsv } from '../lib/csv'

const money = (n) => '৳' + Number(n || 0).toLocaleString('en-IN', { minimumFractionDigits: 2 })
const data = ref(null)
const loading = ref(true)
const printPage = () => window.print()

function exportCsv() {
  if (!data.value) return
  // Raw numbers (not ৳-formatted) so the spreadsheet can sum them.
  const rows = data.value.rows.map((r) => [r.code, r.name, r.type, r.debit || 0, r.credit || 0])
  rows.push(['', 'Totals', '', data.value.total_debit, data.value.total_credit])
  const today = new Date().toISOString().slice(0, 10)
  downloadCsv(`trial-balance-${today}.csv`, ['Code', 'Account', 'Type', 'Debit', 'Credit'], rows)
}

onMounted(async () => {
  try {
    const res = await api.get('/accounting/trial-balance')
    data.value = res.data.data
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div>
    <div class="mb-6 flex items-center justify-between">
      <h1 class="text-2xl font-bold">Trial Balance</h1>
      <div class="flex gap-2 print:hidden">
        <button class="btn-ghost" :disabled="!data || !data.rows.length" @click="exportCsv">⬇️ Export CSV</button>
        <button class="btn-ghost" @click="printPage">🖨️ Print</button>
      </div>
    </div>

    <div v-if="loading" class="py-16 text-center text-slate-400">Loading…</div>

    <div v-else-if="data" class="card overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-left text-sm">
          <thead class="border-b border-slate-200 bg-slate-50 text-xs uppercase text-slate-500 dark:border-slate-700 dark:bg-slate-700/40">
            <tr>
              <th class="px-4 py-3">Code</th><th class="px-4 py-3">Account</th><th class="px-4 py-3">Type</th>
              <th class="px-4 py-3 text-right">Debit</th><th class="px-4 py-3 text-right">Credit</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 dark:divide-slate-700">
            <tr v-if="!data.rows.length"><td colspan="5" class="px-4 py-8 text-center text-slate-400">No postings yet</td></tr>
            <tr v-for="r in data.rows" :key="r.account_id" class="hover:bg-slate-50 dark:hover:bg-slate-700/30">
              <td class="px-4 py-3 font-mono text-slate-500">{{ r.code }}</td>
              <td class="px-4 py-3 font-medium">{{ r.name }}</td>
              <td class="px-4 py-3 capitalize text-slate-400">{{ r.type }}</td>
              <td class="px-4 py-3 text-right">{{ r.debit ? money(r.debit) : '' }}</td>
              <td class="px-4 py-3 text-right">{{ r.credit ? money(r.credit) : '' }}</td>
            </tr>
          </tbody>
          <tfoot>
            <tr class="border-t-2 border-slate-300 bg-slate-50 font-bold dark:border-slate-600 dark:bg-slate-700/40">
              <td class="px-4 py-3" colspan="3">Totals</td>
              <td class="px-4 py-3 text-right">{{ money(data.total_debit) }}</td>
              <td class="px-4 py-3 text-right">{{ money(data.total_credit) }}</td>
            </tr>
          </tfoot>
        </table>
      </div>
      <div class="px-4 py-3 text-center text-xs" :class="Math.abs(data.total_debit - data.total_credit) < 0.005 ? 'text-emerald-600' : 'text-red-600'">
        {{ Math.abs(data.total_debit - data.total_credit) < 0.005 ? '✓ Debits equal credits — books are balanced' : '⚠ Out of balance' }}
      </div>
    </div>
  </div>
</template>
