<script setup>
import CrudPage from '../components/CrudPage.vue'

const typeColors = {
  asset: 'bg-blue-100 text-blue-700 dark:bg-blue-500/20 dark:text-blue-300',
  liability: 'bg-amber-100 text-amber-700 dark:bg-amber-500/20 dark:text-amber-300',
  equity: 'bg-violet-100 text-violet-700 dark:bg-violet-500/20 dark:text-violet-300',
  income: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-500/20 dark:text-emerald-300',
  expense: 'bg-red-100 text-red-700 dark:bg-red-500/20 dark:text-red-300',
}
const typeBadge = (row) => `<span class="badge ${typeColors[row.type] || ''}">${row.type}</span>`
const activeBadge = (row) =>
  row.is_active
    ? '<span class="badge bg-emerald-100 text-emerald-700 dark:bg-emerald-500/20 dark:text-emerald-300">Active</span>'
    : '<span class="badge bg-slate-200 text-slate-600 dark:bg-slate-600 dark:text-slate-300">Inactive</span>'

const typeOptions = [
  { value: 'asset', label: 'Asset' },
  { value: 'liability', label: 'Liability' },
  { value: 'equity', label: 'Equity' },
  { value: 'income', label: 'Income' },
  { value: 'expense', label: 'Expense' },
]

const columns = [
  { key: 'code', label: 'Code' },
  { key: 'name', label: 'Name' },
  { key: 'type', label: 'Type', render: typeBadge },
  { key: 'is_active', label: 'Status', render: activeBadge },
]

const fields = [
  { key: 'code', label: 'Account Code (e.g. 1000)', type: 'text', required: true },
  { key: 'name', label: 'Account Name', type: 'text', required: true },
  { key: 'type', label: 'Type', type: 'select', options: typeOptions },
  { key: 'is_active', label: 'Status', type: 'checkbox' },
]

const newItem = () => ({ code: '', name: '', type: 'asset', is_active: true })
</script>

<template>
  <CrudPage title="Chart of Accounts" endpoint="/accounts" :columns="columns" :fields="fields" :new-item="newItem" />
</template>
