<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import api from '../lib/api'

const money = (n) => '৳' + Number(n || 0).toLocaleString('en-IN', { minimumFractionDigits: 2 })

const accounts = ref([])
const saving = ref(false)
const message = ref('')
const error = ref('')

const today = () => new Date().toISOString().slice(0, 10)
const form = reactive({ expense_account_id: '', paid_from_id: '', amount: 0, date: today(), note: '' })

// Expense accounts (exclude COGS 5000 — that's auto-posted); pay from asset accounts.
const expenseAccounts = computed(() => accounts.value.filter((a) => a.type === 'expense' && a.code !== '5000'))
const payAccounts = computed(() => accounts.value.filter((a) => a.type === 'asset'))
// Operating-expense accounts only (COGS is auto-posted from sales, not a manual expense).
const expenseIds = computed(() => new Set(expenseAccounts.value.map((a) => a.id)))

const recent = ref([])

async function loadRecent() {
  const { data } = await api.get('/journal', { params: { per_page: 30 } })
  // Keep only entries that debit an expense account.
  recent.value = data.data
    .filter((e) => (e.lines || []).some((l) => l.debit > 0 && expenseIds.value.has(l.account_id)))
    .slice(0, 8)
}

async function save() {
  error.value = ''
  message.value = ''
  if (!form.expense_account_id || !form.paid_from_id || Number(form.amount) <= 0) {
    error.value = 'Choose an expense account, a paid-from account and an amount'
    return
  }
  saving.value = true
  const exp = accounts.value.find((a) => a.id === Number(form.expense_account_id))
  try {
    await api.post('/journal', {
      date: form.date,
      reference: 'Expense: ' + (exp?.name || ''),
      note: form.note,
      lines: [
        { account_id: Number(form.expense_account_id), debit: Number(form.amount), credit: 0 },
        { account_id: Number(form.paid_from_id), debit: 0, credit: Number(form.amount) },
      ],
    })
    message.value = 'Expense recorded successfully'
    form.amount = 0
    form.note = ''
    loadRecent()
  } catch (e) {
    error.value = e.response?.data?.message || 'Failed to record expense'
  } finally {
    saving.value = false
  }
}

onMounted(async () => {
  const res = await api.get('/accounts/active')
  accounts.value = res.data.data
  form.expense_account_id = expenseAccounts.value[0]?.id || ''
  form.paid_from_id = payAccounts.value[0]?.id || ''
  loadRecent()
})
</script>

<template>
  <div>
    <h1 class="mb-6 text-2xl font-bold">Record Expense</h1>

    <div class="grid gap-4 lg:grid-cols-2">
      <!-- Form -->
      <div class="card p-5">
        <div v-if="message" class="mb-3 rounded-lg bg-emerald-50 px-3 py-2 text-sm text-emerald-700 dark:bg-emerald-500/10">{{ message }}</div>
        <div v-if="error" class="mb-3 rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600 dark:bg-red-500/10">{{ error }}</div>

        <form class="space-y-4" @submit.prevent="save">
          <div>
            <label class="label">Expense Category</label>
            <select v-model="form.expense_account_id" class="input">
              <option v-for="a in expenseAccounts" :key="a.id" :value="a.id">{{ a.code }} — {{ a.name }}</option>
            </select>
          </div>
          <div class="grid grid-cols-2 gap-3">
            <div>
              <label class="label">Amount</label>
              <input v-model.number="form.amount" type="number" min="0" class="input" />
            </div>
            <div>
              <label class="label">Date</label>
              <input v-model="form.date" type="date" class="input" />
            </div>
          </div>
          <div>
            <label class="label">Paid From</label>
            <select v-model="form.paid_from_id" class="input">
              <option v-for="a in payAccounts" :key="a.id" :value="a.id">{{ a.code }} — {{ a.name }}</option>
            </select>
          </div>
          <div>
            <label class="label">Note</label>
            <input v-model="form.note" class="input" placeholder="Optional" />
          </div>
          <button type="submit" class="btn-primary w-full" :disabled="saving">{{ saving ? 'Saving…' : 'Record Expense' }}</button>
        </form>
      </div>

      <!-- Recent -->
      <div class="card overflow-hidden">
        <div class="border-b border-slate-200 px-5 py-3 font-semibold dark:border-slate-700">Recent Expenses</div>
        <table class="w-full text-left text-sm">
          <tbody class="divide-y divide-slate-100 dark:divide-slate-700">
            <tr v-if="!recent.length"><td class="px-5 py-6 text-center text-slate-400" colspan="3">No expenses yet</td></tr>
            <tr v-for="e in recent" :key="e.id">
              <td class="px-5 py-2.5">
                <div class="font-medium">{{ e.reference }}</div>
                <div class="text-xs text-slate-400">{{ e.entry_no }} · {{ new Date(e.date).toLocaleDateString() }}</div>
              </td>
              <td class="px-5 py-2.5 text-right font-semibold">
                {{ money((e.lines || []).filter((l) => l.debit > 0 && expenseIds.has(l.account_id)).reduce((s, l) => s + l.debit, 0)) }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>
