<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { RouterView, RouterLink, useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()

const sidebarOpen = ref(false)
const isDark = ref(document.documentElement.classList.contains('dark'))

// Premium line icons (Heroicons outline). Each entry is an array of path data.
const ICONS = {
  dashboard: ['M3.75 6A2.25 2.25 0 0 1 6 3.75h2.25A2.25 2.25 0 0 1 10.5 6v2.25a2.25 2.25 0 0 1-2.25 2.25H6a2.25 2.25 0 0 1-2.25-2.25V6ZM3.75 15.75A2.25 2.25 0 0 1 6 13.5h2.25a2.25 2.25 0 0 1 2.25 2.25V18a2.25 2.25 0 0 1-2.25 2.25H6A2.25 2.25 0 0 1 3.75 18v-2.25ZM13.5 6a2.25 2.25 0 0 1 2.25-2.25H18A2.25 2.25 0 0 1 20.25 6v2.25A2.25 2.25 0 0 1 18 10.5h-2.25a2.25 2.25 0 0 1-2.25-2.25V6ZM13.5 15.75a2.25 2.25 0 0 1 2.25-2.25H18a2.25 2.25 0 0 1 2.25 2.25V18A2.25 2.25 0 0 1 18 20.25h-2.25a2.25 2.25 0 0 1-2.25-2.25V15.75Z'],
  cube: ['m21 7.5-9-5.25L3 7.5m18 0-9 5.25m9-5.25v9l-9 5.25M3 7.5l9 5.25M3 7.5v9l9 5.25m0-9v9'],
  tag: ['M9.568 3H5.25A2.25 2.25 0 0 0 3 5.25v4.318c0 .597.237 1.17.659 1.591l9.581 9.581c.699.699 1.78.872 2.607.33a18.095 18.095 0 0 0 5.223-5.223c.542-.827.369-1.908-.33-2.607L11.16 3.66A2.25 2.25 0 0 0 9.568 3Z', 'M6 6h.008v.008H6V6Z'],
  truck: ['M8.25 18.75a1.5 1.5 0 0 1-3 0m3 0a1.5 1.5 0 0 0-3 0m3 0h6m-9 0H3.375a1.125 1.125 0 0 1-1.125-1.125V14.25m17.25 4.5a1.5 1.5 0 0 1-3 0m3 0a1.5 1.5 0 0 0-3 0m3 0h1.125c.621 0 1.129-.504 1.09-1.124a17.902 17.902 0 0 0-3.213-9.193 2.056 2.056 0 0 0-1.58-.86H14.25M16.5 18.75h-6.75M14.25 6v9.75M14.25 6h-.75a3 3 0 0 0-3 3v.75m3-3.75h.375a1.125 1.125 0 0 1 1.125 1.125V9.75m0 9V9.75m0 0H14.25'],
  bag: ['M15.75 10.5V6a3.75 3.75 0 1 0-7.5 0v4.5m11.356-1.993 1.263 12c.07.665-.45 1.243-1.119 1.243H4.25a1.125 1.125 0 0 1-1.12-1.243l1.264-12A1.125 1.125 0 0 1 5.513 7.5h12.974c.576 0 1.059.435 1.119 1.007Z'],
  uturnLeft: ['M9 15 3 9m0 0 6-6M3 9h12a6 6 0 0 1 0 12h-3'],
  uturnRight: ['m15 15 6-6m0 0-6-6m6 6H9a6 6 0 0 0 0 12h3'],
  user: ['M15.75 6a3.75 3.75 0 1 1-7.5 0 3.75 3.75 0 0 1 7.5 0ZM4.501 20.118a7.5 7.5 0 0 1 14.998 0A17.933 17.933 0 0 1 12 21.75c-2.676 0-5.216-.584-7.499-1.632Z'],
  cart: ['M2.25 3h1.386c.51 0 .955.343 1.087.835l.383 1.437M7.5 14.25a3 3 0 0 0-3 3h15.75m-12.75-3h11.218c1.121-2.3 2.1-4.684 2.924-7.138a60.114 60.114 0 0 0-16.536-1.84M7.5 14.25 5.106 5.272M6 20.25a.75.75 0 1 1-1.5 0 .75.75 0 0 1 1.5 0Zm12.75 0a.75.75 0 1 1-1.5 0 .75.75 0 0 1 1.5 0Z'],
  card: ['M2.25 8.25h19.5M2.25 9h19.5m-16.5 5.25h6m-6 2.25h3m-3.75 3h15a2.25 2.25 0 0 0 2.25-2.25V6.75A2.25 2.25 0 0 0 19.5 4.5h-15a2.25 2.25 0 0 0-2.25 2.25v10.5A2.25 2.25 0 0 0 4.5 19.5Z'],
  book: ['M12 6.042A8.967 8.967 0 0 0 6 3.75c-1.052 0-2.062.18-3 .512v14.25A8.987 8.987 0 0 1 6 18c2.305 0 4.408.867 6 2.292m0-14.25a8.966 8.966 0 0 1 6-2.292c1.052 0 2.062.18 3 .512v14.25A8.987 8.987 0 0 0 18 18a8.967 8.967 0 0 0-6 2.292m0-14.25v14.25'],
  chart: ['M3 13.125C3 12.504 3.504 12 4.125 12h2.25c.621 0 1.125.504 1.125 1.125v6.75C7.5 20.496 6.996 21 6.375 21h-2.25A1.125 1.125 0 0 1 3 19.875v-6.75ZM9.75 8.625c0-.621.504-1.125 1.125-1.125h2.25c.621 0 1.125.504 1.125 1.125v11.25c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 0 1-1.125-1.125V8.625ZM16.5 4.125c0-.621.504-1.125 1.125-1.125h2.25C20.496 3 21 3.504 21 4.125v15.75c0 .621-.504 1.125-1.125 1.125h-2.25a1.125 1.125 0 0 1-1.125-1.125V4.125Z'],
  users: ['M15 19.128a9.38 9.38 0 0 0 2.625.372 9.337 9.337 0 0 0 4.121-.952 4.125 4.125 0 0 0-7.533-2.493M15 19.128v-.003c0-1.113-.285-2.16-.786-3.07M15 19.128v.106A12.318 12.318 0 0 1 8.624 21c-2.331 0-4.512-.645-6.374-1.766l-.001-.109a6.375 6.375 0 0 1 11.964-3.07M12 6.375a3.375 3.375 0 1 1-6.75 0 3.375 3.375 0 0 1 6.75 0Zm8.25 2.25a2.625 2.625 0 1 1-5.25 0 2.625 2.625 0 0 1 5.25 0Z'],
}

const groups = [
  { title: null, items: [{ name: 'Dashboard', to: '/dashboard', icon: 'dashboard' }] },
  {
    title: 'Catalog',
    items: [
      { name: 'Products', to: '/products', icon: 'cube', perm: 'product.manage' },
      { name: 'Categories', to: '/categories', icon: 'tag', perm: 'product.manage' },
      { name: 'Suppliers', to: '/suppliers', icon: 'truck', perm: 'product.manage' },
    ],
  },
  {
    title: 'Purchases',
    items: [
      { name: 'Purchases', to: '/purchases', icon: 'bag', perm: 'purchase.manage' },
      { name: 'Purchase Returns', to: '/returns?tab=purchase', icon: 'uturnLeft', perm: 'purchase.manage' },
    ],
  },
  {
    title: 'Sales',
    items: [
      { name: 'Customers', to: '/customers', icon: 'user', perm: 'sales.manage' },
      { name: 'Sales', to: '/sales', icon: 'cart', perm: 'sales.manage' },
      { name: 'Sales Returns', to: '/returns?tab=sale', icon: 'uturnRight', perm: 'sales.manage' },
    ],
  },
  {
    title: 'Finance',
    items: [
      { name: 'Payments', to: '/payments', icon: 'card', anyPerm: ['sales.manage', 'purchase.manage'] },
      { name: 'Ledger', to: '/ledger', icon: 'book', perm: 'report.access' },
      { name: 'Reports', to: '/reports', icon: 'chart', perm: 'report.access' },
    ],
  },
  { title: 'Administration', items: [{ name: 'Users', to: '/users', icon: 'users', perm: 'user.manage' }] },
]

function canSee(item) {
  if (item.anyPerm) return item.anyPerm.some((p) => auth.can(p))
  return !item.perm || auth.can(item.perm)
}

const visibleGroups = computed(() =>
  groups.map((g) => ({ ...g, items: g.items.filter(canSee) })).filter((g) => g.items.length),
)

const collapsed = reactive({})
const toggle = (title) => {
  if (title) collapsed[title] = !collapsed[title]
}

function isActive(to) {
  const [path, qs] = to.split('?')
  if (route.path !== path) return false
  if (!qs) return true
  return route.query.tab === new URLSearchParams(qs).get('tab')
}
const groupActive = (group) => group.items.some((i) => isActive(i.to))

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
        <span class="grid h-9 w-9 place-items-center rounded-xl bg-gradient-to-br from-brand-500 to-indigo-700 text-white shadow-lg shadow-brand-600/30">
          <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8"><path stroke-linecap="round" stroke-linejoin="round" d="m21 7.5-9-5.25L3 7.5m18 0-9 5.25m9-5.25v9l-9 5.25M3 7.5l9 5.25M3 7.5v9l9 5.25m0-9v9" /></svg>
        </span>
        <div class="leading-tight">
          <div class="font-bold tracking-tight">Inventory</div>
          <div class="text-[11px] text-slate-400">Management System</div>
        </div>
      </div>

      <!-- Nav -->
      <nav class="flex-1 space-y-1 overflow-y-auto px-3 py-4">
        <div v-for="group in visibleGroups" :key="group.title || 'main'" :class="group.title ? 'pt-2' : ''">
          <!-- Group header -->
          <button
            v-if="group.title"
            class="mb-0.5 flex w-full items-center justify-between rounded-lg px-3 py-1.5 text-[11px] font-bold uppercase tracking-wider transition hover:bg-slate-100 dark:hover:bg-slate-700/50"
            :class="groupActive(group) ? 'text-brand-600 dark:text-brand-300' : 'text-slate-400'"
            @click="toggle(group.title)"
          >
            <span class="flex items-center gap-2">
              <span v-if="groupActive(group)" class="h-1.5 w-1.5 rounded-full bg-brand-500" />
              {{ group.title }}
            </span>
            <svg class="h-3.5 w-3.5 transition-transform duration-300" :class="collapsed[group.title] ? '-rotate-90' : ''" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5"><path stroke-linecap="round" stroke-linejoin="round" d="m19 9-7 7-7-7" /></svg>
          </button>

          <!-- Collapsible items (smooth grid-rows animation) -->
          <div class="grid transition-all duration-300 ease-in-out" :class="group.title && collapsed[group.title] ? 'grid-rows-[0fr] opacity-0' : 'grid-rows-[1fr] opacity-100'">
            <div class="overflow-hidden">
              <div class="space-y-0.5 py-0.5">
                <RouterLink
                  v-for="item in group.items"
                  :key="item.to"
                  :to="item.to"
                  class="group relative flex items-center gap-3 rounded-xl px-3 py-2.5 text-sm font-medium transition-all duration-200"
                  :class="isActive(item.to)
                    ? 'bg-gradient-to-r from-brand-600 to-indigo-600 text-white shadow-lg shadow-brand-600/30'
                    : 'text-slate-600 hover:bg-slate-100 hover:text-slate-900 dark:text-slate-300 dark:hover:bg-slate-700/60 dark:hover:text-white'"
                  @click="sidebarOpen = false"
                >
                  <!-- active accent bar -->
                  <span v-if="isActive(item.to)" class="absolute -left-3 top-1/2 h-6 w-1 -translate-y-1/2 rounded-r-full bg-white/90" />
                  <svg
                    class="h-5 w-5 shrink-0 transition-transform duration-200 group-hover:scale-110"
                    :class="isActive(item.to) ? 'text-white' : 'text-slate-400 group-hover:text-brand-500'"
                    fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.6"
                  >
                    <path v-for="(d, i) in ICONS[item.icon]" :key="i" stroke-linecap="round" stroke-linejoin="round" :d="d" />
                  </svg>
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
            <svg class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8"><path stroke-linecap="round" stroke-linejoin="round" d="M15.75 9V5.25A2.25 2.25 0 0 0 13.5 3h-6a2.25 2.25 0 0 0-2.25 2.25v13.5A2.25 2.25 0 0 0 7.5 21h6a2.25 2.25 0 0 0 2.25-2.25V15m3 0 3-3m0 0-3-3m3 3H9" /></svg>
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
        <button class="rounded-lg p-2 text-slate-500 transition hover:bg-slate-100 dark:text-slate-300 dark:hover:bg-slate-700" :title="isDark ? 'Light mode' : 'Dark mode'" @click="toggleTheme">
          <svg v-if="isDark" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8"><path stroke-linecap="round" stroke-linejoin="round" d="M12 3v2.25m6.364.386-1.591 1.591M21 12h-2.25m-.386 6.364-1.591-1.591M12 18.75V21m-4.773-4.227-1.591 1.591M5.25 12H3m4.227-4.773L5.636 5.636M15.75 12a3.75 3.75 0 1 1-7.5 0 3.75 3.75 0 0 1 7.5 0Z" /></svg>
          <svg v-else class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="1.8"><path stroke-linecap="round" stroke-linejoin="round" d="M21.752 15.002A9.718 9.718 0 0 1 18 15.75c-5.385 0-9.75-4.365-9.75-9.75 0-1.33.266-2.597.748-3.752A9.753 9.753 0 0 0 3 11.25C3 16.635 7.365 21 12.75 21a9.753 9.753 0 0 0 9.002-5.998Z" /></svg>
        </button>
      </header>

      <main class="flex-1 p-4 lg:p-8">
        <RouterView />
      </main>
    </div>
  </div>
</template>
