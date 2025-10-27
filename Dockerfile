# ============================================================
# 🧱 Runtime Image (uses prebuilt Next.js output)
# ============================================================

FROM node:20-alpine AS runtime

# Set working directory
WORKDIR /app

# Copy dependency files (for runtime only)
COPY package.json package-lock.json* ./

# Install production dependencies
RUN npm ci --omit=dev

# Copy prebuilt Next.js app from CI build artifacts
COPY .next/ .next/
COPY public/ public/
COPY next.config.* ./
COPY package.json ./

# Environment variables
ENV NODE_ENV=production
ENV PORT=3000

# Expose the Next.js port
EXPOSE 3000

# Start the production server
CMD ["npm", "start"]
