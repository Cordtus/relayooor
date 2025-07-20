# Frontend Module Blueprint

## Module Overview

The frontend module provides the user interface for the Relayooor platform, built with Vue 3 and TypeScript. It offers a comprehensive dashboard for monitoring and clearing stuck IBC packets.

## Architecture

### Technology Stack
- **Framework**: Vue 3 with Composition API
- **Language**: TypeScript
- **State Management**: Pinia
- **Styling**: Tailwind CSS
- **Build Tool**: Vite
- **Router**: Vue Router
- **HTTP Client**: Axios
- **WebSocket**: Native WebSocket API
- **Testing**: Vitest + Vue Test Utils

### Directory Structure
```
webapp/
├── src/
│   ├── components/         # Reusable UI components
│   │   ├── clearing/      # Packet clearing components
│   │   ├── monitoring/    # Monitoring dashboard components
│   │   ├── analytics/     # Analytics components
│   │   ├── channels/      # Channel management
│   │   ├── relayers/      # Relayer management
│   │   └── common/        # Shared components
│   ├── views/             # Page-level components
│   │   ├── Dashboard.vue
│   │   ├── Monitoring.vue
│   │   ├── PacketClearing.vue
│   │   ├── Analytics.vue
│   │   ├── Channels.vue
│   │   ├── Relayers.vue
│   │   └── Settings.vue
│   ├── stores/            # Pinia state stores
│   │   ├── auth.ts       # Authentication state
│   │   ├── packets.ts    # Packet data
│   │   ├── chains.ts     # Chain information
│   │   ├── user.ts       # User preferences
│   │   └── websocket.ts  # WebSocket connection
│   ├── services/          # API and business logic
│   │   ├── api/          # API client modules
│   │   ├── wallet/       # Wallet integration
│   │   └── utils/        # Utility functions
│   ├── types/            # TypeScript type definitions
│   ├── router/           # Vue Router configuration
│   ├── assets/           # Static assets
│   └── config/           # Configuration files
├── public/               # Public static files
├── tests/               # Test files
└── package.json         # Dependencies

```

## Key Components

### 1. Dashboard Component
**Purpose**: Main landing page showing overview metrics
**Features**:
- Real-time packet statistics
- Recent activity feed
- Quick actions menu
- Chain status indicators

### 2. Packet Clearing Wizard
**Purpose**: Multi-step process for clearing stuck packets
**Steps**:
1. Select stuck packet
2. Verify ownership
3. Generate clearing token
4. Make payment
5. Track clearing status

### 3. Monitoring Dashboard
**Purpose**: Real-time IBC metrics visualization
**Features**:
- Live packet flow
- Channel congestion maps
- Chain health indicators
- Alert notifications

### 4. Wallet Integration
**Supported Wallets**:
- Keplr
- Leap
- Cosmostation
- Station (Terra)

**Integration Pattern**:
```typescript
// Wallet service interface
interface WalletService {
  connect(): Promise<void>
  disconnect(): void
  signMessage(message: string): Promise<string>
  getAddress(): string
  getChainId(): string
}
```

## State Management

### Pinia Stores

#### Auth Store
```typescript
interface AuthState {
  isAuthenticated: boolean
  user: User | null
  token: string | null
  walletAddress: string | null
}
```

#### Packets Store
```typescript
interface PacketsState {
  stuckPackets: Packet[]
  recentPackets: Packet[]
  filters: PacketFilters
  loading: boolean
}
```

#### WebSocket Store
```typescript
interface WebSocketState {
  connected: boolean
  reconnectAttempts: number
  messageQueue: Message[]
}
```

## API Integration

### API Client Structure
```typescript
// Base API client
class APIClient {
  private axios: AxiosInstance
  
  constructor(baseURL: string) {
    this.axios = axios.create({
      baseURL,
      timeout: 30000
    })
    
    // Request interceptor for auth
    this.axios.interceptors.request.use(
      config => {
        const token = useAuthStore().token
        if (token) {
          config.headers.Authorization = `Bearer ${token}`
        }
        return config
      }
    )
  }
}

// Service modules
export const packetsAPI = new PacketsAPI()
export const clearingAPI = new ClearingAPI()
export const authAPI = new AuthAPI()
```

### Error Handling
```typescript
// Global error handler
export function handleAPIError(error: AxiosError) {
  if (error.response?.status === 401) {
    // Redirect to login
    router.push('/login')
  } else if (error.response?.status === 429) {
    // Show rate limit message
    showToast('Too many requests. Please try again later.')
  } else {
    // Generic error
    showToast(error.response?.data?.message || 'An error occurred')
  }
}
```

## WebSocket Integration

### Connection Management
```typescript
class WebSocketService {
  private ws: WebSocket | null = null
  private reconnectTimer: number | null = null
  
  connect() {
    const wsUrl = import.meta.env.VITE_WS_URL
    this.ws = new WebSocket(wsUrl)
    
    this.ws.onopen = () => {
      console.log('WebSocket connected')
      this.authenticate()
    }
    
    this.ws.onmessage = (event) => {
      const message = JSON.parse(event.data)
      this.handleMessage(message)
    }
    
    this.ws.onclose = () => {
      this.scheduleReconnect()
    }
  }
  
  private handleMessage(message: WSMessage) {
    switch (message.type) {
      case 'packet_update':
        usePacketsStore().updatePacket(message.data)
        break
      case 'clearing_status':
        useClearingStore().updateStatus(message.data)
        break
    }
  }
}
```

## Routing

### Route Configuration
```typescript
const routes = [
  {
    path: '/',
    component: Dashboard,
    meta: { requiresAuth: false }
  },
  {
    path: '/monitoring',
    component: Monitoring,
    meta: { requiresAuth: false }
  },
  {
    path: '/clearing',
    component: PacketClearing,
    meta: { requiresAuth: true }
  },
  {
    path: '/analytics',
    component: Analytics,
    meta: { requiresAuth: false }
  }
]

// Navigation guard
router.beforeEach((to, from, next) => {
  const authStore = useAuthStore()
  
  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    next('/login')
  } else {
    next()
  }
})
```

## Testing Strategy

### Component Testing
```typescript
// Example component test
describe('PacketCard.vue', () => {
  it('displays packet information correctly', () => {
    const packet = createMockPacket()
    const wrapper = mount(PacketCard, {
      props: { packet }
    })
    
    expect(wrapper.find('.packet-id').text()).toBe(packet.id)
    expect(wrapper.find('.packet-status').text()).toBe('Stuck')
  })
  
  it('emits clear event when button clicked', async () => {
    const wrapper = mount(PacketCard, {
      props: { packet: createMockPacket() }
    })
    
    await wrapper.find('.clear-button').trigger('click')
    expect(wrapper.emitted('clear')).toBeTruthy()
  })
})
```

### Store Testing
```typescript
describe('Packets Store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })
  
  it('fetches stuck packets', async () => {
    const store = usePacketsStore()
    await store.fetchStuckPackets()
    
    expect(store.stuckPackets).toHaveLength(5)
    expect(store.loading).toBe(false)
  })
})
```

## Build Configuration

### Vite Config
```typescript
export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  },
  server: {
    port: 5173,
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true
      },
      '/ws': {
        target: 'ws://localhost:8080',
        ws: true
      }
    }
  },
  build: {
    target: 'es2015',
    outDir: 'dist',
    sourcemap: true
  }
})
```

## Environment Configuration

### Environment Variables
```env
# API Configuration
VITE_API_URL=http://localhost:8080/api
VITE_WS_URL=ws://localhost:8080/ws

# Feature Flags
VITE_ENABLE_ANALYTICS=true
VITE_ENABLE_BATCH_CLEARING=false

# External Services
VITE_PRICE_API_URL=https://api.coingecko.com/api/v3
```

## Performance Optimization

### Code Splitting
```typescript
// Lazy load routes
const Analytics = () => import('./views/Analytics.vue')
const Settings = () => import('./views/Settings.vue')
```

### Component Optimization
```typescript
// Use shallowRef for large data sets
const packets = shallowRef<Packet[]>([])

// Memoize expensive computations
const filteredPackets = computed(() => {
  return packets.value.filter(p => p.status === 'stuck')
})
```

## Security Considerations

### Content Security Policy
```html
<meta http-equiv="Content-Security-Policy" 
      content="default-src 'self'; 
               script-src 'self' 'unsafe-inline'; 
               style-src 'self' 'unsafe-inline'; 
               connect-src 'self' ws://localhost:8080">
```

### Input Sanitization
```typescript
// Sanitize user input
import DOMPurify from 'dompurify'

function sanitizeInput(input: string): string {
  return DOMPurify.sanitize(input, { 
    ALLOWED_TAGS: [],
    ALLOWED_ATTR: [] 
  })
}
```

## Deployment

### Docker Configuration
```dockerfile
# Build stage
FROM node:18-alpine as build
WORKDIR /app
COPY package*.json ./
RUN yarn install --frozen-lockfile
COPY . .
RUN yarn build

# Production stage
FROM nginx:alpine
COPY --from=build /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/nginx.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

### Nginx Configuration
```nginx
server {
  listen 80;
  root /usr/share/nginx/html;
  
  location / {
    try_files $uri $uri/ /index.html;
  }
  
  location /api {
    proxy_pass http://api-backend:8080;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
  }
  
  location /ws {
    proxy_pass http://api-backend:8080;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
  }
}
```

## Common Issues

### Issue: Wallet connection fails
**Solution**: Ensure wallet extension is installed and unlocked

### Issue: WebSocket disconnects frequently
**Solution**: Check proxy configuration and increase timeout values

### Issue: API calls fail with CORS errors
**Solution**: Verify API CORS configuration allows frontend origin

### Issue: Build fails with memory error
**Solution**: Increase Node.js memory limit: `NODE_OPTIONS=--max-old-space-size=4096`