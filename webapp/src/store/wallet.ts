import { create } from 'zustand'
import { SigningStargateClient } from '@cosmjs/stargate'

interface WalletState {
  isConnected: boolean
  address: string | null
  client: SigningStargateClient | null
  connect: () => Promise<void>
  disconnect: () => void
  signMessage: (message: string) => Promise<string>
}

export const useWalletStore = create<WalletState>((set) => ({
  isConnected: false,
  address: null,
  client: null,

  connect: async () => {
    if (!window.keplr) {
      throw new Error('Keplr wallet not found')
    }

    // Enable Keplr for Osmosis
    await window.keplr.enable('osmosis-1')
    
    const offlineSigner = window.keplr.getOfflineSigner('osmosis-1')
    const accounts = await offlineSigner.getAccounts()
    
    const client = await SigningStargateClient.connectWithSigner(
      'https://rpc.osmosis.zone',
      offlineSigner
    )

    set({
      isConnected: true,
      address: accounts[0].address,
      client,
    })
  },

  disconnect: () => {
    set({
      isConnected: false,
      address: null,
      client: null,
    })
  },

  signMessage: async (message: string) => {
    if (!window.keplr) {
      throw new Error('Keplr wallet not found')
    }
    
    const { address } = useWalletStore.getState()
    if (!address) {
      throw new Error('Wallet not connected')
    }

    // Sign arbitrary message with Keplr
    const signed = await window.keplr.signArbitrary(
      'osmosis-1',
      address,
      message
    )
    
    // Return base64 encoded signature
    return signed.signature
  },
}))