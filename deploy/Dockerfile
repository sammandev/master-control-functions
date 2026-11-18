ARG BUN_VERSION=1.4.2
ARG NODE_VERSION=24.14.1
ARG GO_VERSION=1.27.3
ARG ALPINE_VERSION=3.23

FROM oven/bun:${BUN_VERSION}-alpine AS bun-toolchain

FROM node:${NODE_VERSION}-alpine AS frontend
ENV NODE_USE_ENV_PROXY=1
COPY --from=bun-toolchain /usr/local/bin/bun /usr/local/bin/bun
WORKDIR /src/frontend
COPY frontend/package.json frontend/pnpm-lock.yaml frontend/pnpm-workspace.yaml ./
# Network-reaching build steps mount the proxy settings as build secrets
# (id=http_proxy/https_proxy/no_proxy, optional — absent when the operator
# exports none), exposed only as environment variables inside that RUN.
RUN --mount=type=cache,target=/root/.bun/install/cache \
    --mount=type=secret,id=http_proxy,env=http_proxy \
    --mount=type=secret,id=https_proxy,env=https_proxy \
    --mount=type=secret,id=no_proxy,env=no_proxy \
    bun pm migrate && bun install --frozen-lockfile
COPY frontend/ ./
# svelte-check runs before any vite invocation, so the gitignored
# src/lib/paraglide output must be compiled here first (also loads the
# inlang plugin modules from cdn.jsdelivr.net; the proxy secrets apply).
RUN --mount=type=secret,id=http_proxy,env=http_proxy \
    --mount=type=secret,id=https_proxy,env=https_proxy \
    --mount=type=secret,id=no_proxy,env=no_proxy \
    node ./node_modules/@inlang/paraglide-js/bin/run.js compile --project ./project.inlang \
        --outdir ./src/lib/paraglide --emit-ts-declarations \
    && bun run check && bun run build

FROM golang:${GO_VERSION}-alpine AS backend
ARG GOPROXY=https://proxy.golang.org,direct
WORKDIR /src/backend
COPY backend/ ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=secret,id=http_proxy,env=http_proxy \
    --mount=type=secret,id=https_proxy,env=https_proxy \
    --mount=type=secret,id=no_proxy,env=no_proxy \
    CGO_ENABLED=0 go build -mod=readonly -trimpath -ldflags='-s -w' \
    -o /out/server ./cmd/server

FROM alpine:${ALPINE_VERSION} AS runtime
RUN apk add --no-cache ca-certificates \
    && addgroup -g 10001 app \
    && adduser -D -H -u 10001 -G app app
WORKDIR /app
COPY --from=backend /out/server /app/server
COPY --from=frontend /src/frontend/build /app/web
USER 10001:10001
ENTRYPOINT ["/app/server"]
CMD ["-addr", "127.0.0.1:24993", "-web", "/app/web"]