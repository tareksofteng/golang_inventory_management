<script setup>
import CrudPage from '../components/CrudPage.vue'

const activeBadge = (row) =>
  row.is_active
    ? '<span class="badge bg-emerald-100 text-emerald-700 dark:bg-emerald-500/20 dark:text-emerald-300">Active</span>'
    : '<span class="badge bg-slate-200 text-slate-600 dark:bg-slate-600 dark:text-slate-300">Inactive</span>'

const columns = [
  { key: 'name', label: 'Name' },
  { key: 'description', label: 'Description' },
  { key: 'is_active', label: 'Status', render: activeBadge },
]

const fields = [
  { key: 'name', label: 'Name', type: 'text', required: true },
  { key: 'description', label: 'Description', type: 'text' },
  { key: 'is_active', label: 'Status', type: 'checkbox' },
]

const newItem = () => ({ name: '', description: '', is_active: true })
</script>

<template>
  <CrudPage title="Categories" endpoint="/categories" :columns="columns" :fields="fields" :new-item="newItem" />
</template>
