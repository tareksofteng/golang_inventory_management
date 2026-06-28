<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import api, { assetUrl } from '../lib/api'
import Modal from '../components/Modal.vue'
import ProductPicker from '../components/ProductPicker.vue'
import InvoiceModal from '../components/InvoiceModal.vue'

const money = (n) => '৳' + Number(n || 0).toLocaleString('en-IN')

const purchases = ref([])
const meta = ref({ page: 1, total: 0, total_pages: 1 })
const page = ref(1)
const loading = ref(false)

const suppliers = ref([])
const products = ref([])

const invoiceId = ref(null) // open the invoice modal for this purchase
const selectedSupplier = computed(() => suppliers.value.find((s) => s.id === Number(form.supplier_id)))

const showModal = ref(false)
const saving = ref(false)
const formError = ref('')
const form = reactive({ supplier_id: '', discount: 0, tax_percent: 0, paid_amount: 0, note: '', items: [] })

// ---- Accounting calculations ----
const subtotal = computed(() => form.items.reduce((s, it) => s + Number(it.quantity || 0) * Number(it.unit_cost || 0), 0))
const discount = computed(() => Math.min(Math.max(Number(form.discount) || 0, 0), subtotal.value))
const taxable = computed(() => subtotal.value - discount.value)
const taxAmount = computed(() => (taxable.value * (Number(form.tax_percent) || 0)) / 100)
const grandTotal = computed(() => taxable.value + taxAmount.value)
const due = computed(() => grandTotal.value - (Number(form.paid_amount) || 0))

async function load() {
  loading.value = true
  try {
    const { data } = await api.get('/purchases', { params: { page: page.value, per_page: 10 } })
    purchases.value = data.data
    meta.value = data.meta
  } finally {
    loading.value = false
  }
}

function openCreate() {
  form.supplier_id = suppliers.value[0]?.id || ''
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
  form.items.push({ product_id: p.id, name: p.name, sku: p.sku, image: p.image, quantity: 1, unit_cost: p.cost_price || 0 })
}
const removeRow = (i) => form.items.splice(i, 1)

async function voidPurchase(p) {
  if (!confirm(`Void purchase ${p.invoice_no}? This reverses its stock and supplier due.`)) return
  await api.delete(`/purchases/${p.id}`)
  if (purchases.value.length === 1 && page.value > 1) page.value--
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
    await api.post('/purchases', {
      supplier_id: Number(form.supplier_id),
      discount: discount.value,
      tax_percent: Number(form.tax_percent) || 0,
      paid_amount: Number(form.paid_amount) || 0,
      note: form.note,
      items: form.items.map((it) => ({
        product_id: Number(it.product_id),
        quantity: Number(it.quantity),
        unit_cost: Number(it.unit_cost),
      })),
    })
    showModal.value = false
    page.value = 1
    load()
  } catch (e) {
    formError.value = e.response?.data?.message || 'Failed to save purchase'
  } finally {
    saving.value = false
  }
}

onMounted(async () => {
  const [sup, prod] = await Promise.all([
    api.get('/suppliers', { params: { per_page: 100 } }),
    api.get('/products', { params: { per_page: 100 } }),
  ])
  suppliers.value = sup.data.data
  products.value = prod.data.data
  load()
})
</script>

<template>
  <div>
    <div class="mb-6 flex items-center justify-between">
      <h1 class="text-2xl font-bold">Purchases</h1>
      <button class="btn-primary" @click="openCreate">+ New Purchase</button>
    </div>

    <div class="card overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-left text-sm">
          <thead class="border-b border-slate-200 bg-slate-50 text-xs uppercase text-slate-500 dark:border-slate-700 dark:bg-slate-700/40">
            <tr>
              <th class="px-4 py-3">Invoice</th><th class="px-4 py-3">Supplier</th>
              <th class="px-4 py-3">Total</th><th class="px-4 py-3">Paid</th>
              <th class="px-4 py-3">Due</th><th class="px-4 py-3">Date</th>
              <th class="px-4 py-3 text-right">Actions</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 dark:divide-slate-700">
            <tr v-if="loading"><td colspan="7" class="px-4 py-10 text-center text-slate-400">Loading…</td></tr>
            <tr v-else-if="!purchases.length"><td colspan="7" class="px-4 py-10 text-center text-slate-400">No purchases yet</td></tr>
            <tr v-for="p in purchases" :key="p.id" class="hover:bg-slate-50 dark:hover:bg-slate-700/30">
              <td class="px-4 py-3 font-medium">{{ p.invoice_no }}</td>
              <td class="px-4 py-3">{{ p.supplier?.name }}</td>
              <td class="px-4 py-3">{{ money(p.total_amount) }}</td>
              <td class="px-4 py-3 text-emerald-600">{{ money(p.paid_amount) }}</td>
              <td class="px-4 py-3">
                <span v-if="p.due > 0" class="badge bg-amber-100 text-amber-700 dark:bg-amber-500/20 dark:text-amber-300">{{ money(p.due) }}</span>
                <span v-else class="badge bg-emerald-100 text-emerald-700 dark:bg-emerald-500/20 dark:text-emerald-300">Paid</span>
              </td>
              <td class="px-4 py-3 text-slate-400">{{ new Date(p.created_at).toLocaleDateString() }}</td>
              <td class="px-4 py-3 text-right">
                <button class="btn-ghost !px-2 !py-1 text-xs" @click="invoiceId = p.id">Invoice</button>
                <button class="btn-danger !px-2 !py-1 text-xs" @click="voidPurchase(p)">Void</button>
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

    <!-- New purchase modal (enterprise layout) -->
    <Modal v-if="showModal" title="New Purchase" size="max-w-4xl" @close="showModal = false">
      <form class="space-y-5" @submit.prevent="save">
        <div v-if="formError" class="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600 dark:bg-red-500/10">{{ formError }}</div>

        <!-- Supplier + product search -->
        <div class="grid gap-4 sm:grid-cols-2">
          <div>
            <label class="label">Supplier</label>
            <select v-model="form.supplier_id" class="input">
              <option v-for="s in suppliers" :key="s.id" :value="s.id">{{ s.name }}</option>
            </select>
            <div v-if="selectedSupplier" class="mt-2 rounded-lg bg-slate-50 px-3 py-2 text-xs text-slate-500 dark:bg-slate-700/40">
              <span v-if="selectedSupplier.phone">📞 {{ selectedSupplier.phone }}</span>
              <span v-if="selectedSupplier.address"> · 📍 {{ selectedSupplier.address }}</span>
              <span v-if="selectedSupplier.due > 0" class="text-amber-600"> · Due: {{ money(selectedSupplier.due) }}</span>
            </div>
          </div>
          <div>
            <label class="label">Add Product</label>
            <ProductPicker :products="products" @select="addProduct" />
          </div>
        </div>

        <!-- Items table -->
        <div class="overflow-hidden rounded-xl border border-slate-200 dark:border-slate-700">
          <table class="w-full text-left text-sm">
            <thead class="bg-slate-50 text-xs uppercase text-slate-500 dark:bg-slate-700/40">
              <tr>
                <th class="px-3 py-2">Product</th>
                <th class="w-24 px-3 py-2">Qty</th>
                <th class="w-32 px-3 py-2">Unit Cost</th>
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
                      <div class="text-xs text-slate-400">{{ it.sku }}</div>
                    </div>
                  </div>
                </td>
                <td class="px-3 py-2"><input v-model.number="it.quantity" type="number" min="1" class="input !py-1.5" /></td>
                <td class="px-3 py-2"><input v-model.number="it.unit_cost" type="number" min="0" class="input !py-1.5" /></td>
                <td class="px-3 py-2 text-right font-medium">{{ money(it.quantity * it.unit_cost) }}</td>
                <td class="px-3 py-2"><button type="button" class="btn-danger !px-2 !py-1 text-xs" @click="removeRow(i)">✕</button></td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Summary + payment -->
        <div class="grid gap-5 sm:grid-cols-2">
          <div>
            <label class="label">Note</label>
            <textarea v-model="form.note" rows="3" class="input" placeholder="Optional remark / reference"></textarea>
          </div>

          <div class="space-y-2 rounded-xl bg-slate-50 p-4 text-sm dark:bg-slate-700/40">
            <div class="flex items-center justify-between">
              <span class="text-slate-500">Subtotal</span>
              <span class="font-medium">{{ money(subtotal) }}</span>
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
              <span>Tax Amount</span>
              <span>{{ money(taxAmount) }}</span>
            </div>
            <div class="flex items-center justify-between border-t border-slate-200 pt-2 text-base font-bold dark:border-slate-600">
              <span>Grand Total</span>
              <span class="text-brand-600">{{ money(grandTotal) }}</span>
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
          <button type="submit" class="btn-primary" :disabled="saving">{{ saving ? 'Saving…' : 'Save Purchase' }}</button>
        </div>
      </form>
    </Modal>

    <InvoiceModal v-if="invoiceId" type="purchase" :id="invoiceId" @close="invoiceId = null" />
  </div>
</template>
