<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import api, { assetUrl } from '../lib/api'
import Modal from '../components/Modal.vue'
import ProductPicker from '../components/ProductPicker.vue'
import InvoiceModal from '../components/InvoiceModal.vue'

const money = (n) => '৳' + Number(n || 0).toLocaleString('en-IN')

const sales = ref([])
const meta = ref({ page: 1, total: 0, total_pages: 1 })
const page = ref(1)
const loading = ref(false)

const customers = ref([])
const products = ref([])

const invoiceId = ref(null)
const selectedCustomer = computed(() => customers.value.find((c) => c.id === Number(form.customer_id)))

const showModal = ref(false)
const saving = ref(false)
const formError = ref('')
const form = reactive({ customer_id: '', discount: 0, tax_percent: 0, paid_amount: 0, note: '', items: [] })

const subtotal = computed(() => form.items.reduce((s, it) => s + Number(it.quantity || 0) * Number(it.unit_price || 0), 0))
const discount = computed(() => Math.min(Math.max(Number(form.discount) || 0, 0), subtotal.value))
const taxable = computed(() => subtotal.value - discount.value)
const taxAmount = computed(() => (taxable.value * (Number(form.tax_percent) || 0)) / 100)
const grandTotal = computed(() => taxable.value + taxAmount.value)
const due = computed(() => grandTotal.value - (Number(form.paid_amount) || 0))

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

function openCreate() {
  form.customer_id = customers.value[0]?.id || ''
  form.discount = 0
  form.tax_percent = 0
  form.paid_amount = 0
  form.note = ''
  form.items = []
  formError.value = ''
  showModal.value = true
}

function addProduct(p) {
  const existing = form.items.find((it) => it.product_id === p.id)
  if (existing) {
    existing.quantity++
    return
  }
  form.items.push({ product_id: p.id, name: p.name, sku: p.sku, image: p.image, stock: p.quantity, quantity: 1, unit_price: p.price || 0 })
}
const removeRow = (i) => form.items.splice(i, 1)

async function voidSale(s) {
  if (!confirm(`Void sale ${s.invoice_no}? This returns its stock and reverses the customer due.`)) return
  await api.delete(`/sales/${s.id}`)
  if (sales.value.length === 1 && page.value > 1) page.value--
  load()
}

async function save() {
  if (!form.items.length) {
    formError.value = 'Add at least one product'
    return
  }
  saving.value = true
  formError.value = ''
  try {
    await api.post('/sales', {
      customer_id: Number(form.customer_id),
      discount: discount.value,
      tax_percent: Number(form.tax_percent) || 0,
      paid_amount: Number(form.paid_amount) || 0,
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
              <th class="px-4 py-3 text-right">Actions</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 dark:divide-slate-700">
            <tr v-if="loading"><td colspan="7" class="px-4 py-10 text-center text-slate-400">Loading…</td></tr>
            <tr v-else-if="!sales.length"><td colspan="7" class="px-4 py-10 text-center text-slate-400">No sales yet</td></tr>
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
              <td class="px-4 py-3 text-right">
                <button class="btn-ghost !px-2 !py-1 text-xs" @click="invoiceId = s.id">Invoice</button>
                <button class="btn-danger !px-2 !py-1 text-xs" @click="voidSale(s)">Void</button>
              </td>
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

    <Modal v-if="showModal" title="New Sale" size="max-w-4xl" @close="showModal = false">
      <form class="space-y-5" @submit.prevent="save">
        <div v-if="formError" class="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600 dark:bg-red-500/10">{{ formError }}</div>

        <div class="grid gap-4 sm:grid-cols-2">
          <div>
            <label class="label">Customer</label>
            <select v-model="form.customer_id" class="input">
              <option v-for="c in customers" :key="c.id" :value="c.id">{{ c.name }}</option>
            </select>
            <div v-if="selectedCustomer" class="mt-2 rounded-lg bg-slate-50 px-3 py-2 text-xs text-slate-500 dark:bg-slate-700/40">
              <span v-if="selectedCustomer.phone">📞 {{ selectedCustomer.phone }}</span>
              <span v-if="selectedCustomer.address"> · 📍 {{ selectedCustomer.address }}</span>
              <span v-if="selectedCustomer.due > 0" class="text-amber-600"> · Due: {{ money(selectedCustomer.due) }}</span>
            </div>
          </div>
          <div>
            <label class="label">Add Product</label>
            <ProductPicker :products="products" @select="addProduct" />
          </div>
        </div>

        <div class="overflow-hidden rounded-xl border border-slate-200 dark:border-slate-700">
          <table class="w-full text-left text-sm">
            <thead class="bg-slate-50 text-xs uppercase text-slate-500 dark:bg-slate-700/40">
              <tr>
                <th class="px-3 py-2">Product</th>
                <th class="w-24 px-3 py-2">Qty</th>
                <th class="w-32 px-3 py-2">Unit Price</th>
                <th class="w-32 px-3 py-2 text-right">Subtotal</th>
                <th class="w-10 px-3 py-2"></th>
              </tr>
            </thead>
            <tbody class="divide-y divide-slate-100 dark:divide-slate-700">
              <tr v-if="!form.items.length"><td colspan="5" class="px-3 py-6 text-center text-slate-400">Search and add products above</td></tr>
              <tr v-for="(it, i) in form.items" :key="it.product_id">
                <td class="px-3 py-2">
                  <div class="flex items-center gap-2">
                    <img v-if="it.image" :src="assetUrl(it.image)" class="h-8 w-8 rounded border border-slate-200 object-cover" />
                    <span v-else class="grid h-8 w-8 place-items-center rounded bg-slate-100 text-[9px] text-slate-400 dark:bg-slate-600">IMG</span>
                    <div>
                      <div class="font-medium">{{ it.name }}</div>
                      <div class="text-xs" :class="it.quantity > it.stock ? 'text-red-500' : 'text-slate-400'">{{ it.sku }} · stock {{ it.stock }}</div>
                    </div>
                  </div>
                </td>
                <td class="px-3 py-2"><input v-model.number="it.quantity" type="number" min="1" :max="it.stock" class="input !py-1.5" /></td>
                <td class="px-3 py-2"><input v-model.number="it.unit_price" type="number" min="0" class="input !py-1.5" /></td>
                <td class="px-3 py-2 text-right font-medium">{{ money(it.quantity * it.unit_price) }}</td>
                <td class="px-3 py-2"><button type="button" class="btn-danger !px-2 !py-1 text-xs" @click="removeRow(i)">✕</button></td>
              </tr>
            </tbody>
          </table>
        </div>

        <div class="grid gap-5 sm:grid-cols-2">
          <div>
            <label class="label">Note</label>
            <textarea v-model="form.note" rows="3" class="input" placeholder="Optional remark / reference"></textarea>
          </div>

          <div class="space-y-2 rounded-xl bg-slate-50 p-4 text-sm dark:bg-slate-700/40">
            <div class="flex items-center justify-between">
              <span class="text-slate-500">Subtotal</span><span class="font-medium">{{ money(subtotal) }}</span>
            </div>
            <div class="flex items-center justify-between">
              <span class="text-slate-500">Discount</span>
              <input v-model.number="form.discount" type="number" min="0" class="input !w-28 !py-1 text-right" />
            </div>
            <div class="flex items-center justify-between">
              <span class="text-slate-500">VAT / Tax (%)</span>
              <input v-model.number="form.tax_percent" type="number" min="0" class="input !w-28 !py-1 text-right" />
            </div>
            <div class="flex items-center justify-between text-slate-500">
              <span>Tax Amount</span><span>{{ money(taxAmount) }}</span>
            </div>
            <div class="flex items-center justify-between border-t border-slate-200 pt-2 text-base font-bold dark:border-slate-600">
              <span>Grand Total</span><span class="text-brand-600">{{ money(grandTotal) }}</span>
            </div>
            <div class="flex items-center justify-between">
              <span class="text-slate-500">Paid</span>
              <input v-model.number="form.paid_amount" type="number" min="0" class="input !w-28 !py-1 text-right" />
            </div>
            <div class="flex items-center justify-between font-semibold">
              <span class="text-slate-500">Due</span>
              <span :class="due > 0 ? 'text-amber-600' : 'text-emerald-600'">{{ money(due) }}</span>
            </div>
          </div>
        </div>

        <div class="flex justify-end gap-2">
          <button type="button" class="btn-ghost" @click="showModal = false">Cancel</button>
          <button type="submit" class="btn-primary" :disabled="saving">{{ saving ? 'Saving…' : 'Save Sale' }}</button>
        </div>
      </form>
    </Modal>

    <InvoiceModal v-if="invoiceId" type="sale" :id="invoiceId" @close="invoiceId = null" />
  </div>
</template>
