import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import type { AxiosAdapter } from 'axios'
import apiClient from './client'

// Drive the client through a fake adapter so the request/response
// interceptors run exactly as they would against a real server, without
// any network. The adapter is the last link in axios's chain, so anything
// the request interceptor did is visible on the config it receives.
describe('apiClient interceptors', () => {
  const originalAdapter = apiClient.defaults.adapter

  beforeEach(() => {
    localStorage.clear()
  })

  afterEach(() => {
    apiClient.defaults.adapter = originalAdapter
    vi.unstubAllGlobals()
  })

  it('attaches the bearer token from localStorage to requests', async () => {
    localStorage.setItem('token', 'jwt-abc')
    const adapter: AxiosAdapter = async (config) => ({
      data: null,
      status: 200,
      statusText: 'OK',
      headers: {},
      config,
    })
    apiClient.defaults.adapter = adapter

    const res = await apiClient.get('/me')
    expect(res.config.headers.Authorization).toBe('Bearer jwt-abc')
  })

  it('sends no Authorization header when there is no token', async () => {
    const adapter: AxiosAdapter = async (config) => ({
      data: null,
      status: 200,
      statusText: 'OK',
      headers: {},
      config,
    })
    apiClient.defaults.adapter = adapter

    const res = await apiClient.get('/me')
    expect(res.config.headers.Authorization).toBeUndefined()
  })

  it('clears the token and redirects to /login on a 401', async () => {
    vi.stubGlobal('location', { href: '' } as Location)
    localStorage.setItem('token', 'jwt-abc')

    apiClient.defaults.adapter = async (config) => {
      const error = new Error('Unauthorized') as Error & { response?: unknown; config?: unknown }
      error.config = config
      error.response = { status: 401, data: null, statusText: 'Unauthorized', headers: {}, config }
      throw error
    }

    await expect(apiClient.get('/me')).rejects.toBeTruthy()
    expect(localStorage.getItem('token')).toBeNull()
    expect(window.location.href).toContain('/login')
  })
})
