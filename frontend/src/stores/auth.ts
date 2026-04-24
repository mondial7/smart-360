import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { User } from '@/types/user'
import apiClient from '@/api/client'

export const useAuthStore = defineStore('auth', () => {
  // State
  const user = ref<User | null>(null)
  const token = ref<string | null>(localStorage.getItem('token'))
  const loading = ref(false)

  // Getters
  const isAuthenticated = computed(() => !!token.value && !!user.value)
  const isAdmin = computed(() => user.value?.role === 'admin')
  const isTeamAdmin = computed(() => user.value?.role === 'team_admin')
  const canManageTeams = computed(() => isAdmin.value || isTeamAdmin.value)

  // Actions
  async function init() {
    if (token.value) {
      await fetchUser()
    }
  }

  async function fetchUser() {
    loading.value = true
    try {
      const response = await apiClient.get('/me')
      user.value = response.data
    } catch {
      logout()
    } finally {
      loading.value = false
    }
  }

  function setToken(newToken: string) {
    token.value = newToken
    localStorage.setItem('token', newToken)
  }

  function logout() {
    user.value = null
    token.value = null
    localStorage.removeItem('token')
  }

  async function loginWithGoogle() {
    const response = await apiClient.get('/auth/google')
    window.location.href = response.data.url
  }

  async function devLogin(email?: string) {
    // Redirect to dev login endpoint which will redirect back with token
    const url = email ? `/api/auth/dev-login?email=${encodeURIComponent(email)}` : '/api/auth/dev-login'
    window.location.href = url
  }

  return {
    user,
    token,
    loading,
    isAuthenticated,
    isAdmin,
    isTeamAdmin,
    canManageTeams,
    init,
    fetchUser,
    setToken,
    logout,
    loginWithGoogle,
    devLogin
  }
})
