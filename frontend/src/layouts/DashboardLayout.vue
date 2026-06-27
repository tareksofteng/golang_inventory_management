<script setup>
import { ref, computed, onMounted } from 'vue'
import { RouterView, RouterLink, useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const router = useRouter()

const sidebarOpen = ref(false)
const isDark = ref(document.documentElement.classList.contains('dark'))

const nav = computed(() =>
  [
    { name: 'Dashboard', to: '/dashboard', icon: '📊', perm: null },
    { name: 'Products', to: '/products', icon: '📦', perm: 'product.manage' },
    { name: 'Categories', to: '/categories', icon: '🏷️', perm: 'product.manage' },
    { name: 'Suppliers', to: '/suppliers', icon: '🚚', perm: 'product.manage' },
    { name: 'Customers', to: '/customers', icon: '👤', perm: 'sales.manage' },
    { name: 'Purchases', to: '/purchases', icon: '🧾', perm: 'purchase.manage' },
    { name: 'Sales', to: '/sales', icon: '🛒', perm: 'sales.manage' },
    { name: 'Reports', to: '/reports', icon: '📑', perm: 'report.access' },
    { name: 'Payments', to: '/payments', icon: '💳', anyPerm: ['sales.manage', 'purchase.manage'] },
    { name: 'Users', to: '/users', icon: '🔐', perm: 'user.manage' },
  ].filter((i) => {
    if (i.anyPerm) return i.anyPerm.some((p) => auth.can(p))
    return !i.perm || auth.can(i.perm)
  }),
)

function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

async function logout() {
  await auth.logout()
  router.push('/login')
}

onMounted(() => {
  // Refresh profile/permissions in the background on first load.
  if (auth.isAuthenticated) auth.fetchMe().catch(() => {})
})
</script>

<template>
  <div class="min-h-screen lg:flex">
    <!-- Sidebar -->
    <aside
      :class="sidebarOpen ? 'translate-x-0' : '-translate-x-full'"
      class="fixed inset-y-0 left-0 z-30 w-64 transform border-r border-slate-200 bg-white transition-transform dark:border-slate-700 dark:bg-slate-800 lg:static lg:translate-x-0"
    >
      <div class="flex h-16 items-center gap-2 border-b border-slate-200 px-6 dark:border-slate-700">
        <span class="grid h-9 w-9 place-items-center rounded-lg bg-brand-600 text-white">📦</span>
        <span class="text-lg font-bold">Inventory</span>
      </div>
      <nav class="space-y-1 p-4">
        <RouterLink
          v-for="item in nav"
          :key="item.to"
          :to="item.to"
          class="flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium text-slate-600 hover:bg-slate-100 dark:text-slate-300 dark:hover:bg-slate-700"
          active-class="!bg-brand-50 !text-brand-700 dark:!bg-brand-600/20 dark:!text-brand-100"
          @click="sidebarOpen = false"
        >
          <span>{{ item.icon }}</span>{{ item.name }}
        </RouterLink>
      </nav>
    </aside>

    <!-- Backdrop (mobile) -->
    <div
      v-if="sidebarOpen"
      class="fixed inset-0 z-20 bg-black/30 lg:hidden"
      @click="sidebarOpen = false"
    />

    <!-- Main -->
    <div class="flex min-h-screen flex-1 flex-col">
      <header
        class="sticky top-0 z-10 flex h-16 items-center justify-between border-b border-slate-200 bg-white/80 px-4 backdrop-blur dark:border-slate-700 dark:bg-slate-800/80 lg:px-8"
      >
        <button class="btn-ghost lg:hidden" @click="sidebarOpen = true">☰</button>
        <div class="hidden font-semibold lg:block">Inventory Management System</div>
        <div class="flex items-center gap-3">
          <button class="btn-ghost !px-2.5" :title="isDark ? 'Light' : 'Dark'" @click="toggleTheme">
            {{ isDark ? '☀️' : '🌙' }}
          </button>
          <div class="text-right leading-tight">
            <div class="text-sm font-medium">{{ auth.user?.name }}</div>
            <div class="text-xs capitalize text-slate-500">{{ auth.user?.role?.replace('_', ' ') }}</div>
          </div>
          <button class="btn-danger !px-3 !py-1.5 text-xs" @click="logout">Logout</button>
        </div>
      </header>

      <main class="flex-1 p-4 lg:p-8">
        <RouterView />
      </main>
    </div>
  </div>
</template>
