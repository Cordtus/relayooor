# Relayooor Web Application

A comprehensive Vue.js dashboard for monitoring and managing IBC (Inter-Blockchain Communication) relay operations, powered by Chainpulse metrics.

## Features

### **Monitoring Dashboard**
- Real-time IBC packet flow visualization
- Network health status for all monitored chains
- Channel utilization heatmaps
- Relayer performance leaderboards
- Stuck packet alerts
- Frontrun competition analysis

### **Channel Analytics**
- Channel performance metrics and success rates
- Volume analysis and routing optimization
- Congestion detection and bottleneck identification
- Channel flow visualization

### **Relayer Analytics**
- Performance leaderboard with multiple metrics
- Market share distribution
- Software version tracking
- Competition dynamics (HHI index)
- Relayer churn analysis

### **Packet Clearing**
- User-friendly interface for clearing stuck IBC transfers
- Wallet integration (Keplr)
- Real-time stuck transfer detection
- One-click clearing functionality
- Transaction history

### **Advanced Analytics**
- Predictive volume and success rate modeling
- Network flow analysis
- Anomaly detection
- Performance optimization recommendations
- Export capabilities

## Tech Stack

- **Framework**: Vue 3 with TypeScript
- **Build Tool**: Vite
- **Styling**: Tailwind CSS
- **State Management**: Pinia
- **Data Fetching**: TanStack Query (Vue Query)
- **Charts**: Chart.js with vue-chartjs
- **Icons**: Lucide Vue
- **Date Handling**: date-fns
- **UI Components**: Radix Vue, shadcn-inspired components

## Getting Started

### Prerequisites

- Node.js 16+
- Yarn package manager
- Running Chainpulse instance (for real metrics)
- API backend running on port 8080

### Installation

```bash
# Install dependencies
yarn install

# Start development server
yarn dev
```

The application will be available at http://localhost:5173

### Configuration

The app expects the API backend to be running on http://localhost:8080. The Vite proxy is configured to forward `/api` requests.

## Available Scripts

- `yarn dev` - Start development server
- `yarn build` - Build for production
- `yarn preview` - Preview production build
- `yarn typecheck` - Run TypeScript type checking

## Project Structure

```
src/
├── components/         # Reusable UI components
│   ├── monitoring/    # Monitoring-specific components
│   ├── channels/      # Channel view components
│   ├── relayers/      # Relayer view components
│   └── analytics/     # Analytics components
├── views/             # Page components
├── router/            # Vue Router configuration
├── services/          # API services
├── types/             # TypeScript type definitions
├── lib/               # Utility functions
└── main.ts           # Application entry point
```

## Data Sources

The application fetches data from:
1. **Chainpulse Metrics** (via API proxy): Real-time Prometheus metrics
2. **API Backend**: Structured monitoring data and user-specific information

## Key Metrics Displayed

### System Metrics
- Total chains monitored
- Transaction and packet counts
- Connection health (reconnects, timeouts, errors)

### IBC Metrics
- Effected vs uneffected packets
- Success rates by channel and relayer
- Frontrun events and competition
- Stuck packet detection

### Derived Insights
- Peak activity periods
- Optimal routing recommendations
- Market concentration (HHI)
- Performance predictions

## Development Tips

1. **Mock Data**: The app includes mock data generators for development without a live Chainpulse instance
2. **Real-time Updates**: Uses TanStack Query for automatic refetching
3. **Responsive Design**: Fully responsive for desktop and tablet views
4. **Type Safety**: Comprehensive TypeScript types for all data structures

## Production Deployment

```bash
# Build for production
yarn build

# Preview production build
yarn preview
```

The build output will be in the `dist/` directory.

## Environment Variables

Create a `.env` file for custom configuration:

```env
VITE_API_URL=http://localhost:8080
```

## Contributing

1. Follow Vue 3 Composition API patterns
2. Use TypeScript for all new components
3. Maintain consistent styling with Tailwind classes
4. Add proper error handling for API calls
5. Include loading states for async operations