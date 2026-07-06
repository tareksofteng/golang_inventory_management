<script setup>
import { ref, watch, onMounted } from 'vue'
import api from '../lib/api'

const money = (n) => '৳' + Number(n || 0).toLocaleString('en-IN', { minimumFractionDigits: 2 })

const accounts = ref([])
const accountId = ref('')
const ledger = ref(null)
const loading = ref(false)
const printPage = () => window.print()

async function load() {
  if (!accountId.value) {
    ledger.value = null
    return
  }
  loading.value = true
  try {
    const { data } = await api.get(`/accounting/ledger/${accountId.value}`)
    ledger.value = data.data
  } catch (e) {
    ledger.value = null
  } finally {
    loading.value = false
  }
}

watch(accountId, load)

onMounted(async () => {
  const res = await api.get('/accounts/active')
  accounts.value = res.data.data
  accountId.value = accounts.value[0]?.id || ''
  load()
})
</script>

<template>
  <div>
    <div class="mb-6 flex items-center justify-between">
      <h1 class="text-2xl font-bold">Account Ledger</h1>
      <button class="btn-ghost print:hidden" @click="printPage">🖨️ Print</button>
    </div>

    <div class="mb-4 max-w-md print:hidden">
      <label class="label">Account</label>
      <select v-model="accountId" class="input">
        <option v-for="a in accounts" :key="a.id" :value="a.id">{{ a.code }} — {{ a.name }}</option>
      </select>
    </div>

    <div v-if="loading" class="py-16 text-center text-slate-400">Loading…</div>

    <div v-else-if="ledger" class="space-y-4">
      <div class="card flex flex-wrap items-center justify-between gap-3 p-5">
        <div>
          <div class="text-lg font-bold">{{ ledger.account.code }} — {{ ledger.account.name }}</div>
          <div class="text-sm capitalize text-slate-400">{{ ledger.account.type }}</div>
        </div>
        <div class="rounded-xl bg-brand-50 px-5 py-3 text-center dark:bg-brand-600/20">
          <div class="text-xs font-semibold uppercase tracking-wider text-brand-600">Balance</div>
          <div class="text-xl font-bold text-brand-600">{{ money(ledger.closing) }}</div>
        </div>
      </div>

      <div class="card overflow-hidden">
        <div class="overflow-x-auto">
          <table class="w-full text-left text-sm">
            <thead class="bg-slate-50 text-xs uppercase text-slate-500 dark:bg-slate-700/40">
              <tr>
                <th class="px-4 py-3">Date</th><th class="px-4 py-3">Entry</th><th class="px-4 py-3">Reference</th>
                <th class="px-4 py-3 text-right">Debit</th><th class="px-4 py-3 text-right">Credit</th><th class="px-4 py-3 text-right">Balance</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100 dark:divide-slate-700">
              <tr v-if="!ledger.entries.length"><td colspan="6" class="px-4 py-8 text-center text-slate-400">No postings</td></tr>
              <tr v-for="(e, i) in ledger.entries" :key="i" class="hover:bg-slate-50 dark:hover:bg-slate-700/30">
                <td class="px-4 py-3 text-slate-400">{{ e.date }}</td>
                <td class="px-4 py-3 font-mono text-slate-500">{{ e.entry_no }}</td>
                <td class="px-4 py-3">{{ e.ref || e.note || '—' }}</td>
                <td class="px-4 py-3 text-right">{{ e.debit ? money(e.debit) : '' }}</td>
                <td class="px-4 py-3 text-right">{{ e.credit ? money(e.credit) : '' }}</td>
                <td class="px-4 py-3 text-right font-semibold">{{ money(e.balance) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <div v-else class="py-16 text-center text-slate-400">Select an account to view its ledger</div>
  </div>
</template>
