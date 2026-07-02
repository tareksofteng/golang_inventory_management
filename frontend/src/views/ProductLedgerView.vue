<script setup>
import { ref, watch, onMounted } from 'vue'
import api, { assetUrl } from '../lib/api'

const products = ref([])
const productId = ref('')
const ledger = ref(null)
const loading = ref(false)

const typeBadge = (t) => {
  const map = {
    Purchase: 'bg-blue-100 text-blue-700 dark:bg-blue-500/20 dark:text-blue-300',
    Sale: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-500/20 dark:text-emerald-300',
    'Purchase Return': 'bg-amber-100 text-amber-700 dark:bg-amber-500/20 dark:text-amber-300',
    'Sale Return': 'bg-violet-100 text-violet-700 dark:bg-violet-500/20 dark:text-violet-300',
  }
  return map[t] || 'bg-slate-100 text-slate-600'
}
const printPage = () => window.print()

async function load() {
  if (!productId.value) {
    ledger.value = null
    return
  }
  loading.value = true
  try {
    const { data } = await api.get(`/ledger/product/${productId.value}`)
    ledger.value = data.data
  } catch (e) {
    ledger.value = null
  } finally {
    loading.value = false
  }
}

watch(productId, load)

onMounted(async () => {
  const res = await api.get('/products', { params: { per_page: 200 } })
  products.value = res.data.data
  productId.value = products.value[0]?.id || ''
  load()
})
</script>

<template>
  <div>
    <div class="mb-6 flex flex-wrap items-center justify-between gap-3">
      <h1 class="text-2xl font-bold">Product Ledger</h1>
      <button class="btn-ghost print:hidden" @click="printPage">🖨️ Print</button>
    </div>

    <div class="mb-4 max-w-md print:hidden">
      <label class="label">Product</label>
      <select v-model="productId" class="input">
        <option v-for="p in products" :key="p.id" :value="p.id">{{ p.name }} — {{ p.sku }}</option>
      </select>
    </div>

    <div v-if="loading" class="py-16 text-center text-slate-400">Loading…</div>

    <div v-else-if="ledger" class="space-y-4">
      <!-- Product header + opening/closing -->
      <div class="card flex flex-wrap items-center gap-4 p-5">
        <img v-if="ledger.image" :src="assetUrl(ledger.image)" class="h-16 w-16 rounded-xl border border-slate-200 object-cover" />
        <span v-else class="grid h-16 w-16 place-items-center rounded-xl bg-slate-100 text-xs text-slate-400 dark:bg-slate-700">IMG</span>
        <div class="mr-auto">
          <div class="text-lg font-bold">{{ ledger.name }}</div>
          <div class="text-sm text-slate-400">{{ ledger.sku }}</div>
        </div>
        <div class="rounded-xl bg-slate-50 px-5 py-3 text-center dark:bg-slate-700/40">
          <div class="text-xs font-semibold uppercase tracking-wider text-slate-400">Opening</div>
          <div class="text-xl font-bold">{{ ledger.opening }}</div>
        </div>
        <div class="rounded-xl bg-brand-50 px-5 py-3 text-center dark:bg-brand-600/20">
          <div class="text-xs font-semibold uppercase tracking-wider text-brand-600">Closing (Current Stock)</div>
          <div class="text-xl font-bold text-brand-600">{{ ledger.closing }} {{ ledger.unit }}</div>
        </div>
      </div>

      <!-- Movement table -->
      <div class="card overflow-hidden">
        <div class="overflow-x-auto">
          <table class="w-full text-left text-sm">
            <thead class="bg-slate-50 text-xs uppercase text-slate-500 dark:bg-slate-700/40">
              <tr>
                <th class="px-4 py-3">Date</th>
                <th class="px-4 py-3">Type</th>
                <th class="px-4 py-3">Reference</th>
                <th class="px-4 py-3 text-center">In</th>
                <th class="px-4 py-3 text-center">Out</th>
                <th class="px-4 py-3 text-right">Balance</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100 dark:divide-slate-700">
              <tr class="bg-slate-50/50 dark:bg-slate-700/20">
                <td class="px-4 py-2.5 text-slate-400" colspan="5">Opening Balance</td>
                <td class="px-4 py-2.5 text-right font-semibold">{{ ledger.opening }}</td>
              </tr>
              <tr v-if="!ledger.entries.length"><td colspan="6" class="px-4 py-8 text-center text-slate-400">No stock movements</td></tr>
              <tr v-for="(e, i) in ledger.entries" :key="i" class="hover:bg-slate-50 dark:hover:bg-slate-700/30">
                <td class="px-4 py-3 text-slate-400">{{ e.date }}</td>
                <td class="px-4 py-3"><span class="badge" :class="typeBadge(e.type)">{{ e.type }}</span></td>
                <td class="px-4 py-3 font-mono text-slate-500">{{ e.ref }}</td>
                <td class="px-4 py-3 text-center font-medium text-emerald-600">{{ e.in || '—' }}</td>
                <td class="px-4 py-3 text-center font-medium text-red-500">{{ e.out || '—' }}</td>
                <td class="px-4 py-3 text-right font-semibold">{{ e.balance }}</td>
              </tr>
            </tbody>
            <tfoot v-if="ledger.entries.length">
              <tr class="border-t-2 border-slate-200 font-bold dark:border-slate-600">
                <td class="px-4 py-3" colspan="5">Closing Balance</td>
                <td class="px-4 py-3 text-right text-brand-600">{{ ledger.closing }}</td>
              </tr>
            </tfoot>
          </table>
        </div>
      </div>
    </div>

    <div v-else class="py-16 text-center text-slate-400">Select a product to view its stock ledger</div>
  </div>
</template>
