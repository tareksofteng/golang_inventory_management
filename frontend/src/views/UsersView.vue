<script setup>
import CrudPage from '../components/CrudPage.vue'

const roleBadge = (row) => {
  const colors = {
    super_admin: 'bg-brand-100 text-brand-700 dark:bg-brand-600/20 dark:text-brand-200',
    admin: 'bg-blue-100 text-blue-700 dark:bg-blue-500/20 dark:text-blue-300',
    manager: 'bg-amber-100 text-amber-700 dark:bg-amber-500/20 dark:text-amber-300',
    salesman: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-500/20 dark:text-emerald-300',
  }
  return `<span class="badge ${colors[row.role] || ''}">${row.role.replace('_', ' ')}</span>`
}
const activeBadge = (row) =>
  row.is_active
    ? '<span class="badge bg-emerald-100 text-emerald-700 dark:bg-emerald-500/20 dark:text-emerald-300">Active</span>'
    : '<span class="badge bg-red-100 text-red-700 dark:bg-red-500/20 dark:text-red-300">Disabled</span>'

const permsCell = (row) =>
  `<span class="badge bg-slate-100 text-slate-600 dark:bg-slate-700 dark:text-slate-300">${(row.permissions || []).length} permission(s)</span>`

const roleOptions = [
  { value: 'super_admin', label: 'Super Admin' },
  { value: 'admin', label: 'Admin' },
  { value: 'manager', label: 'Manager' },
  { value: 'salesman', label: 'Salesman' },
]

// All assignable permissions. Leaving every box unchecked falls back to the
// role's defaults on the backend.
const permissionOptions = [
  { value: 'product.manage', label: 'Product Management' },
  { value: 'purchase.manage', label: 'Purchase Management' },
  { value: 'sales.manage', label: 'Sales Management' },
  { value: 'report.access', label: 'Report Access' },
  { value: 'user.manage', label: 'User Management' },
]

const columns = [
  { key: 'name', label: 'Name' },
  { key: 'email', label: 'Email' },
  { key: 'role', label: 'Role', render: roleBadge },
  { key: 'permissions', label: 'Permissions', render: permsCell },
  { key: 'is_active', label: 'Status', render: activeBadge },
]

// Password field only matters on create; on edit the API ignores it (separate
// change-password endpoint), so we keep the create form simple here.
const fields = [
  { key: 'name', label: 'Name', type: 'text', required: true },
  { key: 'email', label: 'Email', type: 'email', required: true },
  { key: 'password', label: 'Password (min 6)', type: 'password' },
  { key: 'role', label: 'Role', type: 'select', options: roleOptions },
  { key: 'permissions', label: 'Permissions (leave empty to use role defaults)', type: 'checkboxes', options: permissionOptions },
  { key: 'is_active', label: 'Status', type: 'checkbox' },
]

const newItem = () => ({ name: '', email: '', password: '', role: 'salesman', permissions: [], is_active: true })
</script>

<template>
  <CrudPage title="Users" endpoint="/users" :columns="columns" :fields="fields" :new-item="newItem" />
</template>
