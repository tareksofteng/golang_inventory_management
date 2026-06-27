<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import api from '../lib/api'

const money = (n) => '৳' + Number(n || 0).toLocaleString('en-IN')

const mode = ref('supplier') // 'customer' | 'supplier'
const customers = ref([])
const suppliers = ref([])
const partyId = ref('')
const ledger = ref(null)
const loading = ref(false)

const parties = computed(() => (mode.value === 'customer' ? customers.value : suppliers.value))
const printPage = () => window.print()

async function load() {
  if (!partyId.value) {
    ledger.value = null
    return
  }
  loading.value = true
  try {
    const { data } = await api.get(`/ledger/${mode.value}/${partyId.value}`)
    ledger.value = data.data
  } catch (e) {
    ledger.value = null
  } finally {
    loading.value = false
  }
}

function switchMode(m) {
  mode.value = m
  partyId.value = parties.value[0]?.id || ''
  load()
}

watch(partyId, load)

onMounted(async () => {
  const [cus, sup] = await Promise.all([
    api.get('/customers', { params: { per_page: 100 } }),
    api.get('/suppliers', { params: { per_page: 100 } }),
  ])
  customers.value = cus.data.data
  suppliers.value = sup.data.data
  partyId.value = suppliers.value[0]?.id || ''
  load()
})
</script>

<template>
  <div>
    <div class="mb-6 flex flex-wrap items-center justify-between gap-3">
      <h1 class="text-2xl font-bold">Ledger</h1>
      <button class="btn-ghost print:hidden" @click="printPage">🖨️ Print</button>
    </div>

    <div class="mb-4 flex flex-wrap items-end gap-3 print:hidden">
      <div class="flex gap-2">
        <button class="btn" :class="mode === 'supplier' ? 'bg-brand-600 text-white' : 'btn-ghost'" @click="switchMode('supplier')">Supplier</button>
        <button class="btn" :class="mode === 'customer' ? 'bg-brand-600 text-white' : 'btn-ghost'" @click="switchMode('customer')">Customer</button>
      </div>
      <div class="min-w-56">
        <label class="label">{{ mode === 'customer' ? 'Customer' : 'Supplier' }}</label>
        <select v-model="partyId" class="input">
          <option v-for="x in parties" :key="x.id" :value="x.id">{{ x.name }}</option>
        </select>
      </div>
    </div>

    <div v-if="loading" class="py-16 text-center text-slate-400">Loading…</div>

    <div v-else-if="ledger" class="space-y-4">
      <div class="card flex flex-wrap items-center justify-between gap-3 p-4">
        <div>
          <div class="text-sm text-slate-500">Statement for</div>
          <div class="text-lg font-bold">{{ ledger.party_name }}</div>
        </div>
        <div class="text-right">
          <div class="text-sm text-slate-500">Closing Balance</div>
          <div class="text-xl font-bold" :class="ledger.closing_balance > 0 ? 'text-amber-600' : 'text-emerald-600'">
            {{ money(ledger.closing_balance) }}
          </div>
        </div>
      </div>

      <div class="card overflow-hidden">
        <div class="overflow-x-auto">
          <table class="w-full text-left text-sm">
            <thead class="bg-slate-50 text-xs uppercase text-slate-500 dark:bg-slate-700/40">
              <tr>
                <th class="px-4 py-3">Date</th><th class="px-4 py-3">Type</th><th class="px-4 py-3">Reference</th>
                <th class="px-4 py-3 text-right">Debit</th><th class="px-4 py-3 text-right">Credit</th><th class="px-4 py-3 text-right">Balance</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100 dark:divide-slate-700">
              <tr v-if="!ledger.entries.length"><td colspan="6" class="px-4 py-8 text-center text-slate-400">No transactions</td></tr>
              <tr v-for="(e, i) in ledger.entries" :key="i" class="hover:bg-slate-50 dark:hover:bg-slate-700/30">
                <td class="px-4 py-3 text-slate-400">{{ e.date }}</td>
                <td class="px-4 py-3 font-medium">{{ e.type }}</td>
                <td class="px-4 py-3 text-slate-500">{{ e.ref }}</td>
                <td class="px-4 py-3 text-right">{{ e.debit ? money(e.debit) : '—' }}</td>
                <td class="px-4 py-3 text-right text-emerald-600">{{ e.credit ? money(e.credit) : '—' }}</td>
                <td class="px-4 py-3 text-right font-semibold">{{ money(e.balance) }}</td>
              </tr>
            </tbody>
            <tfoot v-if="ledger.entries.length">
              <tr class="border-t-2 border-slate-200 font-bold dark:border-slate-600">
                <td class="px-4 py-3" colspan="5">Closing Balance</td>
                <td class="px-4 py-3 text-right">{{ money(ledger.closing_balance) }}</td>
              </tr>
            </tfoot>
          </table>
        </div>
      </div>
    </div>

    <div v-else class="py-16 text-center text-slate-400">Select a party to view the ledger</div>
  </div>
</template>
