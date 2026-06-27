<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { RouterView, RouterLink, useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()

const sidebarOpen = ref(false)
const isDark = ref(document.documentElement.classList.contains('dark'))

// Grouped navigation. A group with title === null renders ungrouped (top).
const groups = [
  {
    title: null,
    items: [{ name: 'Dashboard', to: '/dashboard', icon: '📊' }],
  },
  {
    title: 'Catalog',
    items: [
      { name: 'Products', to: '/products', icon: '📦', perm: 'product.manage' },
      { name: 'Categories', to: '/categories', icon: '🏷️', perm: 'product.manage' },
      { name: 'Suppliers', to: '/suppliers', icon: '🚚', perm: 'product.manage' },
    ],
  },
  {
    title: 'Purchases',
    items: [
      { name: 'Purchases', to: '/purchases', icon: '🧾', perm: 'purchase.manage' },
      { name: 'Purchase Returns', to: '/returns?tab=purchase', icon: '↩️', perm: 'purchase.manage' },
    ],
  },
  {
    title: 'Sales',
    items: [
      { name: 'Customers', to: '/customers', icon: '👤', perm: 'sales.manage' },
      { name: 'Sales', to: '/sales', icon: '🛒', perm: 'sales.manage' },
      { name: 'Sales Returns', to: '/returns?tab=sale', icon: '↪️', perm: 'sales.manage' },
    ],
  },
  {
    title: 'Finance',
    items: [
      { name: 'Payments', to: '/payments', icon: '💳', anyPerm: ['sales.manage', 'purchase.manage'] },
      { name: 'Ledger', to: '/ledger', icon: '📒', perm: 'report.access' },
      { name: 'Reports', to: '/reports', icon: '📈', perm: 'report.access' },
    ],
  },
  {
    title: 'Administration',
    items: [{ name: 'Users', to: '/users', icon: '🔐', perm: 'user.manage' }],
  },
]

function canSee(item) {
  if (item.anyPerm) return item.anyPerm.some((p) => auth.can(p))
  return !item.perm || auth.can(item.perm)
}

// Only groups that have at least one visible item.
const visibleGroups = computed(() =>
  groups
    .map((g) => ({ ...g, items: g.items.filter(canSee) }))
    .filter((g) => g.items.length),
)

// Collapsible groups (all expanded by default).
const collapsed = reactive({})
const toggle = (title) => {
  if (title) collapsed[title] = !collapsed[title]
}

// Active highlight that understands the ?tab= links.
function isActive(to) {
  const [path, qs] = to.split('?')
  if (route.path !== path) return false
  if (!qs) return true
  const tab = new URLSearchParams(qs).get('tab')
  return route.query.tab === tab
}

function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

async function logout() {
  await auth.logout()
  router.push('/login')
}

const initial = computed(() => (auth.user?.name || '?').charAt(0).toUpperCase())

onMounted(() => {
  if (auth.isAuthenticated) auth.fetchMe().catch(() => {})
})
</script>

<template>
  <div class="min-h-screen bg-slate-100 dark:bg-slate-900 lg:flex">
    <!-- Sidebar -->
    <aside
      :class="sidebarOpen ? 'translate-x-0' : '-translate-x-full'"
      class="fixed inset-y-0 left-0 z-30 flex w-64 transform flex-col border-r border-slate-200 bg-white transition-transform duration-300 dark:border-slate-800 dark:bg-slate-800 lg:static lg:translate-x-0"
    >
      <!-- Brand -->
      <div class="flex h-16 items-center gap-3 border-b border-slate-200 px-5 dark:border-slate-700/60">
        <span class="grid h-9 w-9 place-items-center rounded-xl bg-gradient-to-br from-brand-500 to-indigo-700 text-white shadow-md shadow-brand-600/30">📦</span>
        <div class="leading-tight">
          <div class="font-bold">Inventory</div>
          <div class="text-[11px] text-slate-400">Management System</div>
        </div>
      </div>

      <!-- Nav -->
      <nav class="flex-1 space-y-4 overflow-y-auto px-3 py-4">
        <div v-for="group in visibleGroups" :key="group.title || 'main'">
          <!-- Group header -->
          <button
            v-if="group.title"
            class="mb-1 flex w-full items-center justify-between px-3 py-1 text-[11px] font-semibold uppercase tracking-wider text-slate-400 transition hover:text-slate-600 dark:hover:text-slate-200"
            @click="toggle(group.title)"
          >
            {{ group.title }}
            <svg
              class="h-3 w-3 transition-transform duration-200"
              :class="collapsed[group.title] ? '-rotate-90' : ''"
              fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5"
            >
              <path stroke-linecap="round" stroke-linejoin="round" d="M19 9l-7 7-7-7" />
            </svg>
          </button>

          <!-- Collapsible items -->
          <div class="grid transition-[grid-template-rows] duration-200 ease-out" :class="group.title && collapsed[group.title] ? 'grid-rows-[0fr]' : 'grid-rows-[1fr]'">
            <div class="overflow-hidden">
              <div class="space-y-0.5">
                <RouterLink
                  v-for="item in group.items"
                  :key="item.to"
                  :to="item.to"
                  class="group relative flex items-center gap-3 rounded-xl px-3 py-2.5 text-sm font-medium transition-all duration-150"
                  :class="isActive(item.to)
                    ? 'bg-gradient-to-r from-brand-600 to-indigo-600 text-white shadow-md shadow-brand-600/30'
                    : 'text-slate-600 hover:bg-slate-100 dark:text-slate-300 dark:hover:bg-slate-700/60'"
                  @click="sidebarOpen = false"
                >
                  <span
                    class="grid h-7 w-7 shrink-0 place-items-center rounded-lg text-base transition"
                    :class="isActive(item.to) ? 'bg-white/20' : 'bg-slate-100 group-hover:bg-white dark:bg-slate-700/70 dark:group-hover:bg-slate-600'"
                  >{{ item.icon }}</span>
                  {{ item.name }}
                </RouterLink>
              </div>
            </div>
          </div>
        </div>
      </nav>

      <!-- User footer -->
      <div class="border-t border-slate-200 p-3 dark:border-slate-700/60">
        <div class="flex items-center gap-3 rounded-xl bg-slate-50 p-2.5 dark:bg-slate-700/40">
          <span class="grid h-9 w-9 place-items-center rounded-full bg-gradient-to-br from-brand-500 to-indigo-700 text-sm font-bold text-white">{{ initial }}</span>
          <div class="min-w-0 flex-1 leading-tight">
            <div class="truncate text-sm font-semibold">{{ auth.user?.name }}</div>
            <div class="truncate text-xs capitalize text-slate-400">{{ auth.user?.role?.replace('_', ' ') }}</div>
          </div>
          <button class="rounded-lg p-1.5 text-slate-400 transition hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-500/10" title="Logout" @click="logout">
            <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2">
              <path stroke-linecap="round" stroke-linejoin="round" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a3 3 0 01-3 3H6a3 3 0 01-3-3V7a3 3 0 013-3h4a3 3 0 013 3v1" />
            </svg>
          </button>
        </div>
      </div>
    </aside>

    <!-- Backdrop (mobile) -->
    <div v-if="sidebarOpen" class="fixed inset-0 z-20 bg-black/40 backdrop-blur-sm lg:hidden" @click="sidebarOpen = false" />

    <!-- Main -->
    <div class="flex min-h-screen flex-1 flex-col">
      <header class="sticky top-0 z-10 flex h-16 items-center justify-between border-b border-slate-200 bg-white/70 px-4 backdrop-blur-md dark:border-slate-800 dark:bg-slate-800/70 lg:px-8">
        <div class="flex items-center gap-3">
          <button class="rounded-lg p-2 text-slate-500 hover:bg-slate-100 dark:hover:bg-slate-700 lg:hidden" @click="sidebarOpen = true">
            <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2"><path stroke-linecap="round" d="M4 6h16M4 12h16M4 18h16" /></svg>
          </button>
          <div class="font-semibold text-slate-700 dark:text-slate-200">Inventory Management System</div>
        </div>
        <div class="flex items-center gap-2">
          <button class="rounded-lg p-2 text-lg transition hover:bg-slate-100 dark:hover:bg-slate-700" :title="isDark ? 'Light mode' : 'Dark mode'" @click="toggleTheme">
            {{ isDark ? '☀️' : '🌙' }}
          </button>
        </div>
      </header>

      <main class="flex-1 p-4 lg:p-8">
        <RouterView />
      </main>
    </div>
  </div>
</template>
