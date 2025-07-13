# Web Application Module Deployment

## Overview

The Vue.js web application provides the user interface. It consists of static files served through a web server.

## Requirements

- No runtime requirements (static files)
- Build process requires Node.js 16+
- 128MB RAM for serving

## Build Configuration

### Environment Variables (Build Time)

```bash
VITE_API_URL    # API endpoint (default: /api)
```

### Build Process

```bash
# Install dependencies
yarn install

# Build for production
yarn build

# Output in dist/ directory
```

## Fly.io Deployment

### Dockerfile

```dockerfile
# Build stage
FROM node:18-alpine as builder
WORKDIR /app
COPY package.json yarn.lock ./
RUN yarn install --frozen-lockfile
COPY . .
RUN yarn build

# Serve stage
FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/nginx.conf
EXPOSE 80
```

### nginx.conf

```nginx
server {
    listen 80;
    root /usr/share/nginx/html;
    index index.html;

    # Vue router support
    location / {
        try_files $uri $uri/ /index.html;
    }

    # API proxy
    location /api {
        proxy_pass http://relayooor-api.internal:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    # Cache static assets
    location ~* \.(js|css|png|jpg|jpeg|gif|ico|svg)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }
}
```

### fly.toml Configuration

```toml
app = "relayooor-webapp"
primary_region = "iad"

[build]
  dockerfile = "Dockerfile"

[[services]]
  internal_port = 80
  protocol = "tcp"
  auto_stop_machines = true
  auto_start_machines = true
  min_machines_running = 1

  [services.concurrency]
    type = "requests"
    hard_limit = 1000
    soft_limit = 800

  [[services.ports]]
    port = 80
    handlers = ["http"]
    
  [[services.ports]]
    port = 443
    handlers = ["tls", "http"]

  [[services.http_checks]]
    interval = "30s"
    grace_period = "5s"
    method = "get"
    path = "/"
    protocol = "http"
    timeout = "2s"
```

### Deployment Commands

```bash
# Deploy
fly deploy

# Scale based on traffic
fly autoscale set min=1 max=5

# Check deployment
fly status
```

## Performance Optimization

### Asset Optimization
- All JavaScript and CSS are minified
- Images are compressed
- Code splitting reduces initial bundle size

### Caching Strategy
- Static assets: 1 year cache
- index.html: No cache
- API responses: 5 second cache

## Browser Support

- Chrome/Edge 90+
- Firefox 88+
- Safari 14+

## Troubleshooting

### White screen
- Check browser console for errors
- Verify API URL is correct
- Clear browser cache

### API connection failed
- Verify API service is running
- Check CORS configuration
- Review browser network tab

### Slow loading
- Check CDN configuration
- Review bundle size (target < 500KB)
- Enable gzip compression