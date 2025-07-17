# Webapp Development Guide

## Overview

Vue.js 3 web application for the Relayooor packet clearing platform with TypeScript support.

## Tech Stack

- **Vue 3** with Composition API
- **TypeScript** with relaxed configuration for rapid development
- **Tailwind CSS** with custom design tokens
- **Vite** for fast development and building
- **Pinia** for state management
- **Vue Router** for navigation
- **Tanstack Query** for server state management

## Development Setup

### Prerequisites
- Node.js 20+
- Yarn package manager
- Docker (for M1/M4 Macs)

### Installation

```bash
cd webapp
yarn install
```

### Running Locally

#### Standard Development (Intel Macs/Linux)
```bash
yarn dev
```

#### M1/M4 Mac Development
Due to Vite server binding issues on Apple Silicon:

```bash
# Option 1: Use Docker
docker-compose -f ../docker-compose.full.yml up webapp

# Option 2: Use preview mode
yarn build --mode development
yarn preview

# Option 3: Try different host binding
yarn dev --host 0.0.0.0
```

### Building for Production

```bash
# TypeScript check and build
yarn build

# Build without TypeScript checking (faster)
npx vite build
```

## Project Structure

```
src/
├── components/          # Reusable Vue components
│   ├── ui/             # Base UI components (Button, Badge, etc.)
│   ├── clearing/       # Packet clearing components
│   ├── monitoring/     # Monitoring dashboard components
│   └── __tests__/      # Component tests
├── views/              # Page components
├── stores/             # Pinia stores
├── services/           # API services and external integrations
├── types/              # TypeScript type definitions
├── router/             # Vue Router configuration
├── lib/                # Utility functions
└── style.css          # Global styles and Tailwind imports
```

## Key Features

### Component Architecture
- **Composition API**: All components use `<script setup>` syntax
- **TypeScript**: Fully typed with interfaces for all data structures
- **Props Validation**: Runtime type checking with TypeScript
- **Emits Declaration**: Explicit event emissions

### State Management
- **Pinia Stores**: Centralized state for wallet, connection, and settings
- **Reactive Composables**: Shared logic with Vue composables
- **Server State**: Tanstack Query for API data caching

### Styling System
- **Tailwind CSS**: Utility-first styling
- **Design Tokens**: Semantic color system in `tailwind.config.js`
- **Component Classes**: Reusable styles in `@layer components`
- **Dark Mode Ready**: CSS variables for theme switching

### API Integration
- **Axios**: HTTP client with interceptors
- **WebSocket**: Real-time updates for clearing status
- **Type Safety**: Full TypeScript types for API responses

## Common Development Tasks

### Adding a New Component

1. Create component file in appropriate directory
2. Define props interface
3. Use composition API setup
4. Add unit tests

Example:
```vue
<template>
  <div class="component-class">
    {{ title }}
  </div>
</template>

<script setup lang="ts">
interface Props {
  title: string
  variant?: 'primary' | 'secondary'
}

const props = withDefaults(defineProps<Props>(), {
  variant: 'primary'
})
</script>
```

### Adding a New View

1. Create view component in `src/views/`
2. Add route in `src/router/index.ts`
3. Update navigation if needed

### Working with the API

```typescript
// Add new service method
import { api } from './api'

export const myService = {
  async getData(): Promise<MyData> {
    const response = await api.get('/endpoint')
    return response.data
  }
}
```

### TypeScript Configuration

The project uses a relaxed TypeScript configuration for faster development:
- `strict: false` - Allows more flexible typing
- `skipLibCheck: true` - Faster builds
- `allowImportingTsExtensions: true` - Import .ts files directly

## Testing

```bash
# Run unit tests
yarn test:unit

# Run with coverage
yarn test:unit --coverage
```

## Environment Variables

Create `.env.local` for local development:

```env
VITE_API_URL=http://localhost:8080
VITE_CHAINPULSE_URL=http://localhost:3000
VITE_WS_URL=ws://localhost:8080
```

## Troubleshooting

### Port Already in Use
```bash
# Find process using port 5173
lsof -i :5173
# Kill the process
kill -9 <PID>
```

### TypeScript Errors
- Check imports are using correct paths
- Ensure all types are properly exported
- Run `yarn type-check` to see all errors

### Vite Not Starting on M1/M4
- Use Docker or build & preview approach
- Check Node.js is native ARM64: `node -p "process.arch"`

### Component Not Rendering
- Check console for errors
- Verify props are passed correctly
- Ensure component is imported and registered

## Performance Tips

1. **Lazy Load Routes**: Use dynamic imports for large views
2. **Component Splitting**: Break large components into smaller ones
3. **Memo Computed**: Use `computed()` for expensive calculations
4. **List Rendering**: Always use `:key` with unique values
5. **Image Optimization**: Use appropriate formats and sizes

## Code Style

- Use Composition API with `<script setup>`
- Prefer `const` over `let`
- Use TypeScript interfaces over types
- Follow Vue 3 style guide
- Keep components under 200 lines
- Extract complex logic to composables