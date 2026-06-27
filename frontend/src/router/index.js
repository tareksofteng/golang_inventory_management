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
  return true
})

export default router
