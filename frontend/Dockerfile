# ── Stage 1: Build ──────────────────────────────────────────────────────────────
FROM node:20-alpine AS builder

WORKDIR /app

# Cache dependencies
COPY package.json package-lock.json* ./
RUN npm ci

# Build the static assets
COPY . .
RUN npm run build

# ── Stage 2: Nginx server ──────────────────────────────────────────────────────
FROM nginx:1.25-alpine

# Remove default nginx config
RUN rm /etc/nginx/conf.d/default.conf

# Copy custom nginx config that proxies /api to the backend
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Copy built frontend assets
COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]