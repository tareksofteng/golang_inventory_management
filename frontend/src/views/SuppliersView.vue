<script setup>
import CrudPage from '../components/CrudPage.vue'

const activeBadge = (row) =>
  row.is_active
    ? '<span class="badge bg-emerald-100 text-emerald-700 dark:bg-emerald-500/20 dark:text-emerald-300">Active</span>'
    : '<span class="badge bg-slate-200 text-slate-600 dark:bg-slate-600 dark:text-slate-300">Inactive</span>'

const columns = [
  { key: 'name', label: 'Name' },
  { key: 'email', label: 'Email' },
  { key: 'phone', label: 'Phone' },
  { key: 'is_active', label: 'Status', render: activeBadge },
]

const fields = [
  { key: 'name', label: 'Name', type: 'text', required: true },
  { key: 'email', label: 'Email', type: 'email', required: true },
  { key: 'phone', label: 'Phone', type: 'text' },
  { key: 'address', label: 'Address', type: 'text' },
  { key: 'is_active', label: 'Status', type: 'checkbox' },
]

const newItem = () => ({ name: '', email: '', phone: '', address: '', is_active: true })
</script>

<template>
  <CrudPage title="Suppliers" endpoint="/suppliers" :columns="columns" :fields="fields" :new-item="newItem" />
</template>
