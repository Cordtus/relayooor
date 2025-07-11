import { useState } from 'react'
import { useQuery, useMutation } from '@tanstack/react-query'
import { useWalletStore } from '@/store/wallet'
import { fetchUserTransfers, clearPackets } from '@/services/api'
import toast from 'react-hot-toast'
import { formatDistance } from 'date-fns'

export default function PacketClearing() {
  const { isConnected, address, signMessage } = useWalletStore()
  const [selectedPackets, setSelectedPackets] = useState<string[]>([])
  const [showAllTransfers, setShowAllTransfers] = useState(false)

  const { data: transfers, isLoading, refetch } = useQuery({
    queryKey: ['user-transfers', address, showAllTransfers],
    queryFn: () => fetchUserTransfers(address!),
    enabled: isConnected && !!address,
  })

  // Filter transfers based on view mode
  const displayedTransfers = showAllTransfers 
    ? transfers 
    : transfers?.filter(t => t.status === 'stuck')

  const clearMutation = useMutation({
    mutationFn: async (packetIds: string[]) => {
      // Sign message for authentication
      const message = `Clear packets: ${packetIds.join(',')}`
      const signature = await signMessage(message)
      return clearPackets({
        packetIds,
        wallet: address!,
        signature,
      })
    },
    onSuccess: (data) => {
      toast.success(`Successfully cleared ${data.cleared.length} transfers`)
      setSelectedPackets([])
      refetch()
    },
    onError: (error: any) => {
      toast.error(error.message || 'Failed to clear transfers')
    },
  })

  const handleSelectAll = () => {
    const stuckTransfers = displayedTransfers?.filter(t => t.status === 'stuck') || []
    if (selectedPackets.length === stuckTransfers.length) {
      setSelectedPackets([])
    } else {
      setSelectedPackets(stuckTransfers.map(t => t.id))
    }
  }

  const handleClearSelected = async () => {
    if (selectedPackets.length === 0) {
      toast.error('No transfers selected')
      return
    }
    try {
      await clearMutation.mutateAsync(selectedPackets)
    } catch (error) {
      console.error('Clear failed:', error)
    }
  }

  const handleClearTransfer = async (transferId: string) => {
    try {
      await clearMutation.mutateAsync([transferId])
    } catch (error) {
      console.error('Clear failed:', error)
    }
  }

  const formatAmount = (amount: string, denom: string) => {
    // Convert from smallest unit to display unit
    const displayAmount = (parseInt(amount) / 1000000).toFixed(2)
    const displayDenom = denom.replace('u', '').toUpperCase()
    return `${displayAmount} ${displayDenom}`
  }

  if (!isConnected) {
    return (
      <div className="text-center py-12">
        <div className="mx-auto w-24 h-24 mb-4">
          <svg className="w-full h-full text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z" />
          </svg>
        </div>
        <h3 className="text-lg font-medium text-gray-900 mb-2">Connect Your Wallet</h3>
        <p className="text-gray-500 mb-4">Connect your wallet to view and clear stuck IBC transfers</p>
        <button className="bg-primary-600 hover:bg-primary-700 text-white px-6 py-2 rounded-md font-medium">
          Connect Wallet
        </button>
      </div>
    )
  }

  const stuckCount = transfers?.filter(t => t.status === 'stuck').length || 0

  return (
    <div className="space-y-6">
      <div className="flex justify-between items-center">
        <div>
          <h2 className="text-2xl font-bold text-gray-900">My IBC Transfers</h2>
          {stuckCount > 0 && (
            <p className="text-sm text-red-600 mt-1">
              You have {stuckCount} stuck transfer{stuckCount > 1 ? 's' : ''}
            </p>
          )}
        </div>
        <div className="flex items-center space-x-4">
          <label className="flex items-center">
            <input
              type="checkbox"
              checked={showAllTransfers}
              onChange={(e) => setShowAllTransfers(e.target.checked)}
              className="rounded border-gray-300 text-primary-600 focus:ring-primary-500"
            />
            <span className="ml-2 text-sm text-gray-700">Show all transfers</span>
          </label>
          <button
            onClick={handleClearSelected}
            disabled={selectedPackets.length === 0 || clearMutation.isPending}
            className="bg-red-600 hover:bg-red-700 text-white px-4 py-2 rounded-md text-sm font-medium disabled:opacity-50 flex items-center space-x-2"
          >
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
            </svg>
            <span>Clear Selected ({selectedPackets.length})</span>
          </button>
        </div>
      </div>

      {isLoading ? (
        <div className="animate-pulse text-center py-8">
          <div className="inline-flex items-center">
            <svg className="animate-spin h-5 w-5 mr-3" viewBox="0 0 24 24">
              <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
              <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
            </svg>
            Loading your transfers...
          </div>
        </div>
      ) : displayedTransfers?.length === 0 ? (
        <div className="text-center py-12 bg-white rounded-lg shadow">
          <svg className="mx-auto h-12 w-12 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
          </svg>
          <h3 className="mt-2 text-sm font-medium text-gray-900">
            {showAllTransfers ? 'No transfers found' : 'No stuck transfers'}
          </h3>
          <p className="mt-1 text-sm text-gray-500">
            {showAllTransfers ? 'You have no IBC transfers' : 'All your transfers have been completed successfully'}
          </p>
        </div>
      ) : (
        <div className="bg-white shadow overflow-hidden sm:rounded-md">
          {!showAllTransfers && stuckCount > 0 && (
            <div className="px-4 py-3 border-b border-gray-200 bg-gray-50">
              <label className="flex items-center">
                <input
                  type="checkbox"
                  checked={selectedPackets.length === stuckCount && stuckCount > 0}
                  onChange={handleSelectAll}
                  className="rounded border-gray-300 text-primary-600 focus:ring-primary-500"
                />
                <span className="ml-2 text-sm text-gray-700">Select all stuck transfers</span>
              </label>
            </div>
          )}
          <ul className="divide-y divide-gray-200">
            {displayedTransfers?.map((transfer) => (
              <li key={transfer.id} className="hover:bg-gray-50">
                <div className="px-4 py-4 sm:px-6">
                  <div className="flex items-center justify-between">
                    <div className="flex items-center flex-1">
                      {transfer.status === 'stuck' && (
                        <input
                          type="checkbox"
                          checked={selectedPackets.includes(transfer.id)}
                          onChange={(e) => {
                            if (e.target.checked) {
                              setSelectedPackets([...selectedPackets, transfer.id])
                            } else {
                              setSelectedPackets(selectedPackets.filter(id => id !== transfer.id))
                            }
                          }}
                          className="rounded border-gray-300 text-primary-600 focus:ring-primary-500"
                        />
                      )}
                      <div className={transfer.status === 'stuck' ? 'ml-4' : ''}>
                        <div className="flex items-center space-x-2">
                          <span className="text-sm font-medium text-gray-900">
                            {formatAmount(transfer.amount, transfer.denom)}
                          </span>
                          <span className="text-gray-400">•</span>
                          <span className="text-sm text-gray-600">
                            {transfer.sourceChain} → {transfer.destinationChain}
                          </span>
                          <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                            transfer.status === 'stuck' ? 'bg-red-100 text-red-800' :
                            transfer.status === 'completed' ? 'bg-green-100 text-green-800' :
                            'bg-yellow-100 text-yellow-800'
                          }`}>
                            {transfer.status}
                          </span>
                        </div>
                        <div className="mt-1 text-sm text-gray-500">
                          <span>Channel: {transfer.channelId}</span>
                          <span className="mx-2">•</span>
                          <span>Sent {formatDistance(new Date(transfer.timestamp), new Date(), { addSuffix: true })}</span>
                          {transfer.stuckDuration && (
                            <>
                              <span className="mx-2">•</span>
                              <span className="text-red-600">Stuck for {transfer.stuckDuration}</span>
                            </>
                          )}
                        </div>
                        <div className="mt-1 text-xs text-gray-400">
                          Tx: {transfer.txHash}
                        </div>
                      </div>
                    </div>
                    {transfer.status === 'stuck' && (
                      <button
                        onClick={() => handleClearTransfer(transfer.id)}
                        disabled={clearMutation.isPending}
                        className="ml-4 inline-flex items-center px-3 py-1 border border-transparent text-sm leading-5 font-medium rounded-md text-white bg-red-600 hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500 disabled:opacity-50"
                      >
                        Clear Transfer
                      </button>
                    )}
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