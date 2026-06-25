<script setup>
import CrudPage from '../components/CrudPage.vue'

const money = (n) => '৳' + Number(n || 0).toLocaleString('en-IN')

const dueBadge = (row) =>
  Number(row.due) > 0
    ? `<span class="badge bg-amber-100 text-amber-700 dark:bg-amber-500/20 dark:text-amber-300">${money(row.due)}</span>`
    : '<span class="badge bg-emerald-100 text-emerald-700 dark:bg-emerald-500/20 dark:text-emerald-300">Clear</span>'

const columns = [
  { key: 'name', label: 'Name' },
  { key: 'phone', label: 'Phone' },
  { key: 'address', label: 'Address' },
  { key: 'due', label: 'Due', render: dueBadge },
]

const fields = [
  { key: 'name', label: 'Name', type: 'text', required: true },
  { key: 'email', label: 'Email', type: 'email' },
  { key: 'phone', label: 'Phone', type: 'text' },
  { key: 'address', label: 'Address', type: 'text' },
  { key: 'due', label: 'Opening Due', type: 'number' },
  { key: 'is_active', label: 'Status', type: 'checkbox' },
]

const newItem = () => ({ name: '', email: '', phone: '', address: '', due: 0, is_active: true })
</script>

<template>
  <CrudPage title="Customers" endpoint="/customers" :columns="columns" :fields="fields" :new-item="newItem" />
</template>
