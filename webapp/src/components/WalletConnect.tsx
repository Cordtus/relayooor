import { useState } from 'react'
import { useWalletStore } from '@/store/wallet'
import toast from 'react-hot-toast'

export default function WalletConnect() {
  const { isConnected, address, connect, disconnect } = useWalletStore()
  const [isConnecting, setIsConnecting] = useState(false)

  const handleConnect = async () => {
    setIsConnecting(true)
    try {
      await connect()
      toast.success('Wallet connected')
    } catch (error) {
      toast.error('Failed to connect wallet')
      console.error(error)
    } finally {
      setIsConnecting(false)
    }
  }

  const handleDisconnect = () => {
    disconnect()
    toast.success('Wallet disconnected')
  }

  if (isConnected && address) {
    return (
      <div className="flex items-center space-x-3">
        <span className="text-sm text-gray-700">
          {address.slice(0, 8)}...{address.slice(-6)}
        </span>
        <button
          onClick={handleDisconnect}
          className="text-sm bg-gray-200 hover:bg-gray-300 px-3 py-1.5 rounded-md"
        >
          Disconnect
        </button>
      </div>
    )
  }

  return (
    <button
      onClick={handleConnect}
      disabled={isConnecting}
      className="bg-primary-600 hover:bg-primary-700 text-white px-4 py-2 rounded-md text-sm font-medium disabled:opacity-50"
    >
      {isConnecting ? 'Connecting...' : 'Connect Wallet'}
    </button>
  )
}