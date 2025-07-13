# Relayooor Web Application

Vue.js-based frontend for the IBC packet clearing platform, providing a user-friendly interface for clearing stuck IBC transfers.

## Features

- **Wallet Integration**: Connect with Keplr to view your stuck transfers
- **Packet Clearing Wizard**: 5-step process for secure packet clearing
- **Real-time Updates**: WebSocket integration for live clearing status
- **Multi-chain Support**: Works with Osmosis, Cosmos Hub, Neutron, and more
- **User Statistics**: Track your clearing history and success rates

## Development

### Prerequisites
- Node.js 16+ and Yarn
- Backend API running on port 3000
- Redis for WebSocket support

### Quick Start

```bash
# Install dependencies
yarn install

# Start development server
yarn dev

# Build for production
yarn build

# Run linting
yarn lint

# Run type checking
yarn typecheck
```

### Environment Variables

Create `.env` file:
```env
VITE_API_URL=http://localhost:3000
VITE_WS_URL=ws://localhost:3000
VITE_CHAINPULSE_URL=http://localhost:3001
```

## Project Structure

```
src/
├── components/         # Reusable UI components
│   ├── clearing/      # Packet clearing wizard components
│   ├── common/        # Shared components
│   └── stats/         # Statistics display components
├── views/             # Page components
│   ├── Home.vue       # Landing page
│   ├── Clearing.vue   # Packet clearing interface
│   └── Stats.vue      # User statistics page
├── services/          # API service layers
│   ├── api.ts         # Base API client
│   ├── clearing.ts    # Clearing operations
│   └── stats.ts       # Statistics fetching
├── stores/            # Pinia state management
│   ├── wallet.ts      # Wallet connection state
│   └── clearing.ts    # Clearing operation state
├── types/             # TypeScript type definitions
└── utils/             # Helper functions
```

## Key Components

### ClearingWizard.vue
Main component implementing the 5-step clearing process:
1. **Select** - Choose stuck packets to clear
2. **Fees** - Review service and gas fees
3. **Payment** - Make payment with memo
4. **Clearing** - Monitor clearing progress
5. **Complete** - View results

### WalletConnect.vue
Handles Keplr wallet integration:
- Chain switching
- Address management
- Message signing for authentication

### PacketList.vue
Displays stuck packets with:
- Sortable columns
- Batch selection
- Age indicators
- Value display

## Styling

Using TailwindCSS with custom configuration:
- Responsive design
- Dark mode support (future)
- Custom color palette
- Animation utilities

## Configuration

### Supported Chains
Configure in `src/config/chains.ts`:
```typescript
export const SUPPORTED_CHAINS = {
  'osmosis-1': {
    name: 'Osmosis',
    logo: '/osmosis-logo.svg',
    rpc: 'https://rpc.osmosis.zone',
    denom: 'uosmo'
  },
  // Add more chains...
}
```

### API Integration
All API calls go through `src/services/api.ts`:
- Automatic token refresh
- Error handling
- Request/response interceptors

## Testing

```bash
# Run unit tests (when configured)
yarn test

# Run e2e tests (when configured)
yarn test:e2e
```

## Building for Production

```bash
# Create production build
yarn build

# Preview production build
yarn preview

# Analyze bundle size
yarn build --report
```

## Deployment

The application is configured for static hosting:
- Nginx configuration in `../docker/nginx.conf`
- Environment variables injected at build time
- Health check endpoint at `/health`

### Docker Deployment
```bash
docker build -t relayooor-webapp .
docker run -p 80:80 relayooor-webapp
```

## Debugging

### Development Tools
- Vue DevTools for component inspection
- Network tab for API debugging
- Console for WebSocket messages

### Common Issues

**Wallet Connection Failed**
- Ensure Keplr is installed
- Check chain configuration
- Verify localhost is allowed in Keplr

**API Connection Issues**
- Verify backend is running
- Check CORS configuration
- Confirm API URL in .env

**WebSocket Disconnections**
- Check Redis is running
- Verify WebSocket URL
- Monitor network stability

## Contributing

1. Follow Vue 3 Composition API patterns
2. Use TypeScript for all new code
3. Maintain responsive design
4. Add appropriate error handling
5. Update tests for new features

## Resources

- [Vue 3 Documentation](https://vuejs.org/)
- [Pinia State Management](https://pinia.vuejs.org/)
- [TailwindCSS](https://tailwindcss.com/)
- [Keplr Wallet Docs](https://docs.keplr.app/)