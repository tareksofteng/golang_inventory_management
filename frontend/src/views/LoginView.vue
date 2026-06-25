<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const auth = useAuthStore()
const router = useRouter()

const email = ref('admin@inventory.test')
const password = ref('Admin@123')
const error = ref('')
const loading = ref(false)

async function submit() {
  error.value = ''
  loading.value = true
  try {
    await auth.login(email.value, password.value)
    router.push('/dashboard')
  } catch (e) {
    error.value = e.response?.data?.message || 'Login failed'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="flex min-h-screen items-center justify-center bg-gradient-to-br from-brand-600 to-indigo-800 p-4">
    <div class="card w-full max-w-md p-8">
      <div class="mb-6 text-center">
        <span class="mx-auto mb-3 grid h-14 w-14 place-items-center rounded-2xl bg-brand-600 text-2xl text-white">📦</span>
        <h1 class="text-2xl font-bold">Inventory Admin</h1>
        <p class="text-sm text-slate-500">Sign in to your account</p>
      </div>

      <form class="space-y-4" @submit.prevent="submit">
        <div v-if="error" class="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-600 dark:bg-red-500/10">
          {{ error }}
        </div>
        <div>
          <label class="label">Email</label>
          <input v-model="email" type="email" class="input" required />
        </div>
        <div>
          <label class="label">Password</label>
          <input v-model="password" type="password" class="input" required />
        </div>
        <button type="submit" class="btn-primary w-full" :disabled="loading">
          {{ loading ? 'Signing in…' : 'Sign In' }}
        </button>
      </form>

      <p class="mt-4 text-center text-xs text-slate-400">
        Demo: admin@inventory.test / Admin@123
      </p>
    </div>
  </div>
</template>
