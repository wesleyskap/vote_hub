# Build stage
FROM node:20-alpine AS builder
WORKDIR /app
COPY package.json package-lock.json ./
COPY packages/core/package.json ./packages/core/
COPY apps/web/package.json ./apps/web/
RUN npm ci

COPY . .
# Build production assets using Vite env vars
ARG VITE_RAILS_API_URL=http://localhost:3001
ARG VITE_GO_API_URL=http://localhost:8080
ENV VITE_RAILS_API_URL=$VITE_RAILS_API_URL
ENV VITE_GO_API_URL=$VITE_GO_API_URL

RUN npm run build

# Production stage
FROM nginx:alpine
COPY --from=builder /app/apps/web/dist /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
