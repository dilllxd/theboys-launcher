import '@testing-library/jest-dom'
import { beforeAll, afterEach, afterAll } from 'vitest'
import { setupServer } from 'msw/node'
import { rest } from 'msw'

// Mock Tauri API
const mockTauri = {
  invoke: vi.fn(),
  listen: vi.fn(),
  emit: vi.fn(),
}

global.__TAURI__ = {
  invoke: mockTauri.invoke,
  listen: mockTauri.listen,
  emit: mockTauri.emit,
  tauri: {
    invoke: mockTauri.invoke,
  }
}

// Mock file system API
global.__TAURI_FS__ = {
  exists: vi.fn(),
  readTextFile: vi.fn(),
  writeTextFile: vi.fn(),
  createDir: vi.fn(),
  removeDir: vi.fn(),
}

// Mock dialog API
global.__TAURI_DIALOG__ = {
  open: vi.fn(),
  save: vi.fn(),
  ask: vi.fn(),
  confirm: vi.fn(),
}

// Mock window API
global.__TAURI_WINDOW__ = {
  getCurrent: vi.fn(() => ({
    listen: vi.fn(),
    emit: vi.fn(),
  })),
}

// Setup MSW server for mocking HTTP requests
export const server = setupServer(
  rest.get('/api/health', (req, res, ctx) => {
    return res(ctx.status(200), ctx.json({ status: 'OK' }))
  }),

  rest.get('/api/modpacks', (req, res, ctx) => {
    return res(ctx.status(200), ctx.json([]))
  }),

  rest.get('/api/instances', (req, res, ctx) => {
    return res(ctx.status(200), ctx.json([]))
  })
)

beforeAll(() => {
  server.listen()
})

afterEach(() => {
  server.resetHandlers()
  vi.clearAllMocks()
})

afterAll(() => {
  server.close()
})

// Mock IntersectionObserver
global.IntersectionObserver = vi.fn().mockImplementation(() => ({
  observe: vi.fn(),
  unobserve: vi.fn(),
  disconnect: vi.fn(),
}))

// Mock ResizeObserver
global.ResizeObserver = vi.fn().mockImplementation(() => ({
  observe: vi.fn(),
  unobserve: vi.fn(),
  disconnect: vi.fn(),
}))

// Mock window.matchMedia
Object.defineProperty(window, 'matchMedia', {
  writable: true,
  value: vi.fn().mockImplementation(query => ({
    matches: false,
    media: query,
    onchange: null,
    addListener: vi.fn(), // deprecated
    removeListener: vi.fn(), // deprecated
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
    dispatchEvent: vi.fn(),
  })),
})

// Add custom matchers for better test assertions
expect.extend({
  toBeValidEmail(received: string) {
    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
    const pass = emailRegex.test(received)
    if (pass) {
      return {
        message: () => `expected ${received} not to be a valid email`,
        pass: true,
      }
    } else {
      return {
        message: () => `expected ${received} to be a valid email`,
        pass: false,
      }
    }
  },

  toBeValidUrl(received: string) {
    try {
      new URL(received)
      return {
        message: () => `expected ${received} not to be a valid URL`,
        pass: true,
      }
    } catch {
      return {
        message: () => `expected ${received} to be a valid URL`,
        pass: false,
      }
    }
  }
})

declare global {
  namespace Vi {
    interface Assertion {
      toBeValidEmail(): boolean
      toBeValidUrl(): boolean
    }
  }
}