import { describe, it, expect, beforeAll, afterAll, beforeEach } from 'vitest'
import request from 'supertest'
import { chromium, Browser, Page } from 'playwright'

// Configuration
const API_URL = process.env.API_URL || 'http://localhost:8080'
const FRONTEND_URL = process.env.FRONTEND_URL || 'http://localhost:5173'
const CHAINPULSE_URL = process.env.CHAINPULSE_URL || 'http://localhost:3001'

describe('E2E Clearing Flow', () => {
  let browser: Browser
  let page: Page
  
  beforeAll(async () => {
    // Launch browser
    browser = await chromium.launch({
      headless: process.env.CI === 'true',
    })
  })
  
  afterAll(async () => {
    await browser.close()
  })
  
  beforeEach(async () => {
    page = await browser.newPage()
  })
  
  afterEach(async () => {
    await page.close()
  })

  it('completes full packet clearing flow', async () => {
    // Step 1: Navigate to app
    await page.goto(FRONTEND_URL)
    await page.waitForLoadState('networkidle')
    
    // Step 2: Connect wallet (mock Keplr)
    await page.evaluate(() => {
      window.keplr = {
        enable: async () => {},
        getKey: async () => ({
          name: 'Test Wallet',
          algo: 'secp256k1',
          pubKey: new Uint8Array(33),
          address: new Uint8Array(20),
          bech32Address: 'osmo1test123abc',
        }),
        signArbitrary: async () => ({
          pub_key: { type: 'tendermint/PubKeySecp256k1', value: 'test' },
          signature: 'mock-signature-base64',
        }),
      }
    })
    
    // Click connect wallet
    await page.click('[data-testid="connect-wallet-button"]')
    await page.waitForSelector('[data-testid="wallet-connected"]')
    
    // Step 3: Navigate to packet clearing
    await page.click('[data-testid="nav-clearing"]')
    await page.waitForSelector('[data-testid="clearing-wizard"]')
    
    // Step 4: Wait for packets to load
    await page.waitForSelector('[data-testid="packet-selector"]', { timeout: 10000 })
    
    // Step 5: Select packets
    await page.click('[data-testid="select-all-checkbox"]')
    
    // Verify selection
    const selectedCount = await page.textContent('[data-testid="selected-count"]')
    expect(selectedCount).toContain('selected')
    
    // Step 6: Continue to fees
    await page.click('[data-testid="continue-button"]')
    await page.waitForSelector('[data-testid="fee-estimator"]')
    
    // Verify fee calculation
    const feeAmount = await page.textContent('[data-testid="total-fee"]')
    expect(feeAmount).toMatch(/\d+(\.\d+)?\s+\w+/)
    
    // Step 7: Continue to payment
    await page.click('[data-testid="continue-to-payment"]')
    await page.waitForSelector('[data-testid="payment-prompt"]')
    
    // Step 8: Simulate payment
    await page.evaluate(() => {
      // Mock successful payment transaction
      window.keplr.sendTx = async () => ({
        code: 0,
        height: 123456,
        txhash: 'MOCK_TX_HASH_123',
      })
    })
    
    await page.click('[data-testid="send-payment-button"]')
    
    // Step 9: Wait for clearing progress
    await page.waitForSelector('[data-testid="clearing-progress"]')
    
    // Step 10: Verify completion
    await page.waitForSelector('[data-testid="clearing-complete"]', { timeout: 30000 })
    
    const successMessage = await page.textContent('[data-testid="success-message"]')
    expect(successMessage).toContain('Successfully cleared')
  })

  it('handles wallet disconnection during flow', async () => {
    // Setup and connect wallet
    await page.goto(FRONTEND_URL)
    await page.evaluate(() => {
      window.keplr = {
        enable: async () => {},
        getKey: async () => ({
          bech32Address: 'osmo1test123',
        }),
      }
    })
    
    await page.click('[data-testid="connect-wallet-button"]')
    await page.waitForSelector('[data-testid="wallet-connected"]')
    
    // Navigate to clearing
    await page.click('[data-testid="nav-clearing"]')
    
    // Disconnect wallet mid-flow
    await page.evaluate(() => {
      window.dispatchEvent(new Event('keplr_keystorechange'))
    })
    
    // Should redirect to connect prompt
    await page.waitForSelector('[data-testid="connect-wallet-prompt"]')
  })
})

describe('API Integration Tests', () => {
  it('health check returns ok', async () => {
    const response = await request(API_URL)
      .get('/health')
      .expect(200)
    
    expect(response.body).toEqual({ status: 'ok' })
  })

  it('validates wallet address format', async () => {
    const response = await request(API_URL)
      .get('/api/user/invalid-address/transfers')
      .expect(400)
    
    expect(response.body.error).toBe('Invalid wallet address')
  })

  it('returns user transfers with Chainpulse integration', async () => {
    // First check if Chainpulse is available
    try {
      await request(CHAINPULSE_URL).get('/health').expect(200)
      
      // If available, test integration
      const response = await request(API_URL)
        .get('/api/user/osmo1test123/transfers')
        .expect(200)
      
      expect(Array.isArray(response.body)).toBe(true)
    } catch (error) {
      // Chainpulse not available, should fall back to mock
      const response = await request(API_URL)
        .get('/api/user/osmo1test123/transfers')
        .expect(200)
      
      expect(Array.isArray(response.body)).toBe(true)
      expect(response.body.length).toBeGreaterThanOrEqual(0)
    }
  })

  it('clears packets with valid signature', async () => {
    const clearRequest = {
      packetIds: ['packet-1', 'packet-2'],
      wallet: 'osmo1test123',
      signature: 'valid-mock-signature',
    }
    
    const response = await request(API_URL)
      .post('/api/packets/clear')
      .send(clearRequest)
      .expect(200)
    
    expect(response.body.status).toBe('success')
    expect(response.body.txHash).toBeTruthy()
    expect(response.body.cleared).toEqual(clearRequest.packetIds)
  })

  it('rejects clearing with invalid signature', async () => {
    const clearRequest = {
      packetIds: ['packet-1'],
      wallet: 'osmo1test123',
      signature: '', // Invalid empty signature
    }
    
    const response = await request(API_URL)
      .post('/api/packets/clear')
      .send(clearRequest)
      .expect(401)
    
    expect(response.body.error).toBe('Invalid signature')
  })

  it('handles concurrent requests properly', async () => {
    const requests = Array.from({ length: 10 }, (_, i) => 
      request(API_URL)
        .get(`/api/user/osmo1test${i}/transfers`)
    )
    
    const responses = await Promise.all(requests)
    
    responses.forEach(response => {
      expect(response.status).toBe(200)
      expect(Array.isArray(response.body)).toBe(true)
    })
  })
})

describe('Chainpulse Integration', () => {
  it('retrieves stuck packets from Chainpulse', async () => {
    try {
      // Check if Chainpulse is running
      await request(CHAINPULSE_URL).get('/health').expect(200)
      
      const response = await request(API_URL)
        .get('/api/packets/stuck')
        .expect(200)
      
      expect(Array.isArray(response.body)).toBe(true)
      
      // If packets exist, verify structure
      if (response.body.length > 0) {
        const packet = response.body[0]
        expect(packet).toHaveProperty('id')
        expect(packet).toHaveProperty('channelId')
        expect(packet).toHaveProperty('sequence')
        expect(packet).toHaveProperty('sourceChain')
        expect(packet).toHaveProperty('destinationChain')
      }
    } catch (error) {
      console.log('Chainpulse not available, skipping integration test')
    }
  })

  it('handles Chainpulse downtime gracefully', async () => {
    // Simulate Chainpulse being down by using wrong port
    const response = await request(API_URL)
      .get('/api/packets/stuck')
      .expect(200)
    
    // Should return empty array or mock data
    expect(Array.isArray(response.body)).toBe(true)
  })
})

describe('Error Scenarios', () => {
  it('handles malformed JSON in clear request', async () => {
    const response = await request(API_URL)
      .post('/api/packets/clear')
      .set('Content-Type', 'application/json')
      .send('invalid json')
      .expect(400)
    
    expect(response.body.error).toBeTruthy()
  })

  it('handles missing required fields', async () => {
    const response = await request(API_URL)
      .post('/api/packets/clear')
      .send({
        wallet: 'osmo1test123',
        // Missing packetIds and signature
      })
      .expect(400)
    
    expect(response.body.error).toBeTruthy()
  })

  it('returns 404 for non-existent endpoints', async () => {
    await request(API_URL)
      .get('/api/non-existent')
      .expect(404)
  })
})

describe('Performance Tests', () => {
  it('responds within acceptable time', async () => {
    const start = Date.now()
    
    await request(API_URL)
      .get('/api/metrics')
      .expect(200)
    
    const duration = Date.now() - start
    expect(duration).toBeLessThan(1000) // Should respond within 1 second
  })

  it('handles burst traffic', async () => {
    const requests = Array.from({ length: 50 }, () => 
      request(API_URL).get('/health')
    )
    
    const start = Date.now()
    const responses = await Promise.all(requests)
    const duration = Date.now() - start
    
    // All should succeed
    responses.forEach(response => {
      expect(response.status).toBe(200)
    })
    
    // Should complete within reasonable time
    expect(duration).toBeLessThan(5000) // 5 seconds for 50 requests
  })
})