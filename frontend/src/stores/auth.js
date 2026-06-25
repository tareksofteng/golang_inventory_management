import { defineStore } from 'pinia'
import api from '../lib/api'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    user: JSON.parse(localStorage.getItem('user') || 'null'),
    permissions: JSON.parse(localStorage.getItem('permissions') || '[]'),
    accessToken: localStorage.getItem('access_token') || '',
  }),

  getters: {
    isAuthenticated: (s) => !!s.accessToken,
    // can('user.manage') -> true/false, used to show/hide menu items.
    can: (s) => (perm) => s.permissions.includes(perm),
  },

  actions: {
    async login(email, password) {
      const { data } = await api.post('/auth/login', { email, password })
      const d = data.data
      this.user = d.user
      this.permissions = d.permissions || []
      this.accessToken = d.access_token

      localStorage.setItem('access_token', d.access_token)
      localStorage.setItem('refresh_token', d.refresh_token)
      localStorage.setItem('user', JSON.stringify(d.user))
      localStorage.setItem('permissions', JSON.stringify(this.permissions))
    },

    async fetchMe() {
      const { data } = await api.get('/auth/me')
      this.user = data.data.user
      this.permissions = data.data.permissions || []
      localStorage.setItem('user', JSON.stringify(this.user))
      localStorage.setItem('permissions', JSON.stringify(this.permissions))
    },

    async logout() {
      const refresh = localStorage.getItem('refresh_token')
      try {
        if (refresh) await api.post('/auth/logout', { refresh_token: refresh })
      } catch (_) {
        /* ignore — clear locally regardless */
      }
      this.$reset()
      localStorage.clear()
    },
  },
})
