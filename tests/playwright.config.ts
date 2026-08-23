import { defineConfig } from '@playwright/test'

export default defineConfig({
  testDir: '.',
  testMatch: 'e2e_flow.spec.ts',
  timeout: 60_000,
  use: {
    baseURL: process.env.E2E_BASE || 'http://frontend',
    viewport: { width: 1440, height: 900 },
  },
})
