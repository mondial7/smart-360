import { afterEach, beforeEach, vi } from 'vitest'

// A minimal, deterministic localStorage. We install our own rather than
// trust the ambient one: jsdom provides a Storage, but newer Node versions
// (25+) expose a native experimental `localStorage` global that shadows it
// and behaves inconsistently. An in-memory map keeps tests hermetic.
class MemoryStorage implements Storage {
  private store = new Map<string, string>()
  get length() {
    return this.store.size
  }
  clear() {
    this.store.clear()
  }
  getItem(key: string) {
    return this.store.has(key) ? this.store.get(key)! : null
  }
  key(index: number) {
    return Array.from(this.store.keys())[index] ?? null
  }
  removeItem(key: string) {
    this.store.delete(key)
  }
  setItem(key: string, value: string) {
    this.store.set(key, String(value))
  }
}

const memoryStorage = new MemoryStorage()
vi.stubGlobal('localStorage', memoryStorage)

// jsdom doesn't implement matchMedia; the theme store touches it on init().
// Provide a permissive stub so tests that don't care about it don't crash.
if (!window.matchMedia) {
  window.matchMedia = vi.fn().mockImplementation((query: string) => ({
    matches: false,
    media: query,
    onchange: null,
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    addListener: vi.fn(),
    removeListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })) as unknown as typeof window.matchMedia
}

// Isolate tests: no shared localStorage or mock state leaks between cases.
// Re-stub each time so a test that calls vi.unstubAllGlobals() (e.g. to clean
// up a `location` stub) doesn't leave the next test with the broken native
// localStorage.
beforeEach(() => {
  vi.stubGlobal('localStorage', memoryStorage)
  memoryStorage.clear()
})

afterEach(() => {
  vi.restoreAllMocks()
})
