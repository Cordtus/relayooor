# Vue.js Frontend Style Guide

## Overview
This guide documents the standardized styling patterns and components used throughout the application. All components should follow these patterns to maintain consistency.

## Design Tokens

### Colors

#### Primary Colors
- Primary shades: `primary-50` through `primary-900`
- Main primary: `primary-600`
- Primary hover: `primary-700`

#### Status Colors
- Success: `status-success` (with `-light` and `-dark` variants)
- Warning: `status-warning` (with `-light` and `-dark` variants)
- Error: `status-error` (with `-light` and `-dark` variants)
- Info: `status-info` (with `-light` and `-dark` variants)

#### Surface Colors
- Card background: `surface-card`
- Page background: `surface-background`
- Muted background: `surface-muted`
- Hover state: `surface-hover`

#### Text Colors
- Primary text: `content-primary`
- Secondary text: `content-secondary`
- Muted text: `content-muted`
- Inverse text: `content-inverse`

### Spacing
- Card padding: `p-card` (1.5rem)
- Section spacing: `space-y-section` (2rem)
- Standard gap: `gap-card`

### Shadows
- Card shadow: `shadow-card`
- Card hover shadow: `shadow-card-hover`

## Component Patterns

### Cards
Use the standardized card styles:
```vue
<div class="card-base">
  <!-- Card content -->
</div>

<!-- With hover effect -->
<div class="card-base card-hover">
  <!-- Card content -->
</div>
```

### Buttons
Use the Button component with appropriate variants:
```vue
<Button variant="primary">Primary Action</Button>
<Button variant="secondary">Secondary Action</Button>
<Button variant="ghost">Ghost Button</Button>
<Button variant="danger">Danger Action</Button>

<!-- With sizes -->
<Button size="sm">Small</Button>
<Button size="md">Medium (default)</Button>
<Button size="lg">Large</Button>

<!-- With states -->
<Button :loading="true">Loading...</Button>
<Button :disabled="true">Disabled</Button>
```

### Badges
Use the Badge component for status indicators:
```vue
<Badge variant="success">Success</Badge>
<Badge variant="warning">Warning</Badge>
<Badge variant="error">Error</Badge>
<Badge variant="info">Info</Badge>

<!-- With dot indicator -->
<Badge variant="success" dot>Active</Badge>
```

### Empty States
Use the EmptyState component for placeholder content:
```vue
<EmptyState
  :icon="PackageOpen"
  title="No packets found"
  description="There are no stuck packets to display"
>
  <template #action>
    <Button variant="primary">Refresh</Button>
  </template>
</EmptyState>
```

### Form Elements
Use consistent form styling:
```vue
<!-- Input -->
<label class="label-base">Field Label</label>
<input type="text" class="input-base" />

<!-- With validation states -->
<input type="text" class="input-base border-status-error" />
```

### Status Indicators
Use semantic colors for status:
```vue
<!-- Status dots -->
<span class="status-dot-success" />
<span class="status-dot-warning" />
<span class="status-dot-error" />

<!-- Status badges -->
<StatusBadge status="completed" />
<StatusBadge status="pending" />
<StatusBadge status="failed" />
```

## Layout Patterns

### Page Layout
```vue
<div class="min-h-screen bg-surface-background">
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
    <div class="section-spacing">
      <!-- Page sections -->
    </div>
  </div>
</div>
```

### Card Grid
```vue
<div class="card-grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3">
  <div class="card-base">...</div>
  <div class="card-base">...</div>
  <div class="card-base">...</div>
</div>
```

## Typography

### Headings
- Page title: `text-2xl font-bold text-content-primary`
- Section title: `text-xl font-semibold text-content-primary`
- Card title: `text-lg font-medium text-content-primary`

### Body Text
- Regular: `text-base text-content-primary`
- Secondary: `text-sm text-content-secondary`
- Muted: `text-sm text-content-muted`

## Utilities

### Scrollbar
For custom scrollbars:
```vue
<div class="scrollbar-thin">
  <!-- Scrollable content -->
</div>
```

### Transitions
Always use consistent transitions:
- Duration: `duration-200`
- Colors: `transition-colors`
- Shadow: `transition-shadow`
- All: `transition-all`

### Focus States
Ensure all interactive elements have focus states:
- `focus:outline-none focus:ring-2 focus:ring-primary-500 focus:ring-offset-2`

## Best Practices

1. **Use Design Tokens**: Always use semantic color tokens instead of hardcoded colors
2. **Component First**: Use standardized components (Button, Badge, etc.) instead of custom implementations
3. **Consistent Spacing**: Use the defined spacing scale (card, section) for consistency
4. **Semantic HTML**: Use appropriate HTML elements for better accessibility
5. **Responsive Design**: Use Tailwind's responsive utilities (sm:, md:, lg:)
6. **Dark Mode Ready**: Use semantic colors that can be easily themed

## Migration Guide

### Replacing Old Patterns

#### Old Card Pattern
```vue
<!-- OLD -->
<div class="bg-white rounded-lg shadow p-6">

<!-- NEW -->
<div class="card-base">
```

#### Old Button Pattern
```vue
<!-- OLD -->
<button class="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700">

<!-- NEW -->
<Button variant="primary">
```

#### Old Status Colors
```vue
<!-- OLD -->
<span class="bg-green-100 text-green-800">

<!-- NEW -->
<Badge variant="success">
```

## Component Library

### Core Components
- **Button**: Standardized button with variants
- **Badge**: Status badges with semantic colors
- **EmptyState**: Consistent empty state displays
- **Card**: Base card component (use shadcn card)
- **MetricCard**: Dashboard metric displays

### Composite Components
- **StatusBadge**: Status indicators using Badge
- **ClearingWizard**: Multi-step form using Button
- **PacketSelector**: List with EmptyState

## Resources
- [Tailwind CSS Documentation](https://tailwindcss.com/docs)
- [shadcn/ui Components](https://ui.shadcn.com/)
- [Vue 3 Style Guide](https://vuejs.org/style-guide/)