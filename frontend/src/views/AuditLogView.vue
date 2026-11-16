<script setup>
import { ref, onMounted } from 'vue'
import api from '../lib/api'

const logs = ref([])
const meta = ref({ page: 1, total: 0, total_pages: 1 })
const page = ref(1)
const loading = ref(false)

const methodClasses = {
  POST: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-500/20 dark:text-emerald-300',
  PUT: 'bg-blue-100 text-blue-700 dark:bg-blue-500/20 dark:text-blue-300',
  PATCH: 'bg-blue-100 text-blue-700 dark:bg-blue-500/20 dark:text-blue-300',
  DELETE: 'bg-red-100 text-red-700 dark:bg-red-500/20 dark:text-red-300',
}
const methodClass = (m) => methodClasses[m] || 'bg-slate-100 text-slate-600 dark:bg-slate-700 dark:text-slate-300'

// 2xx = success, anything else = the request was rejected.
const statusClass = (s) =>
  s >= 200 && s < 300
    ? 'text-emerald-600 dark:text-emerald-400'
    : 'text-red-600 dark:text-red-400'

// Turn "/api/v1/products/:id" into a readable "products / :id".
const readablePath = (p) => p.replace(/^\/api\/v1\//, '').replace(/\//g, ' / ')

const fmtTime = (d) =>
  new Date(d).toLocaleString('en-GB', {
    day: '2-digit', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit',
  })

async function load() {
  loading.value = true
  try {
    const { data } = await api.get('/audit-logs', { params: { page: page.value, per_page: 20 } })
    logs.value = data.data
    meta.value = data.meta
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<template>
  <div>
    <div class="mb-6">
      <h1 class="text-2xl font-bold tracking-tight">Audit Log</h1>
      <p class="mt-0.5 text-sm text-slate-500 dark:text-slate-400">
        Every data-changing action, with the user who made it — for traceability and accountability.
      </p>
    </div>

    <div class="card overflow-hidden">
      <div class="overflow-x-auto">
        <table class="w-full text-left text-sm">
          <thead class="border-b border-slate-200 bg-slate-50 text-xs uppercase text-slate-500 dark:border-slate-700 dark:bg-slate-700/40">
            <tr>
              <th class="px-4 py-3">When</th>
              <th class="px-4 py-3">User</th>
              <th class="px-4 py-3">Action</th>
              <th class="px-4 py-3">Resource</th>
              <th class="px-4 py-3 text-right">Status</th>
              <th class="px-4 py-3 text-right">IP</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-slate-100 dark:divide-slate-700">
            <tr v-if="loading"><td colspan="6" class="px-4 py-10 text-center text-slate-400">Loading…</td></tr>
            <tr v-else-if="!logs.length"><td colspan="6" class="px-4 py-10 text-center text-slate-400">No activity recorded yet</td></tr>
            <tr v-for="l in logs" :key="l.id" class="hover:bg-slate-50 dark:hover:bg-slate-700/30">
              <td class="whitespace-nowrap px-4 py-3 text-slate-500 dark:text-slate-400">{{ fmtTime(l.created_at) }}</td>
              <td class="px-4 py-3 font-medium">{{ l.user?.name || '—' }}</td>
              <td class="px-4 py-3">
                <span class="badge" :class="methodClass(l.method)">{{ l.method }}</span>
              </td>
              <td class="px-4 py-3 font-mono text-xs text-slate-500 dark:text-slate-400">{{ readablePath(l.path) }}</td>
              <td class="px-4 py-3 text-right font-semibold" :class="statusClass(l.status)">{{ l.status }}</td>
              <td class="px-4 py-3 text-right text-slate-400">{{ l.ip }}</td>
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
  </div>
</template>
