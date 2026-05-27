# Multi-stage build
FROM --platform=$BUILDPLATFORM node:22-alpine AS builder

# Set working directory
WORKDIR /app

# Copy package files from ui/dashboard
COPY ui/dashboard/package.json ui/dashboard/pnpm-lock.yaml ui/dashboard/pnpm-workspace.yaml ./

# Install dependencies
RUN corepack enable && pnpm install --frozen-lockfile

# Copy source code from ui/dashboard
COPY ui/dashboard/ ./

# Accept build-time variables
ARG VITE_API_BASE_URL=/api/v1

# Build the application
ENV VITE_API_BASE_URL=$VITE_API_BASE_URL
RUN pnpm run build

# Production stage
FROM nginx:alpine

# Copy built assets from builder stage
COPY --from=builder /app/dist /usr/share/nginx/html

# Copy nginx configuration
COPY docker/dashboard/nginx.conf /etc/nginx/nginx.conf

# Expose port 3000
EXPOSE 3000

# Start nginx
CMD ["nginx", "-g", "daemon off;"]
