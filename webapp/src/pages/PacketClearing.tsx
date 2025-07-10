import { useState } from 'react'
import { useQuery, useMutation } from '@tanstack/react-query'
import { useWalletStore } from '@/store/wallet'
import { fetchStuckPackets, clearPackets } from '@/services/api'
import toast from 'react-hot-toast'

export default function PacketClearing() {
  const { isConnected, address } = useWalletStore()
  const [selectedPackets, setSelectedPackets] = useState<string[]>([])
  const [filterByWallet, setFilterByWallet] = useState(false)

  const { data: packets, isLoading, refetch } = useQuery({
    queryKey: ['stuck-packets', filterByWallet ? address : null],
    queryFn: () => fetchStuckPackets(filterByWallet ? address : undefined),
    enabled: isConnected || !filterByWallet,
  })

  const clearMutation = useMutation({
    mutationFn: clearPackets,
    onSuccess: () => {
      toast.success('Packets cleared successfully')
      setSelectedPackets([])
      refetch()
    },
    onError: () => {
      toast.error('Failed to clear packets')
    },
  })

  const handleSelectAll = () => {
    if (selectedPackets.length === packets?.length) {
      setSelectedPackets([])
    } else {
      setSelectedPackets(packets?.map(p => p.id) || [])
    }
  }

  const handleClearSelected = () => {
    if (selectedPackets.length === 0) {
      toast.error('No packets selected')
      return
    }
    clearMutation.mutate(selectedPackets)
  }

  const handleClearChannel = (channelId: string) => {
    const channelPackets = packets?.filter(p => p.channelId === channelId).map(p => p.id) || []
    if (channelPackets.length === 0) {
      toast.error('No packets in this channel')
      return
    }
    clearMutation.mutate(channelPackets)
  }

  if (!isConnected) {
    return (
      <div className="text-center py-12">
        <h3 className="text-lg font-medium text-gray-900 mb-2">Connect Wallet</h3>
        <p className="text-gray-500">Connect your wallet to clear stuck packets</p>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <h2 className="text-2xl font-bold text-gray-900">Packet Clearing</h2>
        <div className="flex items-center space-x-4">
          <label className="flex items-center">
            <input
              type="checkbox"
              checked={filterByWallet}
              onChange={(e) => setFilterByWallet(e.target.checked)}
              className="rounded border-gray-300 text-primary-600 focus:ring-primary-500"
            />
            <span className="ml-2 text-sm text-gray-700">Show only my packets</span>
          </label>
          <button
            onClick={handleClearSelected}
            disabled={selectedPackets.length === 0 || clearMutation.isPending}
            className="bg-primary-600 hover:bg-primary-700 text-white px-4 py-2 rounded-md text-sm font-medium disabled:opacity-50"
          >
            Clear Selected ({selectedPackets.length})
          </button>
        </div>
      </div>

      {isLoading ? (
        <div className="animate-pulse">Loading stuck packets...</div>
      ) : (
        <div className="bg-white shadow overflow-hidden sm:rounded-md">
          <div className="px-4 py-3 border-b border-gray-200 bg-gray-50">
            <label className="flex items-center">
              <input
                type="checkbox"
                checked={selectedPackets.length === packets?.length && packets?.length > 0}
                onChange={handleSelectAll}
                className="rounded border-gray-300 text-primary-600 focus:ring-primary-500"
              />
              <span className="ml-2 text-sm text-gray-700">Select all</span>
            </label>
          </div>
          <ul className="divide-y divide-gray-200">
            {packets?.map((packet) => (
              <li key={packet.id} className="hover:bg-gray-50">
                <div className="px-4 py-4 sm:px-6">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center">
                      <input
                        type="checkbox"
                        checked={selectedPackets.includes(packet.id)}
                        onChange={(e) => {
                          if (e.target.checked) {
                            setSelectedPackets([...selectedPackets, packet.id])
                          } else {
                            setSelectedPackets(selectedPackets.filter(id => id !== packet.id))
                          }
                        }}
                        className="rounded border-gray-300 text-primary-600 focus:ring-primary-500"
                      />
                      <div className="ml-4">
                        <div className="text-sm font-medium text-gray-900">
                          {packet.sourceChain} → {packet.destinationChain}
                        </div>
                        <div className="text-sm text-gray-500">
                          Channel: {packet.channelId} | Sequence: {packet.sequence}
                        </div>
                        <div className="text-sm text-gray-500">
                          Stuck for: {packet.stuckDuration}
                        </div>
                      </div>
                    </div>
                    <button
                      onClick={() => handleClearChannel(packet.channelId)}
                      className="text-sm text-primary-600 hover:text-primary-900"
                    >
                      Clear Channel
                    </button>
                  </div>
                </div>
              </li>
            ))}
          </ul>
        </div>
      )}
    </div>
  )
}