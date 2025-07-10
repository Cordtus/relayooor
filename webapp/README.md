# Relayooor Web App

User-friendly dashboard for IBC monitoring and packet clearing with wallet integration.

## Features

- **Real-time Monitoring**: Live IBC metrics and packet flow visualization
- **Wallet Integration**: Connect Keplr wallet to manage packets
- **Packet Clearing**: Clear individual packets or entire channels
- **Channel Overview**: View all IBC channels and their status
- **Responsive Design**: Works on desktop and mobile devices

## Tech Stack

- React 18 with TypeScript
- Vite for fast development
- TailwindCSS for styling
- React Query for data fetching
- Chart.js for visualizations
- CosmJS for blockchain interactions
- Socket.io for real-time updates

## Prerequisites

- Node.js 20+
- Keplr wallet extension (for packet clearing)

## Development

1. Install dependencies:
```bash
yarn install
```

2. Start development server:
```bash
yarn dev
```

3. Open http://localhost:3000

## API Integration

The app expects the following backend services:
- API backend on port 8080
- WebSocket support for real-time updates

Configure API endpoints in `vite.config.ts`.

## Building for Production

```bash
yarn build
```

The built files will be in the `dist` directory.

## Docker Deployment

Build the Docker image:
```bash
docker build -t relayooor-webapp .
```

Run the container:
```bash
docker run -p 80:80 relayooor-webapp
```

## Environment Variables

Create a `.env` file for configuration:
```env
VITE_API_URL=http://localhost:8080
VITE_WS_URL=ws://localhost:8080
```

## Pages

### Dashboard
Main overview with key metrics:
- Stuck packets count
- Active channels
- Packet flow rate
- Success rate
- Real-time packet flow chart
- Top stuck packets

### Channels
List of all IBC channels with:
- Channel status (open/closed)
- Source and destination chains
- Pending packets count
- Total packets processed

### Packet Clearing
Wallet-connected interface for:
- Viewing stuck packets
- Filtering by wallet address
- Selecting packets to clear
- Clearing entire channels
- Batch operations

### Settings
Configuration options:
- Monitoring preferences
- Refresh intervals
- Service connection status

## Wallet Integration

The app uses Keplr wallet for:
- User authentication
- Signing transactions
- Filtering packets by sender

Supported chains:
- Osmosis
- Cosmos Hub
- Sei Network

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## License

MIT