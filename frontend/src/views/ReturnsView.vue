<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import api from '../lib/api'
import Modal from '../components/Modal.vue'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const money = (n) => '৳' + Number(n || 0).toLocaleString('en-IN')

// Which history we are viewing: 'purchase' | 'sale'
const view = ref(auth.can('purchase.manage') ? 'purchase' : 'sale')
const rows = ref([])
const meta = ref({ page: 1, total: 0, total_pages: 1 })
const page = ref(1)
const loading = ref(false)

const suppliers = ref([])
const customers = ref([])
const products = ref([])

const showModal = ref(false)
const mode = ref('purchase') // create mode
const saving = ref(false)
const formError = ref('')
const form = reactive({ party_id: '', note: '', items: [] })

const parties = computed(() => (mode.value === 'purchase' ? suppliers.value : customers.value))
const total = computed(() =>
  form.items.reduce((s, it) => s + Number(it.quantity || 0) * Number(it.unit_value || 0), 0),
)

async function load() {
  loading.value = true
  try {
    const endpoint = view.value === 'purchase' ? '/returns/purchase' : '/returns/sale'
    const { data } = await api.get(endpoint, { params: { page: page.value, per_page: 10 } })
    rows.value = data.data
    meta.value = data.meta
  } finally {
    loading.value = false
  }
}

function switchView(v) {
  view.value = v
  page.value = 1
  load()
}

const newRow = () => ({ product_id: products.value[0]?.id || '', quantity: 1, unit_value: 0 })

function open(m) {
  mode.value = m
  form.party_id = parties.value[0]?.id || ''
  form.note = ''
  form.items = [newRow()]
  formError.value = ''
  showModal.value = true
}
const addRow = () => form.items.push(newRow())
const removeRow = (i) => form.items.splice(i, 1)

async function save() {
  saving.value = true
  formError.value = ''
  const isPurchase = mode.value === 'purchase'
  const endpoint = isPurchase ? '/returns/purchase' : '/returns/sale'
  const idKey = isPurchase ? 'supplier_id' : 'customer_id'
  try {
    await api.post(endpoint, {
      [idKey]: Number(form.party_id),
      note: form.note,
      items: form.items.map((it) => ({
        product_id: Number(it.product_id),
        quantity: Number(it.quantity),
        unit_value: Number(it.unit_value),
      })),
    })
    showModal.value = false
    view.value = mode.value
    page.value = 1
    load()
  } catch (e) {
    formError.value = e.response?.data?.message || 'Failed to save return'
  } finally {
    saving.value = false
  }
}

onMounted(async () => {
  const [sup, cus, prod] = await Promise.all([
    api.get('/suppliers', { params: { per_page: 100 } }),
    api.get('/customers', { params: { per_page: 100 } }),
    api.get('/products', { params: { per_page: 100 } }),
  ])
  suppliers.value = sup.data.data
  customers.value = cus.data.data
  products.value = prod.data.data
  load()
})
</script>

<template>
  <div>
    <div class="mb-6 flex flex-wrap items-center justify-between gap-3">
      <h1 class="text-2xl font-bold">Returns</h1>
      <div class="flex gap-2">
        <button v-if="auth.can('purchase.manage')" class="btn-ghost" @click="open('purchase')">📤 Purchase Return</button>
        <button v-if="auth.can('sales.manage')" class="btn-primary" @click="open('sale')">📥 Sales Return</button>
      </div>
    </div>

    <!-- View toggle -->
    <div class="mb-4 flex gap-2">
      <button v-if="auth.can('purchase.manage')" class="btn" :class="view === 'purchase' ? 'bg-brand-600 text-white' : 'btn-ghost'" @click="switchView('purchase')">Purchase Returns</button>
      <button v-if="auth.can('sales.manage')" class="btn" :class="view === 'sale' ? 'bg-brand-600 text-white' : 'btn-ghost'" @click="switchView('sale')">Sales Returns</button>
    </div>

    <div class="card overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-left text-sm">
          <thead class="border-b border-slate-200 bg-slate-50 text-xs uppercase text-slate-500 dark:border-slate-700 dark:bg-slate-700/40">
            <tr>
              <th class="px-4 py-3">Invoice</th>
              <th class="px-4 py-3">{{ view === 'purchase' ? 'Supplier' : 'Customer' }}</th>
              <th class="px-4 py-3">Amount</th><th class="px-4 py-3">Note</th><th class="px-4 py-3">Date</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 dark:divide-slate-700">
            <tr v-if="loading"><td colspan="5" class="px-4 py-10 text-center text-slate-400">Loading…</td></tr>
            <tr v-else-if="!rows.length"><td colspan="5" class="px-4 py-10 text-center text-slate-400">No returns yet</td></tr>
            <tr v-for="r in rows" :key="r.id" class="hover:bg-slate-50 dark:hover:bg-slate-700/30">
              <td class="px-4 py-3 font-medium">{{ r.invoice_no }}</td>
              <td class="px-4 py-3">{{ (r.supplier || r.customer)?.name }}</td>
              <td class="px-4 py-3 font-semibold">{{ money(r.total_amount) }}</td>
              <td class="px-4 py-3 text-slate-400">{{ r.note || '—' }}</td>
              <td class="px-4 py-3 text-slate-400">{{ new Date(r.created_at).toLocaleDateString() }}</td>
            </tr>
          </tbody>
        </table>
      </div>
      <div class="flex items-center justify-between border-t border-slate-200 px-4 py-3 text-sm dark:border-slate-700">
        <span class="text-slate-500">Page {{ meta.page }} of {{ meta.total_pages || 1 }} · {{ meta.total }} total</span>
        <div class="flex gap-2">
          <button class="btn-ghost !py-1" :disabled="page <= 1" @click="page--; load()">Prev</button>
          <button class="btn-ghost !py-1" :disabled="page >= meta.total_pages" @click="page++; load()">Next</button>
        </div>
      </div>
    </div>

    <Modal v-if="showModal" :title="mode === 'purchase' ? 'Purchase Return' : 'Sales Return'" @close="showModal = false">
      <form class="space-y-4" @submit.prevent="save">
        <div v-if="formError" class="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600 dark:bg-red-500/10">{{ formError }}</div>

        <div>
          <label class="label">{{ mode === 'purchase' ? 'Supplier' : 'Customer' }}</label>
          <select v-model="form.party_id" class="input">
            <option v-for="x in parties" :key="x.id" :value="x.id">{{ x.name }}</option>
          </select>
        </div>

        <div>
          <div class="mb-1 flex items-center justify-between">
            <label class="label !mb-0">Items</label>
            <button type="button" class="btn-ghost !px-2 !py-1 text-xs" @click="addRow">+ Add Item</button>
          </div>
          <div class="space-y-2">
            <div v-for="(it, i) in form.items" :key="i" class="flex items-center gap-2">
              <select v-model="it.product_id" class="input flex-1">
                <option v-for="p in products" :key="p.id" :value="p.id">{{ p.name }} (stock: {{ p.quantity }})</option>
              </select>
              <input v-model.number="it.quantity" type="number" min="1" class="input w-20" placeholder="Qty" />
              <input v-model.number="it.unit_value" type="number" min="0" class="input w-28" :placeholder="mode === 'purchase' ? 'Cost' : 'Price'" />
              <button type="button" class="btn-danger !px-2 !py-1 text-xs" :disabled="form.items.length === 1" @click="removeRow(i)">✕</button>
            </div>
          </div>
        </div>

        <div class="flex items-center justify-between rounded-lg bg-slate-50 px-3 py-2 dark:bg-slate-700/50">
          <span class="text-sm text-slate-500">Total Return Amount</span>
          <span class="font-semibold">{{ money(total) }}</span>
        </div>

        <div>
          <label class="label">Note</label>
          <input v-model="form.note" class="input" placeholder="Reason for return (optional)" />
        </div>

        <div class="flex justify-end gap-2 pt-2">
          <button type="button" class="btn-ghost" @click="showModal = false">Cancel</button>
          <button type="submit" class="btn-primary" :disabled="saving">{{ saving ? 'Saving…' : 'Save Return' }}</button>
        </div>
      </form>
    </Modal>
  </div>
</template>
