<script setup>
import { ref, computed, onMounted } from 'vue'
import api from '../lib/api'

const props = defineProps({
  type: { type: String, required: true }, // 'purchase' | 'sale'
  id: { type: [Number, String], required: true },
})
const emit = defineEmits(['close'])

const money = (n) => '৳' + Number(n || 0).toLocaleString('en-IN')
const data = ref(null)
const loading = ref(true)

const isPurchase = computed(() => props.type === 'purchase')
const party = computed(() => (isPurchase.value ? data.value?.supplier : data.value?.customer))
const partyLabel = computed(() => (isPurchase.value ? 'Supplier' : 'Bill To'))
const unitOf = (it) => (isPurchase.value ? it.unit_cost : it.unit_price)

const printPage = () => window.print()

onMounted(async () => {
  try {
    const url = isPurchase.value ? `/purchases/${props.id}` : `/sales/${props.id}`
    const res = await api.get(url)
    data.value = res.data.data
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="fixed inset-0 z-40 flex items-start justify-center overflow-y-auto p-4 sm:items-center">
    <div class="no-print absolute inset-0 bg-black/40 backdrop-blur-sm" @click="emit('close')" />

    <div class="printable relative z-10 my-4 w-full max-w-2xl rounded-xl bg-white p-8 shadow-xl dark:bg-white dark:text-slate-800">
      <div v-if="loading" class="py-16 text-center text-slate-400">Loading invoice…</div>

      <div v-else-if="data">
        <!-- Header -->
        <div class="flex items-start justify-between border-b border-slate-200 pb-5">
          <div>
            <div class="flex items-center gap-2">
              <span class="grid h-9 w-9 place-items-center rounded-lg bg-brand-600 text-white">📦</span>
              <span class="text-xl font-bold">Inventory MS</span>
            </div>
            <p class="mt-1 text-xs text-slate-500">Dhaka, Bangladesh · +880 1700-000000</p>
          </div>
          <div class="text-right">
            <h2 class="text-2xl font-bold uppercase tracking-wide text-slate-700">
              {{ isPurchase ? 'Purchase' : 'Invoice' }}
            </h2>
            <p class="mt-1 text-sm font-semibold text-brand-600">{{ data.invoice_no }}</p>
            <p class="text-xs text-slate-500">{{ new Date(data.created_at).toLocaleString() }}</p>
          </div>
        </div>

        <!-- Party -->
        <div class="grid grid-cols-2 gap-4 py-5 text-sm">
          <div>
            <div class="mb-1 text-xs font-semibold uppercase text-slate-400">{{ partyLabel }}</div>
            <div class="font-bold">{{ party?.name }}</div>
            <div class="text-slate-500">{{ party?.phone }}</div>
            <div class="text-slate-500">{{ party?.address }}</div>
            <div v-if="party?.email" class="text-slate-500">{{ party.email }}</div>
          </div>
          <div class="text-right">
            <div class="mb-1 text-xs font-semibold uppercase text-slate-400">Status</div>
            <span
              class="badge"
              :class="data.due > 0 ? 'bg-amber-100 text-amber-700' : 'bg-emerald-100 text-emerald-700'"
            >{{ data.due > 0 ? 'Due ' + money(data.due) : 'Paid' }}</span>
          </div>
        </div>

        <!-- Items -->
        <table class="w-full text-left text-sm">
          <thead class="border-y border-slate-200 text-xs uppercase text-slate-500">
            <tr>
              <th class="py-2">#</th>
              <th class="py-2">Item</th>
              <th class="py-2 text-center">Qty</th>
              <th class="py-2 text-right">Rate</th>
              <th class="py-2 text-right">Amount</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100">
            <tr v-for="(it, i) in data.items" :key="it.id">
              <td class="py-2 text-slate-400">{{ i + 1 }}</td>
              <td class="py-2">
                <div class="font-medium">{{ it.product?.name }}</div>
                <div class="text-xs text-slate-400">{{ it.product?.sku }}</div>
              </td>
              <td class="py-2 text-center">{{ it.quantity }}</td>
              <td class="py-2 text-right">{{ money(unitOf(it)) }}</td>
              <td class="py-2 text-right">{{ money(it.subtotal) }}</td>
            </tr>
          </tbody>
        </table>

        <!-- Totals -->
        <div class="mt-4 flex justify-end">
          <div class="w-64 space-y-1 text-sm">
            <div class="flex justify-between"><span class="text-slate-500">Subtotal</span><span>{{ money(data.subtotal || data.total_amount) }}</span></div>
            <div v-if="data.discount" class="flex justify-between"><span class="text-slate-500">Discount</span><span>- {{ money(data.discount) }}</span></div>
            <div v-if="data.tax_amount" class="flex justify-between"><span class="text-slate-500">VAT/Tax ({{ data.tax_percent }}%)</span><span>{{ money(data.tax_amount) }}</span></div>
            <div class="flex justify-between border-t border-slate-200 pt-1 text-base font-bold"><span>Grand Total</span><span>{{ money(data.total_amount) }}</span></div>
            <div class="flex justify-between text-emerald-600"><span>Paid</span><span>{{ money(data.paid_amount) }}</span></div>
            <div class="flex justify-between font-semibold text-amber-600"><span>Due</span><span>{{ money(data.due) }}</span></div>
          </div>
        </div>

        <p v-if="data.note" class="mt-4 border-t border-slate-200 pt-3 text-xs text-slate-500">Note: {{ data.note }}</p>
        <p class="mt-6 text-center text-xs text-slate-400">Thank you for your business!</p>

        <!-- Actions -->
        <div class="no-print mt-6 flex justify-end gap-2">
          <button class="btn-ghost" @click="emit('close')">Close</button>
          <button class="btn-primary" @click="printPage">🖨️ Print</button>
        </div>
      </div>
    </div>
  </div>
</template>
