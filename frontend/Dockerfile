# Stage 1: Build the Vue.js application
FROM node:20-alpine AS builder

# Set working directory
WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies (including devDependencies needed for the build)
RUN npm ci --include=dev

# Copy source code
COPY . .

# Build the application (type-check + bundle)
RUN npm run build

# Stage 2: Serve with nginx
FROM nginx:alpine

# Copy built application from builder
COPY --from=builder /app/dist /usr/share/nginx/html

# Copy custom nginx configuration
COPY nginx.conf /etc/nginx/conf.d/default.conf

# Expose port
EXPOSE 80

# Start nginx
# daemon off keeps nginx in foreground for Docker
CMD ["nginx", "-g", "daemon off;"]
