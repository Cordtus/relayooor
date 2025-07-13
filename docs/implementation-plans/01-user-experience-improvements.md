# User Experience Improvements Implementation Plan

## 1. Direct Wallet Payment Integration

### Current State
- Users manually copy payment details and send transaction
- Manual transaction hash entry required

### Proposed Implementation
```typescript
// webapp/src/services/wallet.ts
interface DirectPaymentRequest {
  chainId: string;
  fromAddress: string;
  toAddress: string;
  amount: Coin[];
  memo: string;
}

async function sendDirectPayment(request: DirectPaymentRequest): Promise<string> {
  if (!window.keplr) throw new Error('Keplr not available');
  
  // Add timeout for wallet response
  const timeoutPromise = new Promise<never>((_, reject) => {
    setTimeout(() => reject(new Error('Wallet response timeout')), 30000);
  });
  
  try {
    const offlineSigner = window.keplr.getOfflineSigner(request.chainId);
    const signingClient = await SigningStargateClient.connectWithSigner(
      getRpcEndpoint(request.chainId),
      offlineSigner
    );
    
    // Simulate transaction first
    const simulation = await signingClient.simulate(
      request.fromAddress,
      [{
        typeUrl: '/cosmos.bank.v1beta1.MsgSend',
        value: {
          fromAddress: request.fromAddress,
          toAddress: request.toAddress,
          amount: request.amount
        }
      }],
      request.memo
    );
    
    // Execute with gas buffer
    const result = await Promise.race([
      signingClient.sendTokens(
        request.fromAddress,
        request.toAddress,
        request.amount,
        {
          amount: [{ denom: request.amount[0].denom, amount: '0' }],
          gas: String(Math.ceil(simulation * 1.2))
        },
        request.memo
      ),
      timeoutPromise
    ]);
    
    return result.transactionHash;
  } catch (error) {
    if (error.message.includes('Request rejected')) {
      throw new Error('Transaction rejected by user');
    }
    throw error;
  }
}
```

### Component Updates
```vue
<!-- webapp/src/components/clearing/PaymentPrompt.vue -->
<template>
  <div class="payment-options">
    <!-- Direct Payment Option (Primary) -->
    <button @click="payWithWallet" class="primary-payment-btn">
      Pay with Wallet
    </button>
    
    <!-- QR Code Option (Secondary) -->
    <button @click="showQRCode" class="secondary-payment-btn">
      Pay with Mobile
    </button>
    
    <!-- Manual Payment (Tertiary) -->
    <details class="manual-payment-details">
      <summary>Pay Manually</summary>
      <!-- Existing manual payment UI -->
    </details>
  </div>
</template>
```

## 2. QR Code for Mobile Payments

### Implementation
```typescript
// webapp/src/utils/qrcode.ts
import QRCode from 'qrcode';

interface CosmosPaymentURI {
  address: string;
  amount: string;
  denom: string;
  memo: string;
}

export function generateCosmosURI(payment: CosmosPaymentURI): string {
  // cosmos:<address>?amount=<amount>&denom=<denom>&memo=<memo>
  const params = new URLSearchParams({
    amount: payment.amount,
    denom: payment.denom,
    memo: payment.memo
  });
  return `cosmos:${payment.address}?${params.toString()}`;
}

export async function generateQRCode(uri: string): Promise<string> {
  // Check memo length for QR capacity
  if (uri.length > 1000) {
    console.warn('URI may be too long for reliable QR scanning');
  }
  
  return QRCode.toDataURL(uri, {
    width: 256,
    margin: 2,
    errorCorrectionLevel: 'H', // High error correction
    color: {
      dark: '#000000',
      light: '#FFFFFF'
    }
  });
}

export function validateMemo(memo: string): { valid: boolean; error?: string } {
  if (!memo) {
    return { valid: false, error: 'Memo is required' };
  }
  
  if (memo.length > 200) {
    return { valid: false, error: 'Memo too long (max 200 characters)' };
  }
  
  // Check for required token pattern
  const tokenPattern = /^CLR-[A-Za-z0-9]{8}-[A-Za-z0-9]{4}-[A-Za-z0-9]{4}-[A-Za-z0-9]{4}-[A-Za-z0-9]{12}$/;
  if (!tokenPattern.test(memo)) {
    return { valid: false, error: 'Invalid memo format' };
  }
  
  return { valid: true };
}
```

### Component Integration
```vue
<!-- webapp/src/components/clearing/QRCodePayment.vue -->
<template>
  <div class="qr-payment-modal">
    <h3>Scan with Mobile Wallet</h3>
    <img :src="qrCodeDataUrl" alt="Payment QR Code" />
    <p class="payment-amount">{{ formatAmount(amount, denom) }}</p>
    <p class="payment-memo">Memo: {{ truncateMemo(memo) }}</p>
    <button @click="copyURI" class="copy-uri-btn">
      Copy Payment Link
    </button>
  </div>
</template>
```

## 3. Simple Mode for Non-Technical Users

### Implementation
```vue
<!-- webapp/src/components/clearing/SimpleMode.vue -->
<template>
  <div class="simple-mode-container">
    <div class="stuck-summary">
      <h2>You have {{ stuckCount }} stuck transfers</h2>
      <div v-if="multipleChains" class="chain-breakdown">
        <div v-for="chain in chainBreakdown" :key="chain.id" class="chain-item">
          <span>{{ chain.name }}: {{ formatAmount(chain.total, chain.denom) }}</span>
        </div>
      </div>
      <p v-else class="total-value">
        Worth {{ formatAmount(totalAmount, primaryDenom) }}
        <span v-if="usdAvailable"> (~{{ formatUSD(totalValueUSD) }})</span>
      </p>
    </div>
    
    <button @click="clearAllStuckPackets" class="big-action-button">
      <span class="action-text">Clear All Transfers</span>
      <span class="fee-text">for {{ formatMultiChainFees(fees) }}</span>
      <span v-if="savingsAvailable" class="savings-text">
        Save ~{{ formatUSD(savedFees) }} vs manual retry
      </span>
    </button>
    
    <p class="simple-explanation">
      We'll automatically clear all your stuck transfers. 
      The fee covers our service and network costs.
    </p>
    
    <a href="#" @click.prevent="switchToAdvanced" class="advanced-link">
      Need more control? Use advanced mode
    </a>
  </div>
</template>
```

### Store Integration
```typescript
// webapp/src/stores/clearing.ts
interface ClearingMode {
  mode: 'simple' | 'advanced';
  autoSelectAll: boolean;
}

const clearingMode = ref<ClearingMode>({
  mode: 'simple',
  autoSelectAll: true
});

// Persist user preference
watch(clearingMode, (newMode) => {
  localStorage.setItem('clearing-mode', JSON.stringify(newMode));
});
```

## 4. Improved Error Messages

### Error Message System
```typescript
// webapp/src/utils/errors.ts
export const ERROR_MESSAGES: Record<string, ErrorInfo> = {
  INSUFFICIENT_BALANCE: {
    title: "Insufficient Balance",
    message: "You need at least {amount} {denom} to complete this transaction",
    action: "Add funds to your wallet and try again",
    icon: "wallet-alert"
  },
  TIMEOUT_EXPIRED: {
    title: "Request Expired",
    message: "Your clearing request has expired for security reasons",
    action: "Start a new clearing request",
    icon: "clock-alert"
  },
  CHANNEL_CLOSED: {
    title: "Channel Unavailable",
    message: "The IBC channel for this transfer is temporarily closed",
    action: "Try again later or contact support",
    icon: "channel-alert"
  },
  PAYMENT_MISMATCH: {
    title: "Payment Amount Incorrect",
    message: "The payment amount doesn't match the required fee",
    action: "Send exactly {required} {denom}",
    icon: "amount-alert"
  }
};

export function formatErrorMessage(code: string, context?: Record<string, any>): ErrorInfo {
  const template = ERROR_MESSAGES[code] || ERROR_MESSAGES.UNKNOWN;
  return {
    ...template,
    message: interpolate(template.message, context),
    action: interpolate(template.action, context)
  };
}
```

### Error Display Component
```vue
<!-- webapp/src/components/ui/ErrorAlert.vue -->
<template>
  <div class="error-alert" :class="errorClass">
    <Icon :name="error.icon" class="error-icon" />
    <div class="error-content">
      <h4>{{ error.title }}</h4>
      <p>{{ error.message }}</p>
      <p class="error-action">
        <strong>What to do:</strong> {{ error.action }}
      </p>
    </div>
    <button @click="$emit('dismiss')" class="dismiss-btn">
      <Icon name="x" />
    </button>
  </div>
</template>
```

## 5. Human-Readable Fee Display

### Fee Formatting Utilities
```typescript
// webapp/src/utils/formatting.ts
const priceCache = new Map<string, { price: number; timestamp: number }>();
const PRICE_CACHE_TTL = 5 * 60 * 1000; // 5 minutes

export function formatTokenAmount(amount: string, denom: string): string {
  const exponent = getDenomExponent(denom);
  const value = BigInt(amount);
  const divisor = BigInt(10 ** exponent);
  
  const whole = value / divisor;
  const fraction = value % divisor;
  
  if (fraction === 0n) {
    return `${whole} ${getDenomSymbol(denom)}`;
  }
  
  const fractionStr = fraction.toString().padStart(exponent, '0');
  const trimmed = fractionStr.replace(/0+$/, '');
  return `${whole}.${trimmed} ${getDenomSymbol(denom)}`;
}

export async function getPriceUSD(denom: string): Promise<number | null> {
  // Check cache first
  const cached = priceCache.get(denom);
  if (cached && Date.now() - cached.timestamp < PRICE_CACHE_TTL) {
    return cached.price;
  }
  
  try {
    const response = await fetch(`/api/v1/prices/${denom}`);
    if (!response.ok) throw new Error('Price fetch failed');
    
    const data = await response.json();
    const price = data.price;
    
    // Cache the result
    priceCache.set(denom, { price, timestamp: Date.now() });
    return price;
  } catch (error) {
    console.error('Failed to fetch price:', error);
    
    // Return last known price if available
    if (cached) {
      return cached.price;
    }
    
    return null;
  }
}

export function formatUSDEstimate(amount: string, denom: string): string {
  const price = getPriceUSD(denom);
  if (!price) return '';
  
  const value = parseFloat(formatTokenAmount(amount, denom));
  const usd = value * price;
  
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
    minimumFractionDigits: 2,
    maximumFractionDigits: 2
  }).format(usd);
}

// Fee comparison context
export function calculateSavings(ourFee: string, estimatedRetryGas: string, denom: string): string {
  const ourAmount = BigInt(ourFee);
  const retryAmount = BigInt(estimatedRetryGas) * 3n; // Assume 3 retry attempts
  const savings = retryAmount - ourAmount;
  
  if (savings <= 0n) return '';
  return formatTokenAmount(savings.toString(), denom);
}
```

### Fee Display Component Updates
```vue
<!-- webapp/src/components/clearing/FeeEstimator.vue -->
<template>
  <div class="fee-breakdown">
    <div class="fee-row">
      <span>Service Fee:</span>
      <span class="fee-amount">
        {{ formatTokenAmount(fees.serviceFee, fees.denom) }}
        <span class="usd-estimate">{{ formatUSDEstimate(fees.serviceFee, fees.denom) }}</span>
      </span>
    </div>
    
    <div class="fee-row">
      <span>Network Fee:</span>
      <span class="fee-amount">
        ~{{ formatTokenAmount(fees.estimatedGasFee, fees.denom) }}
        <Tooltip content="Actual network fee may vary slightly" />
      </span>
    </div>
    
    <div class="fee-total">
      <span>Total:</span>
      <span class="total-amount">
        {{ formatTokenAmount(fees.totalRequired, fees.denom) }}
        <span class="usd-estimate">{{ formatUSDEstimate(fees.totalRequired, fees.denom) }}</span>
      </span>
    </div>
    
    <div class="fee-comparison" v-if="savingsAmount">
      <Icon name="trending-down" />
      <span>~{{ savingsAmount }} cheaper than retrying manually</span>
    </div>
  </div>
</template>
```

## 6. Tooltips for Technical Terms

### Tooltip System
```typescript
// webapp/src/utils/tooltips.ts
export const TERM_DEFINITIONS: Record<string, string> = {
  'channel': 'An IBC channel is a connection between two blockchains that allows them to exchange tokens and data',
  'sequence': 'A unique number identifying each transfer in order',
  'packet': 'A bundle of data being transferred between blockchains',
  'timeout': 'The time limit for a transfer to complete before it expires',
  'relayer': 'A service that helps move transfers between blockchains',
  'stuck': 'A transfer that hasn\'t completed within the expected time (usually 15+ minutes)',
  'memo': 'A message attached to your payment that tells our system which transfers to clear',
  'gas': 'Network fees paid to process transactions on the blockchain'
};
```

### Tooltip Component
```vue
<!-- webapp/src/components/ui/Tooltip.vue -->
<template>
  <span class="tooltip-wrapper" @click="handleClick">
    <slot />
    <Icon name="help-circle" class="tooltip-icon" />
    <div 
      class="tooltip-content" 
      v-if="showTooltip"
      :class="{ 'tooltip-mobile': isMobile }"
    >
      {{ content || TERM_DEFINITIONS[term] }}
    </div>
  </span>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'

const props = defineProps<{
  term?: string
  content?: string
}>()

const showTooltip = ref(false)
const isMobile = ref(false)

onMounted(() => {
  isMobile.value = 'ontouchstart' in window
})

const handleClick = () => {
  if (isMobile.value) {
    showTooltip.value = !showTooltip.value
  }
}

// Desktop hover handled via CSS
</script>
```

### Integration in Packet Display
```vue
<!-- Update PacketSelector.vue -->
<template>
  <div class="packet-info">
    <span>
      <Tooltip term="channel">Channel</Tooltip>: 
      {{ packet.sourceChannel }}
    </span>
    <span>
      <Tooltip term="sequence">Sequence</Tooltip>: 
      {{ packet.sequence }}
    </span>
    <span>
      <Tooltip term="stuck">Stuck for</Tooltip>: 
      {{ formatDuration(packet.age) }}
    </span>
  </div>
</template>
```

## Testing Considerations

1. **Wallet Integration Tests**
   - Mock Keplr responses
   - Test transaction signing flow
   - Handle wallet rejection scenarios

2. **QR Code Tests**
   - Verify correct URI generation
   - Test with different mobile wallets
   - Ensure proper encoding of special characters

3. **Simple Mode Tests**
   - Verify automatic packet selection
   - Test fee calculations
   - Ensure mode persistence

4. **Error Handling Tests**
   - Test all error scenarios
   - Verify error message formatting
   - Check action suggestions work

5. **Fee Display Tests**
   - Test with different denominations
   - Verify USD calculations
   - Test edge cases (very small/large amounts)