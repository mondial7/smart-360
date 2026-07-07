import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'

// Mock the API client the store depends on. Declared before the store import
// so the store picks up the mock.
vi.mock('@/api/client', () => ({
  default: { get: vi.fn(), post: vi.fn() },
}))

import apiClient from '@/api/client'
import { useAuthStore } from './auth'
import type { User } from '@/types/user'

const mockGet = apiClient.get as unknown as ReturnType<typeof vi.fn>

function makeUser(role: User['role']): User {
  return { id: '1', email: 'a@b.com', name: 'Ada', role } as User
}

describe('auth store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    mockGet.mockReset()
  })

  it('starts unauthenticated with no token', () => {
    const auth = useAuthStore()
    expect(auth.token).toBeNull()
    expect(auth.isAuthenticated).toBe(false)
    expect(auth.isAdmin).toBe(false)
  })

  it('derives role getters from the current user', () => {
    const auth = useAuthStore()

    auth.user = makeUser('admin')
    expect(auth.isAdmin).toBe(true)
    expect(auth.canManageTeams).toBe(true)

    auth.user = makeUser('team_admin')
    expect(auth.isAdmin).toBe(false)
    expect(auth.isTeamAdmin).toBe(true)
    expect(auth.canManageTeams).toBe(true)

    auth.user = makeUser('member')
    expect(auth.canManageTeams).toBe(false)
  })

  it('is authenticated only when both token and user are present', () => {
    const auth = useAuthStore()
    auth.setToken('jwt-123')
    expect(auth.isAuthenticated).toBe(false) // no user yet
    auth.user = makeUser('member')
    expect(auth.isAuthenticated).toBe(true)
  })

  it('setToken persists the token to localStorage', () => {
    const auth = useAuthStore()
    auth.setToken('jwt-123')
    expect(auth.token).toBe('jwt-123')
    expect(localStorage.getItem('token')).toBe('jwt-123')
  })

  it('logout clears user, token, and localStorage', () => {
    const auth = useAuthStore()
    auth.setToken('jwt-123')
    auth.user = makeUser('admin')

    auth.logout()

    expect(auth.user).toBeNull()
    expect(auth.token).toBeNull()
    expect(localStorage.getItem('token')).toBeNull()
    expect(auth.isAuthenticated).toBe(false)
  })

  it('fetchUser populates the user on success', async () => {
    const user = makeUser('admin')
    mockGet.mockResolvedValueOnce({ data: user })

    const auth = useAuthStore()
    await auth.fetchUser()

    expect(mockGet).toHaveBeenCalledWith('/me')
    expect(auth.user).toEqual(user)
    expect(auth.loading).toBe(false)
  })

  it('fetchUser logs out when the request fails', async () => {
    mockGet.mockRejectedValueOnce(new Error('401'))

    const auth = useAuthStore()
    auth.setToken('jwt-123')
    await auth.fetchUser()

    expect(auth.user).toBeNull()
    expect(auth.token).toBeNull()
    expect(auth.loading).toBe(false)
  })
})
