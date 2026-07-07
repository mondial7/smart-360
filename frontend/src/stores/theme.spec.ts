import { beforeEach, describe, expect, it } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useThemeStore } from './theme'

describe('theme store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    document.documentElement.classList.remove('dark')
  })

  it('defaults to the system preference (light) when untouched', () => {
    const theme = useThemeStore()
    expect(theme.preference).toBe('system')
    expect(theme.activeTheme).toBe('light')
    expect(theme.isDark).toBe(false)
  })

  it('setTheme persists the preference and reflects it in getters', () => {
    const theme = useThemeStore()
    theme.setTheme('dark')

    expect(theme.preference).toBe('dark')
    expect(theme.activeTheme).toBe('dark')
    expect(theme.isDark).toBe(true)
    expect(localStorage.getItem('theme-preference')).toBe('dark')
  })

  it('applies and removes the dark class on <html>', () => {
    const theme = useThemeStore()

    theme.setTheme('dark')
    expect(document.documentElement.classList.contains('dark')).toBe(true)

    theme.setTheme('light')
    expect(document.documentElement.classList.contains('dark')).toBe(false)
  })

  it('toggleTheme flips a system default to dark', () => {
    const theme = useThemeStore()
    theme.toggleTheme()
    expect(theme.preference).toBe('dark')
    expect(theme.isDark).toBe(true)
  })
})
