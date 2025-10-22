# ====== Stage 1: Build the site ======
FROM node:20-alpine AS builder

# Enable corepack for modern package managers (optional but recommended)
RUN corepack enable

WORKDIR /app

# Copy dependency manifests first (for caching)
COPY package*.json ./

# Install all dependencies cleanly
RUN npm ci

# Copy the rest of your project files
COPY . .

# Build the Nextra/Next.js app
RUN npm run build

# ====== Stage 2: Run optimized site ======
FROM node:20-alpine AS runner

WORKDIR /app

# Copy production build only
COPY --from=builder /app/package*.json ./
COPY --from=builder /app/.next ./.next
COPY --from=builder /app/public ./public
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/next.config.* ./
COPY --from=builder /app/tailwind.config.* ./
COPY --from=builder /app/postcss.config.* ./

ENV NODE_ENV=production
EXPOSE 3000

# Next.js 15 uses "next start" to serve the built app
CMD ["npm", "start"]

