import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'

import DashboardLayout from '../layouts/DashboardLayout.vue'
import LoginView from '../views/LoginView.vue'
import DashboardView from '../views/DashboardView.vue'
import ProductsView from '../views/ProductsView.vue'
import CategoriesView from '../views/CategoriesView.vue'
import SuppliersView from '../views/SuppliersView.vue'
import CustomersView from '../views/CustomersView.vue'
import PurchasesView from '../views/PurchasesView.vue'
import SalesView from '../views/SalesView.vue'
import ReportsView from '../views/ReportsView.vue'
import PaymentsView from '../views/PaymentsView.vue'
import ReturnsView from '../views/ReturnsView.vue'
import LedgerView from '../views/LedgerView.vue'
import UsersView from '../views/UsersView.vue'

const routes = [
  { path: '/login', name: 'login', component: LoginView, meta: { public: true } },
  {
    path: '/',
    component: DashboardLayout,
    children: [
      { path: '', redirect: '/dashboard' },
      { path: 'dashboard', name: 'dashboard', component: DashboardView },
      { path: 'products', name: 'products', component: ProductsView, meta: { perm: 'product.manage' } },
      { path: 'categories', name: 'categories', component: CategoriesView, meta: { perm: 'product.manage' } },
      { path: 'suppliers', name: 'suppliers', component: SuppliersView, meta: { perm: 'product.manage' } },
      { path: 'customers', name: 'customers', component: CustomersView, meta: { perm: 'sales.manage' } },
      { path: 'purchases', name: 'purchases', component: PurchasesView, meta: { perm: 'purchase.manage' } },
      { path: 'sales', name: 'sales', component: SalesView, meta: { perm: 'sales.manage' } },
      { path: 'reports', redirect: '/reports/sales' },
      { path: 'reports/:type', name: 'reports', component: ReportsView, meta: { perm: 'report.access' } },
      { path: 'payments', name: 'payments', component: PaymentsView, meta: { anyPerm: ['sales.manage', 'purchase.manage'] } },
      { path: 'returns', name: 'returns', component: ReturnsView, meta: { anyPerm: ['sales.manage', 'purchase.manage'] } },
      { path: 'ledger', redirect: '/ledger/supplier' },
      { path: 'ledger/:mode', name: 'ledger', component: LedgerView, meta: { perm: 'report.access' } },
      { path: 'users', name: 'users', component: UsersView, meta: { perm: 'user.manage' } },
    ],
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

// Global guard: redirect to /login when unauthenticated; block pages the user's
// role lacks permission for.
router.beforeEach((to) => {
  const auth = useAuthStore()
  if (to.meta.public) return true
  if (!auth.isAuthenticated) return { name: 'login' }
  if (to.meta.perm && !auth.can(to.meta.perm)) return { name: 'dashboard' }
  if (to.meta.anyPerm && !to.meta.anyPerm.some((p) => auth.can(p))) return { name: 'dashboard' }
  return true
})

export default router
