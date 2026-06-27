<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import api from '../lib/api'
import Modal from '../components/Modal.vue'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const money = (n) => '৳' + Number(n || 0).toLocaleString('en-IN')

const payments = ref([])
const meta = ref({ page: 1, total: 0, total_pages: 1 })
const page = ref(1)
const typeFilter = ref('')
const loading = ref(false)

const customers = ref([])
const suppliers = ref([])

const showModal = ref(false)
const mode = ref('customer') // 'customer' | 'supplier'
const saving = ref(false)
const formError = ref('')
const form = reactive({ party_id: '', amount: 0, method: 'cash', note: '' })

// The party list + the currently-selected party's due, for the modal.
const parties = computed(() => (mode.value === 'customer' ? customers.value : suppliers.value))
const selectedDue = computed(() => {
  const p = parties.value.find((x) => x.id === Number(form.party_id))
  return p ? p.due : 0
})

async function load() {
  loading.value = true
  try {
    const { data } = await api.get('/payments', { params: { page: page.value, per_page: 10, type: typeFilter.value } })
    payments.value = data.data
    meta.value = data.meta
  } finally {
    loading.value = false
  }
}

function open(m) {
  mode.value = m
  form.party_id = parties.value[0]?.id || ''
  form.amount = 0
  form.method = 'cash'
  form.note = ''
  formError.value = ''
  showModal.value = true
}

async function save() {
  saving.value = true
  formError.value = ''
  const endpoint = mode.value === 'customer' ? '/payments/customer' : '/payments/supplier'
  const idKey = mode.value === 'customer' ? 'customer_id' : 'supplier_id'
  try {
    await api.post(endpoint, {
      [idKey]: Number(form.party_id),
      amount: Number(form.amount),
      method: form.method,
      note: form.note,
    })
    showModal.value = false
    page.value = 1
    await refreshParties()
    load()
  } catch (e) {
    formError.value = e.response?.data?.message || 'Failed to record payment'
  } finally {
    saving.value = false
  }
}

async function refreshParties() {
  const [cus, sup] = await Promise.all([
    api.get('/customers', { params: { per_page: 100 } }),
    api.get('/suppliers', { params: { per_page: 100 } }),
  ])
  customers.value = cus.data.data
  suppliers.value = sup.data.data
}

onMounted(async () => {
  await refreshParties()
  load()
})
</script>

<template>
  <div>
    <div class="mb-6 flex flex-wrap items-center justify-between gap-3">
      <h1 class="text-2xl font-bold">Payments</h1>
      <div class="flex gap-2">
        <button v-if="auth.can('sales.manage')" class="btn-primary" @click="open('customer')">📥 Receive from Customer</button>
        <button v-if="auth.can('purchase.manage')" class="btn-ghost" @click="open('supplier')">📤 Pay Supplier</button>
      </div>
    </div>

    <!-- Filter -->
    <div class="mb-4 flex gap-2">
      <button class="btn" :class="typeFilter === '' ? 'bg-brand-600 text-white' : 'btn-ghost'" @click="typeFilter = ''; page = 1; load()">All</button>
      <button class="btn" :class="typeFilter === 'customer' ? 'bg-brand-600 text-white' : 'btn-ghost'" @click="typeFilter = 'customer'; page = 1; load()">Customer Receipts</button>
      <button class="btn" :class="typeFilter === 'supplier' ? 'bg-brand-600 text-white' : 'btn-ghost'" @click="typeFilter = 'supplier'; page = 1; load()">Supplier Payments</button>
    </div>

    <div class="card overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-left text-sm">
          <thead class="border-b border-slate-200 bg-slate-50 text-xs uppercase text-slate-500 dark:border-slate-700 dark:bg-slate-700/40">
            <tr>
              <th class="px-4 py-3">Type</th><th class="px-4 py-3">Party</th>
              <th class="px-4 py-3">Amount</th><th class="px-4 py-3">Method</th>
              <th class="px-4 py-3">Note</th><th class="px-4 py-3">Date</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 dark:divide-slate-700">
            <tr v-if="loading"><td colspan="6" class="px-4 py-10 text-center text-slate-400">Loading…</td></tr>
            <tr v-else-if="!payments.length"><td colspan="6" class="px-4 py-10 text-center text-slate-400">No payments yet</td></tr>
            <tr v-for="p in payments" :key="p.id" class="hover:bg-slate-50 dark:hover:bg-slate-700/30">
              <td class="px-4 py-3">
                <span v-if="p.party_type === 'customer'" class="badge bg-emerald-100 text-emerald-700 dark:bg-emerald-500/20 dark:text-emerald-300">Received</span>
                <span v-else class="badge bg-blue-100 text-blue-700 dark:bg-blue-500/20 dark:text-blue-300">Paid</span>
              </td>
              <td class="px-4 py-3 font-medium">{{ p.party_name }}</td>
              <td class="px-4 py-3 font-semibold">{{ money(p.amount) }}</td>
              <td class="px-4 py-3 capitalize text-slate-500">{{ p.method }}</td>
              <td class="px-4 py-3 text-slate-400">{{ p.note || '—' }}</td>
              <td class="px-4 py-3 text-slate-400">{{ new Date(p.created_at).toLocaleDateString() }}</td>
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

    <Modal v-if="showModal" :title="mode === 'customer' ? 'Receive from Customer' : 'Pay Supplier'" @close="showModal = false">
      <form class="space-y-4" @submit.prevent="save">
        <div v-if="formError" class="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600 dark:bg-red-500/10">{{ formError }}</div>

        <div>
          <label class="label">{{ mode === 'customer' ? 'Customer' : 'Supplier' }}</label>
          <select v-model="form.party_id" class="input">
            <option v-for="x in parties" :key="x.id" :value="x.id">{{ x.name }} (due: {{ money(x.due) }})</option>
          </select>
          <p class="mt-1 text-xs text-amber-600">Outstanding due: {{ money(selectedDue) }}</p>
        </div>

        <div class="grid grid-cols-2 gap-3">
          <div>
            <label class="label">Amount</label>
            <input v-model.number="form.amount" type="number" min="1" class="input" />
          </div>
          <div>
            <label class="label">Method</label>
            <select v-model="form.method" class="input">
              <option value="cash">Cash</option><option value="bank">Bank</option><option value="mobile">Mobile</option>
            </select>
          </div>
        </div>

        <div>
          <label class="label">Note</label>
          <input v-model="form.note" class="input" placeholder="Optional" />
        </div>

        <div class="flex justify-end gap-2 pt-2">
          <button type="button" class="btn-ghost" @click="showModal = false">Cancel</button>
          <button type="submit" class="btn-primary" :disabled="saving">{{ saving ? 'Saving…' : 'Save Payment' }}</button>
        </div>
      </form>
    </Modal>
  </div>
</template>
