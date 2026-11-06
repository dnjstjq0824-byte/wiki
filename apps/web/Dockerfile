# Production build for the Treewiki web frontend (Nuxt, node-server preset).
# Context is the repository root.
# Build: docker build -t treewiki-web -f apps/web/Dockerfile .
FROM node:22-alpine AS deps
RUN corepack enable
WORKDIR /repo
COPY package.json pnpm-workspace.yaml pnpm-lock.yaml ./
COPY apps/web/package.json apps/web/package.json
RUN pnpm install --filter @treewiki/web --frozen-lockfile

FROM node:22-alpine AS builder
RUN corepack enable
WORKDIR /repo
COPY --from=deps /repo/node_modules ./node_modules
COPY --from=deps /repo/apps/web/node_modules ./apps/web/node_modules
COPY package.json pnpm-workspace.yaml ./
COPY apps/web ./apps/web
ENV NITRO_PRESET=node-server
RUN pnpm --filter @treewiki/web build

FROM node:22-alpine AS runtime
RUN adduser -D -H -u 10001 treewiki
WORKDIR /app
COPY --from=builder /repo/apps/web/.output ./
RUN chown -R treewiki:treewiki /app
USER treewiki
ENV PORT=3000
EXPOSE 3000
HEALTHCHECK --interval=15s --timeout=3s --retries=3 \
  CMD node -e "fetch('http://127.0.0.1:3000/').then(r=>{if(!r.ok)process.exit(1)}).catch(()=>process.exit(1))"
CMD ["node", "server/index.mjs"]
