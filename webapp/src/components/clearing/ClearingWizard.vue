<template>
  <div class="clearing-wizard">
    <!-- Progress indicator -->
    <div class="mb-8">
      <div class="flex items-center justify-between">
        <div 
          v-for="(step, index) in steps" 
          :key="step.id"
          class="flex items-center"
        >
          <div 
            :class="[
              'w-10 h-10 rounded-full flex items-center justify-center font-medium',
              currentStep > index ? 'bg-green-500 text-white' :
              currentStep === index ? 'bg-blue-500 text-white' :
              'bg-gray-200 text-gray-600'
            ]"
          >
            <CheckIcon v-if="currentStep > index" class="w-5 h-5" />
            <span v-else>{{ index + 1 }}</span>
          </div>
          <div 
            v-if="index < steps.length - 1"
            :class="[
              'h-1 w-24 mx-2',
              currentStep > index ? 'bg-green-500' : 'bg-gray-200'
            ]"
          />
        </div>
      </div>
      <div class="flex items-center justify-between mt-2">
        <div 
          v-for="step in steps" 
          :key="step.id"
          class="text-xs text-gray-600 w-10 text-center first:text-left last:text-right"
        >
          {{ step.name }}
        </div>
      </div>
    </div>

    <!-- Step content -->
    <div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
      <!-- Step 1: Select packets -->
      <div v-if="currentStep === 0">
        <h2 class="text-xl font-semibold mb-4">Select Packets to Clear</h2>
        <PacketSelector 
          v-model:selected="selectedPackets"
          :stuck-packets="stuckPackets"
          @update:selected="handlePacketSelection"
        />
        <div class="mt-6 flex justify-end">
          <button
            @click="proceedToFees"
            :disabled="selectedPackets.length === 0"
            class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            Continue
          </button>
        </div>
      </div>

      <!-- Step 2: Review fees -->
      <div v-if="currentStep === 1">
        <h2 class="text-xl font-semibold mb-4">Review Fees</h2>
        <FeeEstimator 
          :token="clearingToken"
          :packet-count="selectedPackets.length"
        />
        <div class="mt-6 flex justify-between">
          <button
            @click="currentStep--"
            class="px-4 py-2 border border-gray-300 text-gray-700 rounded-md hover:bg-gray-50"
          >
            Back
          </button>
          <button
            @click="proceedToPayment"
            class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
          >
            Continue to Payment
          </button>
        </div>
      </div>

      <!-- Step 3: Make payment -->
      <div v-if="currentStep === 2">
        <h2 class="text-xl font-semibold mb-4">Make Payment</h2>
        <PaymentPrompt 
          :token="clearingToken"
          :payment-address="paymentAddress"
          :memo="paymentMemo"
          @payment-sent="handlePaymentSent"
        />
        <div class="mt-6 flex justify-between">
          <button
            @click="currentStep--"
            class="px-4 py-2 border border-gray-300 text-gray-700 rounded-md hover:bg-gray-50"
          >
            Back
          </button>
        </div>
      </div>

      <!-- Step 4: Clearing progress -->
      <div v-if="currentStep === 3">
        <h2 class="text-xl font-semibold mb-4">Clearing in Progress</h2>
        <ClearingProgress 
          :token="clearingToken?.token"
          :status="clearingStatus"
          @complete="handleComplete"
        />
      </div>

      <!-- Step 5: Complete -->
      <div v-if="currentStep === 4">
        <h2 class="text-xl font-semibold mb-4">Clearing Complete!</h2>
        <div class="text-center py-8">
          <CheckCircleIcon class="w-16 h-16 text-green-500 mx-auto mb-4" />
          <p class="text-lg mb-2">
            Successfully cleared {{ clearingStatus?.execution?.packetsCleared || 0 }} packets
          </p>
          <div v-if="clearingStatus?.execution?.txHashes?.length" class="mt-4">
            <p class="text-sm text-gray-600 mb-2">Transaction Hashes:</p>
            <div v-for="hash in clearingStatus.execution.txHashes" :key="hash" class="mb-1">
              <a 
                :href="getExplorerUrl(hash)"
                target="_blank"
                rel="noopener noreferrer"
                class="text-blue-600 hover:underline text-sm font-mono"
              >
                {{ hash.slice(0, 10) }}...{{ hash.slice(-10) }}
              </a>
            </div>
          </div>
          <button
            @click="startNew"
            class="mt-6 px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700"
          >
            Clear More Packets
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { CheckIcon, CheckCircleIcon } from 'lucide-vue-next'
import { clearingService, type ClearingToken, type ClearingStatus } from '@/services/clearing'
import { useWalletStore } from '@/stores/wallet'
import PacketSelector from './PacketSelector.vue'
import FeeEstimator from './FeeEstimator.vue'
import PaymentPrompt from './PaymentPrompt.vue'
import ClearingProgress from './ClearingProgress.vue'

const walletStore = useWalletStore()

const steps = [
  { id: 'select', name: 'Select' },
  { id: 'fees', name: 'Fees' },
  { id: 'payment', name: 'Payment' },
  { id: 'clearing', name: 'Clearing' },
  { id: 'complete', name: 'Complete' }
]

const currentStep = ref(0)
const selectedPackets = ref<any[]>([])
const clearingToken = ref<ClearingToken | null>(null)
const paymentAddress = ref('')
const paymentMemo = ref('')
const clearingStatus = ref<ClearingStatus | null>(null)

// Mock stuck packets for development
const stuckPackets = ref([
  {
    id: '1',
    chain: 'osmosis-1',
    channel: 'channel-0',
    sequence: 12345,
    sender: walletStore.address,
    receiver: 'cosmos1xyz...',
    amount: '1000000',
    denom: 'uosmo',
    age: 1820,
    attempts: 3
  },
  {
    id: '2',
    chain: 'osmosis-1',
    channel: 'channel-0',
    sequence: 12346,
    sender: walletStore.address,
    receiver: 'cosmos1abc...',
    amount: '500000',
    denom: 'uosmo',
    age: 920,
    attempts: 1
  }
])

const handlePacketSelection = (packets: any[]) => {
  selectedPackets.value = packets
}

const proceedToFees = async () => {
  try {
    // Request clearing token
    const response = await clearingService.requestToken({
      walletAddress: walletStore.address!,
      chainId: walletStore.chainId!,
      type: 'packet',
      targets: {
        packets: selectedPackets.value.map(p => ({
          chain: p.chain,
          channel: p.channel,
          sequence: p.sequence
        }))
      }
    })
    
    clearingToken.value = response.token
    paymentAddress.value = response.paymentAddress
    paymentMemo.value = response.memo
    
    currentStep.value++
  } catch (error) {
    console.error('Failed to request token:', error)
    // Show error toast
  }
}

const proceedToPayment = () => {
  currentStep.value++
}

const handlePaymentSent = async (txHash: string) => {
  try {
    // Verify payment
    await clearingService.verifyPayment(clearingToken.value!.token, txHash)
    
    // Move to clearing step
    currentStep.value++
    
    // Subscribe to status updates
    const unsubscribe = clearingService.subscribeToUpdates(
      clearingToken.value!.token,
      (status) => {
        clearingStatus.value = status
      }
    )
    
    // Cleanup on unmount
    onUnmounted(unsubscribe)
  } catch (error) {
    console.error('Failed to verify payment:', error)
    // Show error toast
  }
}

const handleComplete = () => {
  currentStep.value++
}

const startNew = () => {
  // Reset wizard
  currentStep.value = 0
  selectedPackets.value = []
  clearingToken.value = null
  clearingStatus.value = null
}

const getExplorerUrl = (txHash: string): string => {
  // Get explorer URL based on chain
  const explorers: Record<string, string> = {
    'osmosis-1': 'https://www.mintscan.io/osmosis/tx/',
    'cosmoshub-4': 'https://www.mintscan.io/cosmos/tx/',
  }
  
  const chain = selectedPackets.value[0]?.chain || 'osmosis-1'
  const baseUrl = explorers[chain] || explorers['osmosis-1']
  
  return `${baseUrl}${txHash}`
}

// Import lifecycle hook
import { onUnmounted } from 'vue'
</script>

<style scoped>
.clearing-wizard {
  max-width: 48rem;
  margin: 0 auto;
}
</style>