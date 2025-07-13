<template>
  <div class="payment-prompt">
    <div class="space-y-6">
      <!-- Payment details -->
      <div class="bg-gray-50 rounded-lg p-6">
        <h3 class="text-sm font-medium text-gray-700 mb-4">Payment Details</h3>
        
        <div class="space-y-4">
          <!-- Amount -->
          <div>
            <label class="text-sm text-gray-600">Amount</label>
            <div class="flex items-center gap-2 mt-1">
              <p class="font-mono text-lg font-semibold">
                {{ formatAmount() }}
              </p>
              <button
                @click="copyAmount"
                class="text-gray-400 hover:text-gray-600"
                title="Copy amount"
              >
                <CopyIcon class="h-4 w-4" />
              </button>
            </div>
          </div>

          <!-- Recipient -->
          <div>
            <label class="text-sm text-gray-600">Send to</label>
            <div class="flex items-center gap-2 mt-1">
              <p class="font-mono text-sm text-gray-900 break-all">
                {{ paymentAddress }}
              </p>
              <button
                @click="copyAddress"
                class="text-gray-400 hover:text-gray-600 flex-shrink-0"
                title="Copy address"
              >
                <CopyIcon class="h-4 w-4" />
              </button>
            </div>
          </div>

          <!-- Memo -->
          <div>
            <label class="text-sm text-gray-600">
              Memo 
              <span class="text-red-500">*</span>
              <span class="text-xs text-gray-500 ml-1">(Required - must be exact)</span>
            </label>
            <div class="flex items-center gap-2 mt-1">
              <p class="font-mono text-xs text-gray-900 break-all bg-white rounded border border-gray-300 p-2">
                {{ memo }}
              </p>
              <button
                @click="copyMemo"
                class="text-gray-400 hover:text-gray-600 flex-shrink-0"
                title="Copy memo"
              >
                <CopyIcon class="h-4 w-4" />
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Instructions -->
      <div class="bg-blue-50 rounded-lg p-4">
        <div class="flex">
          <InfoIcon class="h-5 w-5 text-blue-400 mt-0.5 flex-shrink-0" />
          <div class="ml-3">
            <p class="text-sm text-blue-800 font-medium mb-1">Important Instructions:</p>
            <ol class="text-sm text-blue-800 space-y-1 list-decimal list-inside">
              <li>Copy all details exactly as shown above</li>
              <li>The memo must be included and exact - it contains your clearing authorization</li>
              <li>Send from the wallet address you connected: {{ formatWalletAddress() }}</li>
              <li>After sending, paste the transaction hash below</li>
            </ol>
          </div>
        </div>
      </div>

      <!-- Transaction hash input -->
      <div>
        <label class="block text-sm font-medium text-gray-700 mb-2">
          Transaction Hash
        </label>
        <div class="flex gap-2">
          <input
            v-model="txHash"
            type="text"
            placeholder="Enter transaction hash after sending payment"
            class="flex-1 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 font-mono text-sm"
          />
          <button
            @click="verifyPayment"
            :disabled="!txHash || verifying"
            class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
          >
            <LoaderIcon v-if="verifying" class="h-4 w-4 animate-spin" />
            <span>{{ verifying ? 'Verifying...' : 'Verify Payment' }}</span>
          </button>
        </div>
        <p v-if="error" class="mt-2 text-sm text-red-600">
          {{ error }}
        </p>
      </div>

      <!-- Timer -->
      <div class="text-center">
        <p class="text-sm text-gray-600">
          Token expires in: 
          <span class="font-mono font-medium">{{ timeRemaining }}</span>
        </p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { Copy as CopyIcon, Info as InfoIcon, Loader as LoaderIcon } from 'lucide-vue-next'
import { clearingService, type ClearingToken } from '@/services/clearing'
import { useWalletStore } from '@/stores/wallet'

const props = defineProps<{
  token: ClearingToken | null
  paymentAddress: string
  memo: string
}>()

const emit = defineEmits<{
  'payment-sent': [txHash: string]
}>()

const walletStore = useWalletStore()

const txHash = ref('')
const verifying = ref(false)
const error = ref('')
const timeRemaining = ref('')

let timerInterval: ReturnType<typeof setInterval> | null = null

const formatAmount = (): string => {
  if (!props.token) return '0'
  return clearingService.formatAmount(props.token.totalRequired, props.token.acceptedDenom)
}

const formatWalletAddress = (): string => {
  const address = walletStore.address || ''
  if (address.length <= 20) return address
  return `${address.slice(0, 10)}...${address.slice(-8)}`
}

const copyAmount = async () => {
  if (!props.token) return
  await navigator.clipboard.writeText(props.token.totalRequired)
  // Show toast notification
}

const copyAddress = async () => {
  await navigator.clipboard.writeText(props.paymentAddress)
  // Show toast notification
}

const copyMemo = async () => {
  await navigator.clipboard.writeText(props.memo)
  // Show toast notification
}

const verifyPayment = async () => {
  if (!txHash.value || !props.token) return
  
  verifying.value = true
  error.value = ''
  
  try {
    const result = await clearingService.verifyPayment(props.token.token, txHash.value)
    
    if (result.verified) {
      emit('payment-sent', txHash.value)
    } else {
      error.value = result.message || 'Payment verification failed'
    }
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Verification failed'
  } finally {
    verifying.value = false
  }
}

const updateTimer = () => {
  if (!props.token) return
  
  const now = Date.now() / 1000
  const remaining = props.token.expiresAt - now
  
  if (remaining <= 0) {
    timeRemaining.value = '00:00'
    if (timerInterval) {
      clearInterval(timerInterval)
    }
    return
  }
  
  const minutes = Math.floor(remaining / 60)
  const seconds = Math.floor(remaining % 60)
  timeRemaining.value = `${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`
}

onMounted(() => {
  updateTimer()
  timerInterval = setInterval(updateTimer, 1000)
})

onUnmounted(() => {
  if (timerInterval) {
    clearInterval(timerInterval)
  }
})
</script>