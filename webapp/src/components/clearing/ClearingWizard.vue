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
        <div v-if="loading" class="text-center py-8">
          <div class="inline-flex items-center">
            <svg class="animate-spin h-5 w-5 mr-3 text-blue-500" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
              <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
              <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            Loading stuck packets...
          </div>
        </div>
        <PacketSelector 
          v-else
          v-model:selected="selectedPackets"
          :stuck-packets="stuckPackets"
          @update:selected="handlePacketSelection"
        />
        <div class="mt-6 flex justify-end">
          <Button
            @click="proceedToFees"
            :disabled="selectedPackets.length === 0 || loading"
            variant="primary"
          >
            Continue
          </Button>
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
          <Button
            @click="currentStep--"
            variant="secondary"
          >
            Back
          </Button>
          <Button
            @click="proceedToPayment"
            variant="primary"
          >
            Continue to Payment
          </Button>
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
          <Button
            @click="currentStep--"
            variant="secondary"
          >
            Back
          </Button>
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
          <Button
            @click="startNew"
            variant="primary"
            class="mt-6"
          >
            Clear More Packets
          </Button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { CheckIcon, CheckCircleIcon } from 'lucide-vue-next'
import { clearingService, type ClearingToken, type ClearingStatus, type TokenResponse } from '@/services/clearing'
import { packetsService, type UserTransfer, type StuckPacket } from '@/services/packets'
import { useWalletStore } from '@/stores/wallet'
import { configService } from '@/services/config'
import PacketSelector from './PacketSelector.vue'
import FeeEstimator from './FeeEstimator.vue'
import PaymentPrompt from './PaymentPrompt.vue'
import ClearingProgress from './ClearingProgress.vue'
import Button from '@/components/ui/Button.vue'

// Helper function to infer chain from wallet address prefix
async function inferChainFromAddress(address: string): Promise<string | null> {
  if (!address) return null
  
  // Extract the prefix (everything before the '1')
  const match = address.match(/^([a-z]+)1/)
  if (!match) return null
  
  const prefix = match[1]
  const chain = await configService.getChainByPrefix(prefix)
  return chain?.chain_id || null
}

// Helper to parse stuck duration
function parseStuckDuration(duration: string | undefined): number {
  if (!duration) return 0
  const match = duration.match(/(\d+)([hms])/)
  if (!match) return 0
  const [, num, unit] = match
  const value = parseInt(num)
  switch (unit) {
    case 'h': return value * 3600
    case 'm': return value * 60
    case 's': return value
    default: return 0
  }
}

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
const stuckPackets = ref<any[]>([])
const loading = ref(true)

// Load stuck packets for user
onMounted(async () => {
  if (walletStore.isConnected && walletStore.address) {
    try {
      loading.value = true
      const packets = await packetsService.getUserStuckPackets(walletStore.address)
      // Map UserTransfer to StuckPacket format
      stuckPackets.value = packets
        .filter(p => p.status === 'stuck')
        .map(p => ({
          id: p.id,
          chain: inferChainFromAddress(p.sender) || p.sourceChain,
          channel: p.channelId,
          sequence: p.sequence,
          sender: p.sender,
          receiver: p.receiver,
          amount: p.amount,
          denom: p.denom,
          age: parseStuckDuration(p.stuckDuration),
          attempts: 0
        }))
    } catch (error) {
      console.error('Failed to load stuck packets:', error)
      // Clear packets on error instead of using mock data
      stuckPackets.value = []
      // Optionally show error to user
      if (error instanceof Error) {
        console.error('Error loading packets:', error.message)
      }
    } finally {
      loading.value = false
    }
  }
})

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
          chain: p.sourceChain,
          channel: p.channelId,
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
  const chain = selectedPackets.value[0]?.sourceChain || 'osmosis-1'
  return configService.getExplorerUrl(chain, txHash)
}

</script>

<style scoped>
.clearing-wizard {
  max-width: 48rem;
  margin: 0 auto;
}
</style>