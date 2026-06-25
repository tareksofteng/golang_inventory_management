<script setup>
import { ref, reactive, watch, onMounted } from 'vue'
import api from '../lib/api'
import Modal from './Modal.vue'

const props = defineProps({
  title: { type: String, required: true },
  endpoint: { type: String, required: true }, // e.g. '/products'
  columns: { type: Array, required: true }, // [{ key, label, render? }]
  fields: { type: Array, required: true }, // form fields [{ key,label,type,options?,required? }]
  newItem: { type: Function, default: () => ({}) }, // factory for a blank form
})

const items = ref([])
const meta = ref({ page: 1, per_page: 10, total: 0, total_pages: 1 })
const loading = ref(false)
const search = ref('')
const page = ref(1)

const showModal = ref(false)
const editingId = ref(null)
const form = reactive({})
const formError = ref('')
const fieldErrors = ref({})
const saving = ref(false)

let searchTimer
watch(search, () => {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    page.value = 1
    load()
  }, 350)
})
watch(page, load)

async function load() {
  loading.value = true
  try {
    const { data } = await api.get(props.endpoint, {
      params: { page: page.value, per_page: 10, search: search.value },
    })
    items.value = data.data
    meta.value = data.meta
  } finally {
    loading.value = false
  }
}

function openCreate() {
  editingId.value = null
  Object.assign(form, props.newItem())
  formError.value = ''
  fieldErrors.value = {}
  showModal.value = true
}

function openEdit(row) {
  editingId.value = row.id
  Object.assign(form, props.newItem(), row)
  formError.value = ''
  fieldErrors.value = {}
  showModal.value = true
}

async function save() {
  saving.value = true
  formError.value = ''
  fieldErrors.value = {}

  // Coerce types so numbers/booleans are sent correctly.
  const payload = {}
  for (const f of props.fields) {
    let v = form[f.key]
    if (f.type === 'number') v = v === '' || v == null ? 0 : Number(v)
    payload[f.key] = v
  }

  try {
    if (editingId.value) await api.put(`${props.endpoint}/${editingId.value}`, payload)
    else await api.post(props.endpoint, payload)
    showModal.value = false
    load()
  } catch (e) {
    formError.value = e.response?.data?.message || 'Something went wrong'
    fieldErrors.value = e.response?.data?.errors || {}
  } finally {
    saving.value = false
  }
}

async function remove(row) {
  if (!confirm(`Delete "${row.name || row.email}"?`)) return
  await api.delete(`${props.endpoint}/${row.id}`)
  // If we deleted the last row on a page, step back a page.
  if (items.value.length === 1 && page.value > 1) page.value--
  else load()
}

onMounted(load)
defineExpose({ load })
</script>

<template>
  <div>
    <div class="mb-6 flex flex-wrap items-center justify-between gap-3">
      <h1 class="text-2xl font-bold">{{ title }}</h1>
      <div class="flex items-center gap-2">
        <input v-model="search" class="input w-56" placeholder="Search..." />
        <button class="btn-primary whitespace-nowrap" @click="openCreate">+ Add New</button>
      </div>
    </div>

    <div class="card overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-left text-sm">
          <thead class="border-b border-slate-200 bg-slate-50 text-xs uppercase text-slate-500 dark:border-slate-700 dark:bg-slate-700/40">
            <tr>
              <th v-for="col in columns" :key="col.key" class="px-4 py-3 font-medium">{{ col.label }}</th>
              <th class="px-4 py-3 text-right font-medium">Actions</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 dark:divide-slate-700">
            <tr v-if="loading">
              <td :colspan="columns.length + 1" class="px-4 py-10 text-center text-slate-400">Loading…</td>
            </tr>
            <tr v-else-if="!items.length">
              <td :colspan="columns.length + 1" class="px-4 py-10 text-center text-slate-400">No records found</td>
            </tr>
            <tr v-for="row in items" :key="row.id" class="hover:bg-slate-50 dark:hover:bg-slate-700/30">
              <td v-for="col in columns" :key="col.key" class="px-4 py-3">
                <span v-html="col.render ? col.render(row) : row[col.key]" />
              </td>
              <td class="px-4 py-3 text-right">
                <button class="btn-ghost !px-2 !py-1 text-xs" @click="openEdit(row)">Edit</button>
                <button class="btn-danger !px-2 !py-1 text-xs" @click="remove(row)">Delete</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="flex items-center justify-between border-t border-slate-200 px-4 py-3 text-sm dark:border-slate-700">
        <span class="text-slate-500">
          Page {{ meta.page }} of {{ meta.total_pages || 1 }} · {{ meta.total }} total
        </span>
        <div class="flex gap-2">
          <button class="btn-ghost !py-1" :disabled="page <= 1" @click="page--">Prev</button>
          <button class="btn-ghost !py-1" :disabled="page >= meta.total_pages" @click="page++">Next</button>
        </div>
      </div>
    </div>

    <!-- Create / Edit modal -->
    <Modal v-if="showModal" :title="editingId ? `Edit ${title}` : `New ${title}`" @close="showModal = false">
      <form class="space-y-4" @submit.prevent="save">
        <div v-if="formError" class="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600 dark:bg-red-500/10">
          {{ formError }}
        </div>

        <div v-for="f in fields" :key="f.key">
          <label class="label">{{ f.label }}</label>

          <select v-if="f.type === 'select'" v-model="form[f.key]" class="input">
            <option v-for="opt in f.options" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
          </select>

          <label v-else-if="f.type === 'checkbox'" class="flex items-center gap-2">
            <input v-model="form[f.key]" type="checkbox" class="h-4 w-4 rounded" />
            <span class="text-sm text-slate-500">{{ f.hint || 'Active' }}</span>
          </label>

          <input
            v-else
            v-model="form[f.key]"
            :type="f.type || 'text'"
            class="input"
            :placeholder="f.label"
          />

          <p v-if="fieldErrors[f.key]" class="mt-1 text-xs text-red-500">{{ fieldErrors[f.key] }}</p>
        </div>

        <div class="flex justify-end gap-2 pt-2">
          <button type="button" class="btn-ghost" @click="showModal = false">Cancel</button>
          <button type="submit" class="btn-primary" :disabled="saving">{{ saving ? 'Saving…' : 'Save' }}</button>
        </div>
      </form>
    </Modal>
  </div>
</template>
