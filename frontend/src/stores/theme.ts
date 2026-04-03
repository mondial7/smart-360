import { defineStore } from 'pinia'
import { ref, computed, watch } from 'vue'

type Theme = 'light' | 'dark' | 'system'

export const useThemeStore = defineStore('theme', () => {
  // State
  const preference = ref<Theme>('system')
  const systemPreference = ref<'light' | 'dark'>('light')

  // Getters
  const activeTheme = computed<'light' | 'dark'>(() => {
    if (preference.value === 'system') {
      return systemPreference.value
    }
    return preference.value
  })

  const isDark = computed(() => activeTheme.value === 'dark')

  // Actions
  function init() {
    // Load saved preference from localStorage
    const saved = localStorage.getItem('theme-preference')
    if (saved && (saved === 'light' || saved === 'dark' || saved === 'system')) {
      preference.value = saved as Theme
    }

    // Detect system preference
    detectSystemPreference()

    // Listen for system preference changes
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
    mediaQuery.addEventListener('change', handleSystemPreferenceChange)

    // Apply initial theme
    applyTheme()
  }

  function detectSystemPreference() {
    const isDarkMode = window.matchMedia('(prefers-color-scheme: dark)').matches
    systemPreference.value = isDarkMode ? 'dark' : 'light'
  }

  function handleSystemPreferenceChange(e: MediaQueryListEvent) {
    systemPreference.value = e.matches ? 'dark' : 'light'
  }

  function setTheme(theme: Theme) {
    preference.value = theme
    localStorage.setItem('theme-preference', theme)
    applyTheme()
  }

  function toggleTheme() {
    if (preference.value === 'system') {
      // If currently system, toggle to opposite of current system
      setTheme(systemPreference.value === 'dark' ? 'light' : 'dark')
    } else {
      // Toggle between light and dark
      setTheme(preference.value === 'dark' ? 'light' : 'dark')
    }
  }

  function applyTheme() {
    const html = document.documentElement

    if (activeTheme.value === 'dark') {
      html.classList.add('dark')
    } else {
      html.classList.remove('dark')
    }

    // Update meta theme-color for mobile browsers
    updateMetaThemeColor()
  }

  function updateMetaThemeColor() {
    const metaThemeColor = document.querySelector('meta[name="theme-color"]')
    const color = activeTheme.value === 'dark' ? '#0f172a' : '#ffffff'

    if (metaThemeColor) {
      metaThemeColor.setAttribute('content', color)
    } else {
      const meta = document.createElement('meta')
      meta.name = 'theme-color'
      meta.content = color
      document.head.appendChild(meta)
    }
  }

  // Watch for changes and reapply
  watch([preference, systemPreference], () => {
    applyTheme()
  })

  return {
    preference,
    activeTheme,
    isDark,
    init,
    setTheme,
    toggleTheme,
  }
})
