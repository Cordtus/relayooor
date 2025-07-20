import { vi } from 'vitest'

// Mock import.meta.env
vi.stubGlobal('import.meta', {
  env: {
    VITE_API_BASE_URL: 'http://localhost:8080',
    VITE_WS_HOST: 'localhost:8080',
    MODE: 'test',
    DEV: false,
    PROD: false,
    SSR: false
  }
})

// Mock window.location
Object.defineProperty(window, 'location', {
  value: {
    protocol: 'http:',
    host: 'localhost:5173',
    hostname: 'localhost',
    port: '5173',
    pathname: '/',
    search: '',
    hash: ''
  },
  writable: true
})