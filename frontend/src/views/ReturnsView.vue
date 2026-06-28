<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import api, { assetUrl } from '../lib/api'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const route = useRoute()
const money = (n) => 'BDT ' + Number(n || 0).toLocaleString('en-IN', { minimumFractionDigits: 2 })

// Mode from the sidebar links (?tab=purchase|sale).
const mode = computed(() => {
  if (route.query.tab === 'sale' || route.query.tab === 'purchase') return route.query.tab
  return auth.can('purchase.manage') ? 'purchase' : 'sale'
})
const isPurchase = computed(() => mode.value === 'purchase')
const labels = computed(() =>
  isPurchase.value
    ? { title: 'Purchase Returns', sub: 'Return goods to suppliers and adjust stock', find: 'Find Purchase Invoice', party: 'Supplier', orderedCol: 'Purchased', unitCol: 'Unit Cost', no: 'Purchase No.' }
    : { title: 'Sales Returns', sub: 'Accept customer returns and adjust stock', find: 'Find Sale Invoice', party: 'Customer', orderedCol: 'Sold', unitCol: 'Unit Price', no: 'Sale No.' },
)

const search = ref('')
const lookup = ref(null)
const positions = ref([])
const note = ref('')
const error = ref('')
const searching = ref(false)
const saving = ref(false)

const recent = ref([])
const recentMeta = ref({ total: 0 })

const total = computed(() => positions.value.reduce((s, p) => s + Number(p.return_qty || 0) * Number(p.unit_value || 0), 0))

async function doSearch() {
  if (!search.value.trim()) return
  searching.value = true
  error.value = ''
  lookup.value = null
  try {
    const { data } = await api.get(`/returns/${mode.value}/lookup`, { params: { invoice: search.value.trim() } })
    lookup.value = data.data
    positions.value = data.data.items.map((it) => ({ ...it, return_qty: 0, unit_value: it.unit_value }))
  } catch (e) {
    error.value = e.response?.data?.message || 'Invoice not found'
  } finally {
    searching.value = false
  }
}

async function save() {
  const items = positions.value.filter((p) => Number(p.return_qty) > 0)
  if (!items.length) {
    error.value = 'Enter a return quantity for at least one item'
    return
  }
  saving.value = true
  error.value = ''
  const idKey = isPurchase.value ? 'purchase_id' : 'sale_id'
  try {
    await api.post(`/returns/${mode.value}`, {
      [idKey]: lookup.value.source_id,
      note: note.value,
      items: items.map((p) => ({ product_id: p.product_id, quantity: Number(p.return_qty), unit_value: Number(p.unit_value) })),
    })
    // reset and refresh
    search.value = ''
    lookup.value = null
    positions.value = []
    note.value = ''
    loadRecent()
  } catch (e) {
    error.value = e.response?.data?.message || 'Failed to save return'
  } finally {
    saving.value = false
  }
}

async function loadRecent() {
  const { data } = await api.get(`/returns/${mode.value}`, { params: { per_page: 8 } })
  recent.value = data.data
  recentMeta.value = data.meta
}

const today = new Date().toLocaleDateString()

watch(mode, () => {
  search.value = ''
  lookup.value = null
  positions.value = []
  error.value = ''
  loadRecent()
})
onMounted(loadRecent)
</script>

<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-2xl font-bold">{{ labels.title }}</h1>
      <p class="text-sm text-slate-500">{{ labels.sub }}</p>
    </div>

    <!-- Find invoice -->
    <div class="card p-5">
      <label class="mb-2 flex items-center gap-2 text-xs font-bold uppercase tracking-wider text-brand-600">
        <svg class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" stroke-linejoin="round" d="m21 21-4.3-4.3m1.8-4.45a6.25 6.25 0 1 1-12.5 0 6.25 6.25 0 0 1 12.5 0Z" /></svg>
        {{ labels.find }}
      </label>
      <div class="flex gap-2">
        <input
          v-model="search"
          class="input"
          :placeholder="isPurchase ? 'e.g. PUR-000006' : 'e.g. SAL-000006'"
          @keyup.enter="doSearch"
        />
        <button class="btn-primary whitespace-nowrap" :disabled="searching" @click="doSearch">{{ searching ? 'Searching…' : 'Search' }}</button>
      </div>
      <p v-if="error" class="mt-2 text-sm text-red-600">{{ error }}</p>
    </div>

    <template v-if="lookup">
      <!-- Party + invoice details -->
      <div class="grid gap-4 lg:grid-cols-2">
        <div class="card p-5">
          <div class="mb-2 text-xs font-bold uppercase tracking-wider text-brand-600">{{ labels.party }}</div>
          <div class="text-lg font-bold">{{ lookup.party_name || '—' }}</div>
          <div class="text-sm text-slate-500">{{ lookup.party_phone }}</div>
          <div class="text-sm text-slate-500">{{ lookup.party_address }}</div>
        </div>
        <div class="card p-5 text-sm">
          <div class="mb-2 text-xs font-bold uppercase tracking-wider text-brand-600">Invoice Details</div>
          <div class="flex justify-between py-0.5"><span class="text-slate-500">{{ labels.no }}</span><span class="font-mono font-semibold">{{ lookup.invoice_no }}</span></div>
          <div class="flex justify-between py-0.5"><span class="text-slate-500">Date</span><span class="font-medium">{{ new Date(lookup.date).toLocaleDateString() }}</span></div>
        </div>
      </div>

      <!-- Return positions -->
      <div class="card overflow-hidden">
        <div class="border-b border-slate-200 px-5 py-3 font-semibold dark:border-slate-700">Return Positions</div>
        <div class="overflow-x-auto">
          <table class="w-full text-left text-sm">
            <thead class="bg-slate-50 text-xs uppercase text-slate-500 dark:bg-slate-700/40">
              <tr>
                <th class="px-4 py-3">#</th>
                <th class="px-4 py-3">Product</th>
                <th class="px-4 py-3 text-center">{{ labels.orderedCol }}</th>
                <th class="px-4 py-3 text-center">Already Returned</th>
                <th class="px-4 py-3 text-center">Available</th>
                <th class="w-28 px-4 py-3 text-center">Return Qty</th>
                <th class="w-32 px-4 py-3 text-center">{{ labels.unitCol }}</th>
                <th class="px-4 py-3 text-right">Amount</th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100 dark:divide-slate-700">
              <tr v-for="(p, i) in positions" :key="p.product_id">
                <td class="px-4 py-3 text-slate-400">{{ i + 1 }}</td>
                <td class="px-4 py-3">
                  <div class="flex items-center gap-3">
                    <img v-if="p.image" :src="assetUrl(p.image)" class="h-9 w-9 rounded-lg border border-slate-200 object-cover" />
                    <span v-else class="grid h-9 w-9 place-items-center rounded-lg bg-slate-100 text-[9px] text-slate-400 dark:bg-slate-700">IMG</span>
                    <div>
                      <div class="font-medium">{{ p.name }}</div>
                      <div class="text-xs text-slate-400">{{ p.sku }}</div>
                    </div>
                  </div>
                </td>
                <td class="px-4 py-3 text-center">{{ p.ordered }}</td>
                <td class="px-4 py-3 text-center text-slate-400">{{ p.already_returned || '—' }}</td>
                <td class="px-4 py-3 text-center font-semibold text-emerald-600">{{ p.available }}</td>
                <td class="px-4 py-3">
                  <input v-model.number="p.return_qty" type="number" min="0" :max="p.available" class="input !py-1.5 text-center" :disabled="p.available <= 0" />
                </td>
                <td class="px-4 py-3">
                  <input v-model.number="p.unit_value" type="number" min="0" class="input !py-1.5 text-center" />
                </td>
                <td class="px-4 py-3 text-right font-semibold">{{ money(p.return_qty * p.unit_value) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <!-- Footer: note + total + save -->
      <div class="grid gap-4 lg:grid-cols-3">
        <div class="card p-5 lg:col-span-2">
          <div class="mb-3 text-xs font-bold uppercase tracking-wider text-slate-400">Return Date: {{ today }}</div>
          <label class="label">Note</label>
          <textarea v-model="note" rows="2" class="input" placeholder="Reason for return (optional)"></textarea>
        </div>
        <div class="card flex flex-col justify-between p-5">
          <div class="rounded-xl bg-amber-50 p-4 text-center dark:bg-amber-500/10">
            <div class="text-xs font-bold uppercase tracking-wider text-amber-600">Total Return</div>
            <div class="mt-1 text-2xl font-extrabold text-amber-600">{{ money(total) }}</div>
          </div>
          <button class="btn-primary mt-3 w-full" :disabled="saving || total <= 0" @click="save">
            {{ saving ? 'Saving…' : '↩ Save Return' }}
          </button>
        </div>
      </div>
    </template>

    <!-- Recent returns -->
    <div class="card overflow-hidden">
      <div class="flex items-center justify-between border-b border-slate-200 px-5 py-3 dark:border-slate-700">
        <span class="font-semibold">Recent Returns</span>
        <span class="text-xs text-slate-400">{{ recentMeta.total }} total</span>
      </div>
      <div class="overflow-x-auto">
        <table class="w-full text-left text-sm">
          <thead class="bg-slate-50 text-xs uppercase text-slate-500 dark:bg-slate-700/40">
            <tr>
              <th class="px-4 py-3">Return No.</th>
              <th class="px-4 py-3">{{ labels.party }}</th>
              <th class="px-4 py-3 text-right">Amount</th>
              <th class="px-4 py-3">Date</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 dark:divide-slate-700">
            <tr v-if="!recent.length"><td colspan="4" class="px-4 py-8 text-center text-slate-400">No returns yet</td></tr>
            <tr v-for="r in recent" :key="r.id" class="hover:bg-slate-50 dark:hover:bg-slate-700/30">
              <td class="px-4 py-3 font-medium">{{ r.invoice_no }}</td>
              <td class="px-4 py-3">{{ (r.supplier || r.customer)?.name }}</td>
              <td class="px-4 py-3 text-right font-semibold">{{ money(r.total_amount) }}</td>
              <td class="px-4 py-3 text-slate-400">{{ new Date(r.created_at).toLocaleDateString() }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>
