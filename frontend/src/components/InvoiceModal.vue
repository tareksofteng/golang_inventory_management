<script setup>
import { ref, computed, onMounted } from 'vue'
import api, { assetUrl } from '../lib/api'

const props = defineProps({
  // 'purchase' | 'sale' | 'purchase-return' | 'sale-return'
  type: { type: String, required: true },
  id: { type: [Number, String], required: true },
})
const emit = defineEmits(['close'])

const money = (n) => '৳ ' + Number(n || 0).toLocaleString('en-IN', { minimumFractionDigits: 2 })
const data = ref(null)
const loading = ref(true)

const CONFIG = {
  purchase: { endpoint: (id) => `/purchases/${id}`, party: 'supplier', unit: 'unit_cost', heading: 'Purchase', partyLabel: 'Supplier', isReturn: false },
  sale: { endpoint: (id) => `/sales/${id}`, party: 'customer', unit: 'unit_price', heading: 'Invoice', partyLabel: 'Bill To', isReturn: false },
  'purchase-return': { endpoint: (id) => `/returns/purchase/${id}`, party: 'supplier', unit: 'unit_cost', heading: 'Purchase Return', partyLabel: 'Supplier', isReturn: true },
  'sale-return': { endpoint: (id) => `/returns/sale/${id}`, party: 'customer', unit: 'unit_price', heading: 'Sales Return', partyLabel: 'Customer', isReturn: true },
}
const cfg = computed(() => CONFIG[props.type] || CONFIG.sale)
const party = computed(() => data.value?.[cfg.value.party])
const unitOf = (it) => it[cfg.value.unit]
const printPage = () => window.print()

onMounted(async () => {
  try {
    const res = await api.get(cfg.value.endpoint(props.id))
    data.value = res.data.data
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="fixed inset-0 z-40 flex items-start justify-center overflow-y-auto p-4 sm:items-center">
    <div class="no-print absolute inset-0 bg-slate-900/60 backdrop-blur-sm" @click="emit('close')" />

    <div class="printable relative z-10 my-4 w-full max-w-2xl overflow-hidden rounded-2xl bg-white text-slate-800 shadow-2xl">
      <div v-if="loading" class="py-20 text-center text-slate-400">Loading invoice…</div>

      <div v-else-if="data">
        <!-- Gradient header band -->
        <div class="relative bg-gradient-to-r from-brand-600 to-indigo-700 px-8 py-6 text-white">
          <div class="flex items-start justify-between">
            <div class="flex items-center gap-3">
              <span class="grid h-11 w-11 place-items-center rounded-xl bg-white/15 text-xl backdrop-blur">📦</span>
              <div>
                <div class="text-lg font-bold leading-tight">Inventory MS</div>
                <div class="text-xs text-white/70">Dhaka, Bangladesh · +880 1700-000000</div>
              </div>
            </div>
            <div class="text-right">
              <div class="text-2xl font-extrabold uppercase tracking-widest">{{ cfg.heading }}</div>
              <div class="mt-1 font-mono text-sm text-white/90">{{ data.invoice_no }}</div>
            </div>
          </div>
        </div>

        <div class="px-8 py-6">
          <!-- Meta row -->
          <div class="flex items-start justify-between border-b border-dashed border-slate-200 pb-5">
            <div>
              <div class="mb-1 text-[11px] font-semibold uppercase tracking-wider text-slate-400">
                {{ cfg.partyLabel }}
              </div>
              <div class="text-base font-bold text-slate-700">{{ party?.name }}</div>
              <div class="text-sm text-slate-500">{{ party?.phone }}</div>
              <div class="text-sm text-slate-500">{{ party?.address }}</div>
              <div v-if="party?.email" class="text-sm text-slate-500">{{ party.email }}</div>
            </div>
            <div class="text-right text-sm">
              <div class="text-slate-500">Date</div>
              <div class="font-medium text-slate-700">{{ new Date(data.created_at).toLocaleDateString() }}</div>
              <div class="mt-2 text-slate-500">Time</div>
              <div class="font-medium text-slate-700">{{ new Date(data.created_at).toLocaleTimeString() }}</div>
            </div>
          </div>

          <!-- Items -->
          <table class="mt-5 w-full text-left text-sm">
            <thead>
              <tr class="bg-slate-50 text-[11px] uppercase tracking-wider text-slate-500">
                <th class="rounded-l-lg px-3 py-2.5">Item</th>
                <th class="px-3 py-2.5 text-center">Qty</th>
                <th class="px-3 py-2.5 text-right">Rate</th>
                <th class="rounded-r-lg px-3 py-2.5 text-right">Amount</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="it in data.items" :key="it.id" class="border-b border-slate-100">
                <td class="px-3 py-3">
                  <div class="flex items-center gap-3">
                    <img v-if="it.product?.image" :src="assetUrl(it.product.image)" class="h-10 w-10 rounded-lg border border-slate-200 object-cover" />
                    <span v-else class="grid h-10 w-10 place-items-center rounded-lg bg-slate-100 text-[9px] text-slate-400">IMG</span>
                    <div>
                      <div class="font-semibold text-slate-700">{{ it.product?.name }}</div>
                      <div class="text-xs text-slate-400">{{ it.product?.sku }}</div>
                    </div>
                  </div>
                </td>
                <td class="px-3 py-3 text-center font-medium">{{ it.quantity }}</td>
                <td class="px-3 py-3 text-right">{{ money(unitOf(it)) }}</td>
                <td class="px-3 py-3 text-right font-medium">{{ money(it.subtotal) }}</td>
              </tr>
            </tbody>
          </table>

          <!-- Totals -->
          <div class="mt-5 flex justify-end">
            <div class="w-72 space-y-1.5 text-sm">
              <div class="flex justify-between text-slate-500"><span>Subtotal</span><span class="text-slate-700">{{ money(data.subtotal || data.total_amount) }}</span></div>
              <div v-if="data.discount" class="flex justify-between text-slate-500"><span>Discount</span><span class="text-rose-500">− {{ money(data.discount) }}</span></div>
              <div v-if="data.tax_amount" class="flex justify-between text-slate-500"><span>VAT / Tax ({{ data.tax_percent }}%)</span><span class="text-slate-700">{{ money(data.tax_amount) }}</span></div>
              <div class="my-1.5 flex items-center justify-between rounded-lg bg-gradient-to-r from-brand-600 to-indigo-700 px-3 py-2 text-white">
                <span class="text-sm font-semibold">{{ cfg.isReturn ? 'Total Return' : 'Grand Total' }}</span><span class="text-lg font-extrabold">{{ money(data.total_amount) }}</span>
              </div>
              <template v-if="!cfg.isReturn">
                <div class="flex justify-between text-slate-500"><span>Paid</span><span class="font-medium text-emerald-600">{{ money(data.paid_amount) }}</span></div>
                <div class="flex justify-between border-t border-slate-200 pt-1.5 font-semibold"><span class="text-slate-600">Balance Due</span><span :class="data.due > 0 ? 'text-amber-600' : 'text-emerald-600'">{{ money(data.due) }}</span></div>
              </template>
            </div>
          </div>

          <!-- Footer -->
          <div class="mt-8 flex items-end justify-between">
            <div class="text-xs text-slate-400">
              <p v-if="data.note" class="mb-2">Note: {{ data.note }}</p>
              <p>This is a computer-generated document.</p>
            </div>
            <div class="text-center">
              <div class="w-40 border-t border-slate-300 pt-1 text-xs text-slate-500">Authorized Signature</div>
            </div>
          </div>

          <div class="mt-6 rounded-lg bg-slate-50 py-2 text-center text-xs font-medium text-slate-400">
            Thank you for your business 🙏
          </div>
        </div>
      </div>

      <!-- Actions -->
      <div v-if="!loading" class="no-print flex justify-end gap-2 border-t border-slate-100 bg-slate-50 px-8 py-4">
        <button class="btn-ghost" @click="emit('close')">Close</button>
        <button class="btn-primary" @click="printPage">🖨️ Print / Save PDF</button>
      </div>
    </div>
  </div>
</template>
