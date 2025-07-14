import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import WalletConnect from '@/components/WalletConnect.vue'
import { useWalletStore } from '@/stores/wallet'

// Mock Keplr
const mockKeplr = {
  enable: vi.fn(),
  getKey: vi.fn(),
  signArbitrary: vi.fn(),
  getOfflineSigner: vi.fn(),
}

// Mock window.keplr
Object.defineProperty(window, 'keplr', {
  value: mockKeplr,
  writable: true,
})

describe('WalletConnect', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('renders connect button when wallet not connected', () => {
    const wrapper = mount(WalletConnect)
    
    expect(wrapper.find('button').text()).toContain('Connect Wallet')
    expect(wrapper.find('[data-testid="wallet-icon"]').exists()).toBe(true)
  })

  it('renders wallet info when connected', async () => {
    const walletStore = useWalletStore()
    walletStore.isConnected = true
    walletStore.address = 'osmo1test123abc'
    walletStore.chainId = 'osmosis-1'
    
    const wrapper = mount(WalletConnect)
    
    expect(wrapper.text()).toContain('osmo1test...test123abc')
    expect(wrapper.find('button').text()).toContain('Disconnect')
  })

  it('shows dropdown menu when clicking connected wallet', async () => {
    const walletStore = useWalletStore()
    walletStore.isConnected = true
    walletStore.address = 'osmo1test123abc'
    
    const wrapper = mount(WalletConnect)
    
    // Initially dropdown is hidden
    expect(wrapper.find('[data-testid="wallet-dropdown"]').exists()).toBe(false)
    
    // Click wallet button
    await wrapper.find('button').trigger('click')
    
    // Dropdown should be visible
    expect(wrapper.find('[data-testid="wallet-dropdown"]').exists()).toBe(true)
    expect(wrapper.text()).toContain('Copy Address')
    expect(wrapper.text()).toContain('View on Explorer')
    expect(wrapper.text()).toContain('Disconnect')
  })

  it('connects wallet when clicking connect button', async () => {
    const walletStore = useWalletStore()
    const connectSpy = vi.spyOn(walletStore, 'connect')
    
    // Mock successful connection
    mockKeplr.enable.mockResolvedValue(undefined)
    mockKeplr.getKey.mockResolvedValue({
      bech32Address: 'osmo1newaddress',
      pubKey: new Uint8Array(),
      address: new Uint8Array(),
      algo: 'secp256k1',
    })
    
    const wrapper = mount(WalletConnect)
    
    await wrapper.find('button').trigger('click')
    
    expect(connectSpy).toHaveBeenCalled()
  })

  it('disconnects wallet when clicking disconnect', async () => {
    const walletStore = useWalletStore()
    walletStore.isConnected = true
    walletStore.address = 'osmo1test123'
    
    const disconnectSpy = vi.spyOn(walletStore, 'disconnect')
    
    const wrapper = mount(WalletConnect)
    
    // Open dropdown
    await wrapper.find('button').trigger('click')
    
    // Click disconnect
    await wrapper.find('[data-testid="disconnect-button"]').trigger('click')
    
    expect(disconnectSpy).toHaveBeenCalled()
  })

  it('copies address to clipboard', async () => {
    const walletStore = useWalletStore()
    walletStore.isConnected = true
    walletStore.address = 'osmo1test123'
    
    // Mock clipboard API
    const writeTextMock = vi.fn()
    Object.assign(navigator, {
      clipboard: {
        writeText: writeTextMock,
      },
    })
    
    const wrapper = mount(WalletConnect)
    
    // Open dropdown
    await wrapper.find('button').trigger('click')
    
    // Click copy
    await wrapper.find('[data-testid="copy-address-button"]').trigger('click')
    
    expect(writeTextMock).toHaveBeenCalledWith('osmo1test123')
    expect(wrapper.text()).toContain('Copied!')
  })

  it('opens explorer in new tab', async () => {
    const walletStore = useWalletStore()
    walletStore.isConnected = true
    walletStore.address = 'osmo1test123'
    walletStore.chainId = 'osmosis-1'
    
    // Mock window.open
    const openMock = vi.fn()
    window.open = openMock
    
    const wrapper = mount(WalletConnect)
    
    // Open dropdown
    await wrapper.find('button').trigger('click')
    
    // Click explorer link
    await wrapper.find('[data-testid="explorer-link"]').trigger('click')
    
    expect(openMock).toHaveBeenCalledWith(
      'https://www.mintscan.io/osmosis/address/osmo1test123',
      '_blank',
      'noopener,noreferrer'
    )
  })

  it('shows error when Keplr not installed', async () => {
    // Remove Keplr
    delete (window as any).keplr
    
    const wrapper = mount(WalletConnect)
    
    await wrapper.find('button').trigger('click')
    
    expect(wrapper.emitted('error')).toBeTruthy()
    expect(wrapper.emitted('error')?.[0]).toEqual([
      { message: 'Please install Keplr wallet extension' }
    ])
    
    // Restore mock
    (window as any).keplr = mockKeplr
  })

  it('handles connection errors gracefully', async () => {
    const walletStore = useWalletStore()
    const connectSpy = vi.spyOn(walletStore, 'connect').mockRejectedValue(
      new Error('User rejected')
    )
    
    const wrapper = mount(WalletConnect)
    
    await wrapper.find('button').trigger('click')
    
    expect(connectSpy).toHaveBeenCalled()
    expect(wrapper.emitted('error')).toBeTruthy()
  })

  it('closes dropdown when clicking outside', async () => {
    const walletStore = useWalletStore()
    walletStore.isConnected = true
    walletStore.address = 'osmo1test123'
    
    const wrapper = mount(WalletConnect, {
      attachTo: document.body,
    })
    
    // Open dropdown
    await wrapper.find('button').trigger('click')
    expect(wrapper.find('[data-testid="wallet-dropdown"]').exists()).toBe(true)
    
    // Click outside
    document.body.click()
    await wrapper.vm.$nextTick()
    
    expect(wrapper.find('[data-testid="wallet-dropdown"]').exists()).toBe(false)
    
    wrapper.unmount()
  })
})