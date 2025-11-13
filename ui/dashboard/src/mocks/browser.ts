// Browser setup for MSW

import { setupWorker } from 'msw/browser'
import { handlers } from './handlers'

// Setup the service worker with the handlers
export const worker = setupWorker(...handlers)

// Start the worker with optional configuration
export async function startMockServiceWorker() {
  if (import.meta.env.DEV || import.meta.env.VITE_ENABLE_MOCKS === 'true') {
    await worker.start({
      onUnhandledRequest: 'bypass', // Don't warn about unhandled requests
      serviceWorker: {
        url: '/mockServiceWorker.js',
      },
    })
    console.log('[MSW] Mocking enabled - API requests will be intercepted')
  }
}
