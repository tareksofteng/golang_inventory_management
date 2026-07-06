<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import api from '../lib/api'
import Modal from '../components/Modal.vue'

const money = (n) => '৳' + Number(n || 0).toLocaleString('en-IN', { minimumFractionDigits: 2 })

const entries = ref([])
const meta = ref({ page: 1, total: 0, total_pages: 1 })
const page = ref(1)
const loading = ref(false)
const accounts = ref([])

const showModal = ref(false)
const saving = ref(false)
const formError = ref('')
const form = reactive({ date: '', reference: '', note: '', lines: [] })

const viewing = ref(null)

const totalDebit = computed(() => form.lines.reduce((s, l) => s + Number(l.debit || 0), 0))
const totalCredit = computed(() => form.lines.reduce((s, l) => s + Number(l.credit || 0), 0))
const diff = computed(() => totalDebit.value - totalCredit.value)
const balanced = computed(() => Math.abs(diff.value) < 0.005 && totalDebit.value > 0)

const entryTotal = (e) => (e.lines || []).reduce((s, l) => s + Number(l.debit || 0), 0)
const today = () => new Date().toISOString().slice(0, 10)

async function load() {
  loading.value = true
  try {
    const { data } = await api.get('/journal', { params: { page: page.value, per_page: 10 } })
    entries.value = data.data
    meta.value = data.meta
  } finally {
    loading.value = false
  }
}

const newLine = () => ({ account_id: accounts.value[0]?.id || '', debit: 0, credit: 0 })

function openCreate() {
  form.date = today()
  form.reference = ''
  form.note = ''
  form.lines = [newLine(), newLine()]
  formError.value = ''
  showModal.value = true
}
const addLine = () => form.lines.push(newLine())
const removeLine = (i) => form.lines.splice(i, 1)

// Enforce that a line is either a debit or a credit, never both.
function onDebit(l) {
  if (Number(l.debit) > 0) l.credit = 0
}
function onCredit(l) {
  if (Number(l.credit) > 0) l.debit = 0
}

async function save() {
  if (!balanced.value) {
    formError.value = 'Total debit must equal total credit (and be greater than 0)'
    return
  }
  saving.value = true
  formError.value = ''
  try {
    await api.post('/journal', {
      date: form.date,
      reference: form.reference,
      note: form.note,
      lines: form.lines
        .filter((l) => Number(l.debit) > 0 || Number(l.credit) > 0)
        .map((l) => ({ account_id: Number(l.account_id), debit: Number(l.debit) || 0, credit: Number(l.credit) || 0 })),
    })
    showModal.value = false
    page.value = 1
    load()
  } catch (e) {
    formError.value = e.response?.data?.message || 'Failed to post entry'
  } finally {
    saving.value = false
  }
}

onMounted(async () => {
  const res = await api.get('/accounts/active')
  accounts.value = res.data.data
  load()
})
</script>

<template>
  <div>
    <div class="mb-6 flex items-center justify-between">
      <h1 class="text-2xl font-bold">Journal Entries</h1>
      <button class="btn-primary" @click="openCreate">+ New Entry</button>
    </div>

    <div class="card overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-left text-sm">
          <thead class="border-b border-slate-200 bg-slate-50 text-xs uppercase text-slate-500 dark:border-slate-700 dark:bg-slate-700/40">
            <tr>
              <th class="px-4 py-3">Entry No.</th><th class="px-4 py-3">Date</th>
              <th class="px-4 py-3">Reference</th><th class="px-4 py-3">Lines</th>
              <th class="px-4 py-3 text-right">Amount</th><th class="px-4 py-3 text-right">View</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 dark:divide-slate-700">
            <tr v-if="loading"><td colspan="6" class="px-4 py-10 text-center text-slate-400">Loading…</td></tr>
            <tr v-else-if="!entries.length"><td colspan="6" class="px-4 py-10 text-center text-slate-400">No journal entries yet</td></tr>
            <tr v-for="e in entries" :key="e.id" class="hover:bg-slate-50 dark:hover:bg-slate-700/30">
              <td class="px-4 py-3 font-medium">{{ e.entry_no }}</td>
              <td class="px-4 py-3 text-slate-400">{{ new Date(e.date).toLocaleDateString() }}</td>
              <td class="px-4 py-3">{{ e.reference || '—' }}</td>
              <td class="px-4 py-3 text-slate-400">{{ (e.lines || []).length }}</td>
              <td class="px-4 py-3 text-right font-semibold">{{ money(entryTotal(e)) }}</td>
              <td class="px-4 py-3 text-right"><button class="btn-ghost !px-2 !py-1 text-xs" @click="viewing = e">View</button></td>
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

    <!-- New entry modal -->
    <Modal v-if="showModal" title="New Journal Entry" size="max-w-3xl" @close="showModal = false">
      <form class="space-y-4" @submit.prevent="save">
        <div v-if="formError" class="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600 dark:bg-red-500/10">{{ formError }}</div>

        <div class="grid gap-3 sm:grid-cols-3">
          <div><label class="label">Date</label><input v-model="form.date" type="date" class="input" /></div>
          <div class="sm:col-span-2"><label class="label">Reference</label><input v-model="form.reference" class="input" placeholder="e.g. Opening capital" /></div>
        </div>

        <div class="overflow-hidden rounded-xl border border-slate-200 dark:border-slate-700">
          <table class="w-full text-left text-sm">
            <thead class="bg-slate-50 text-xs uppercase text-slate-500 dark:bg-slate-700/40">
              <tr><th class="px-3 py-2">Account</th><th class="w-32 px-3 py-2">Debit</th><th class="w-32 px-3 py-2">Credit</th><th class="w-10 px-3 py-2"></th></tr>
            </thead>
            <tbody class="divide-y divide-slate-100 dark:divide-slate-700">
              <tr v-for="(l, i) in form.lines" :key="i">
                <td class="px-3 py-2">
                  <select v-model="l.account_id" class="input !py-1.5">
                    <option v-for="a in accounts" :key="a.id" :value="a.id">{{ a.code }} — {{ a.name }}</option>
                  </select>
                </td>
                <td class="px-3 py-2"><input v-model.number="l.debit" type="number" min="0" class="input !py-1.5 text-right" @input="onDebit(l)" /></td>
                <td class="px-3 py-2"><input v-model.number="l.credit" type="number" min="0" class="input !py-1.5 text-right" @input="onCredit(l)" /></td>
                <td class="px-3 py-2"><button type="button" class="btn-danger !px-2 !py-1 text-xs" :disabled="form.lines.length <= 2" @click="removeLine(i)">✕</button></td>
              </tr>
            </tbody>
            <tfoot class="bg-slate-50 font-semibold dark:bg-slate-700/40">
              <tr>
                <td class="px-3 py-2 text-right">Totals</td>
                <td class="px-3 py-2 text-right">{{ money(totalDebit) }}</td>
                <td class="px-3 py-2 text-right">{{ money(totalCredit) }}</td>
                <td></td>
              </tr>
            </tfoot>
          </table>
        </div>

        <div class="flex items-center justify-between">
          <button type="button" class="btn-ghost text-sm" @click="addLine">+ Add Line</button>
          <span class="text-sm font-medium" :class="balanced ? 'text-emerald-600' : 'text-amber-600'">
            {{ balanced ? '✓ Balanced' : 'Difference: ' + money(Math.abs(diff)) }}
          </span>
        </div>

        <div><label class="label">Note</label><input v-model="form.note" class="input" placeholder="Optional" /></div>

        <div class="flex justify-end gap-2">
          <button type="button" class="btn-ghost" @click="showModal = false">Cancel</button>
          <button type="submit" class="btn-primary" :disabled="saving || !balanced">{{ saving ? 'Posting…' : 'Post Entry' }}</button>
        </div>
      </form>
    </Modal>

    <!-- View entry modal -->
    <Modal v-if="viewing" :title="viewing.entry_no" @close="viewing = null">
      <div class="space-y-3">
        <div class="flex justify-between text-sm text-slate-500">
          <span>{{ new Date(viewing.date).toLocaleDateString() }}</span><span>{{ viewing.reference }}</span>
        </div>
        <table class="w-full text-left text-sm">
          <thead class="bg-slate-50 text-xs uppercase text-slate-500 dark:bg-slate-700/40">
            <tr><th class="px-3 py-2">Account</th><th class="px-3 py-2 text-right">Debit</th><th class="px-3 py-2 text-right">Credit</th></tr>
          </thead>
          <tbody class="divide-y divide-slate-100 dark:divide-slate-700">
            <tr v-for="l in viewing.lines" :key="l.id">
              <td class="px-3 py-2">{{ l.account?.code }} — {{ l.account?.name }}</td>
              <td class="px-3 py-2 text-right">{{ l.debit ? money(l.debit) : '' }}</td>
              <td class="px-3 py-2 text-right">{{ l.credit ? money(l.credit) : '' }}</td>
            </tr>
          </tbody>
        </table>
        <p v-if="viewing.note" class="text-xs text-slate-500">Note: {{ viewing.note }}</p>
      </div>
    </Modal>
  </div>
</template>
