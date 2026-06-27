<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import api from '../lib/api'
import Modal from '../components/Modal.vue'

const money = (n) => '৳' + Number(n || 0).toLocaleString('en-IN')

const sales = ref([])
const meta = ref({ page: 1, total: 0, total_pages: 1 })
const page = ref(1)
const loading = ref(false)

const customers = ref([])
const products = ref([])
const productMap = computed(() => Object.fromEntries(products.value.map((p) => [p.id, p])))

const showModal = ref(false)
const saving = ref(false)
const formError = ref('')
const form = reactive({ customer_id: '', paid_amount: 0, note: '', items: [] })

const total = computed(() =>
  form.items.reduce((s, it) => s + Number(it.quantity || 0) * Number(it.unit_price || 0), 0),
)

async function load() {
  loading.value = true
  try {
    const { data } = await api.get('/sales', { params: { page: page.value, per_page: 10 } })
    sales.value = data.data
    meta.value = data.meta
  } finally {
    loading.value = false
  }
}

const newRow = () => {
  const p = products.value[0]
  return { product_id: p?.id || '', quantity: 1, unit_price: p?.price || 0 }
}

function openCreate() {
  form.customer_id = customers.value[0]?.id || ''
  form.paid_amount = 0
  form.note = ''
  form.items = [newRow()]
  formError.value = ''
  showModal.value = true
}
const addRow = () => form.items.push(newRow())
const removeRow = (i) => form.items.splice(i, 1)

// When the product changes, prefill the selling price from the product.
function onProductChange(it) {
  const p = productMap.value[it.product_id]
  if (p) it.unit_price = p.price
}

async function save() {
  saving.value = true
  formError.value = ''
  try {
    await api.post('/sales', {
      customer_id: Number(form.customer_id),
      paid_amount: Number(form.paid_amount),
      note: form.note,
      items: form.items.map((it) => ({
        product_id: Number(it.product_id),
        quantity: Number(it.quantity),
        unit_price: Number(it.unit_price),
      })),
    })
    showModal.value = false
    page.value = 1
    load()
  } catch (e) {
    formError.value = e.response?.data?.message || 'Failed to save sale'
  } finally {
    saving.value = false
  }
}

onMounted(async () => {
  const [cus, prod] = await Promise.all([
    api.get('/customers', { params: { per_page: 100 } }),
    api.get('/products', { params: { per_page: 100 } }),
  ])
  customers.value = cus.data.data
  products.value = prod.data.data
  load()
})
</script>

<template>
  <div>
    <div class="mb-6 flex items-center justify-between">
      <h1 class="text-2xl font-bold">Sales</h1>
      <button class="btn-primary" @click="openCreate">+ New Sale</button>
    </div>

    <div class="card overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-left text-sm">
          <thead class="border-b border-slate-200 bg-slate-50 text-xs uppercase text-slate-500 dark:border-slate-700 dark:bg-slate-700/40">
            <tr>
              <th class="px-4 py-3">Invoice</th><th class="px-4 py-3">Customer</th>
              <th class="px-4 py-3">Total</th><th class="px-4 py-3">Paid</th>
              <th class="px-4 py-3">Due</th><th class="px-4 py-3">Date</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 dark:divide-slate-700">
            <tr v-if="loading"><td colspan="6" class="px-4 py-10 text-center text-slate-400">Loading…</td></tr>
            <tr v-else-if="!sales.length"><td colspan="6" class="px-4 py-10 text-center text-slate-400">No sales yet</td></tr>
            <tr v-for="s in sales" :key="s.id" class="hover:bg-slate-50 dark:hover:bg-slate-700/30">
              <td class="px-4 py-3 font-medium">{{ s.invoice_no }}</td>
              <td class="px-4 py-3">{{ s.customer?.name }}</td>
              <td class="px-4 py-3">{{ money(s.total_amount) }}</td>
              <td class="px-4 py-3 text-emerald-600">{{ money(s.paid_amount) }}</td>
              <td class="px-4 py-3">
                <span v-if="s.due > 0" class="badge bg-amber-100 text-amber-700 dark:bg-amber-500/20 dark:text-amber-300">{{ money(s.due) }}</span>
                <span v-else class="badge bg-emerald-100 text-emerald-700 dark:bg-emerald-500/20 dark:text-emerald-300">Paid</span>
              </td>
              <td class="px-4 py-3 text-slate-400">{{ new Date(s.created_at).toLocaleDateString() }}</td>
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

    <Modal v-if="showModal" title="New Sale" @close="showModal = false">
      <form class="space-y-4" @submit.prevent="save">
        <div v-if="formError" class="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600 dark:bg-red-500/10">{{ formError }}</div>

        <div>
          <label class="label">Customer</label>
          <select v-model="form.customer_id" class="input">
            <option v-for="c in customers" :key="c.id" :value="c.id">{{ c.name }}</option>
          </select>
        </div>

        <div>
          <div class="mb-1 flex items-center justify-between">
            <label class="label !mb-0">Items</label>
            <button type="button" class="btn-ghost !px-2 !py-1 text-xs" @click="addRow">+ Add Item</button>
          </div>
          <div class="space-y-2">
            <div v-for="(it, i) in form.items" :key="i" class="flex items-center gap-2">
              <select v-model="it.product_id" class="input flex-1" @change="onProductChange(it)">
                <option v-for="p in products" :key="p.id" :value="p.id">
                  {{ p.name }} (stock: {{ p.quantity }})
                </option>
              </select>
              <input v-model.number="it.quantity" type="number" min="1" class="input w-20" placeholder="Qty" />
              <input v-model.number="it.unit_price" type="number" min="0" class="input w-28" placeholder="Price" />
              <button type="button" class="btn-danger !px-2 !py-1 text-xs" :disabled="form.items.length === 1" @click="removeRow(i)">✕</button>
            </div>
          </div>
        </div>

        <div class="grid grid-cols-2 gap-3">
          <div>
            <label class="label">Paid Amount</label>
            <input v-model.number="form.paid_amount" type="number" min="0" class="input" />
          </div>
          <div>
            <label class="label">Total</label>
            <div class="input bg-slate-50 font-semibold dark:bg-slate-700/50">{{ money(total) }}</div>
          </div>
        </div>

        <div>
          <label class="label">Note</label>
          <input v-model="form.note" class="input" placeholder="Optional" />
        </div>

        <div class="flex justify-end gap-2 pt-2">
          <button type="button" class="btn-ghost" @click="showModal = false">Cancel</button>
          <button type="submit" class="btn-primary" :disabled="saving">{{ saving ? 'Saving…' : 'Save Sale' }}</button>
        </div>
      </form>
    </Modal>
  </div>
</template>
