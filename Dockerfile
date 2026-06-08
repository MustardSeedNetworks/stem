# =============================================================================
# Stem container image — multi-stage, CGO-free
# =============================================================================
# Stem builds with CGO_ENABLED=0 across all targets (pure-Go networking, no
# libpcap), so the runtime stage doesn't need the libpcap0.8 / libcap2-bin
# packages that niac and seed ship. Otherwise the shape mirrors them so all
# three projects converge on the same container build pattern.
# =============================================================================

# -----------------------------------------------------------------------------
# Stage 1: build the embedded React/Vite UI
# -----------------------------------------------------------------------------
FROM node:26-bookworm AS ui-build
WORKDIR /src/ui
COPY ui/package.json ui/package-lock.json ./
RUN npm ci
COPY ui/ ./
# Vite's @locales alias resolves to ../internal/i18n/locales (sibling to
# ui/); bring that tree into the build context so TypeScript can resolve
# the imports.
COPY internal/i18n/locales /src/internal/i18n/locales
RUN npm run build
# Vite outputs to ../internal/api/ui (via outDir in vite.config.ts); copy that
# tree out so the next stage can mount it via COPY --from.
RUN mkdir -p /out && cp -r ../internal/api/ui /out/ui

# -----------------------------------------------------------------------------
# Stage 2: build the Go binary (pure Go, no CGO)
# -----------------------------------------------------------------------------
FROM golang:1.26-bookworm AS go-build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Place the prebuilt UI where //go:embed expects it, then build with CGO off.
COPY --from=ui-build /out/ui ./internal/api/ui
ARG VERSION=dev
ARG COMMIT=unknown
ARG BUILD_DATE=unknown
RUN UI_HASH=$(find internal/api/ui -type f -exec md5sum {} \; | sort | md5sum | cut -d' ' -f1) \
    && CGO_ENABLED=0 go build -trimpath -buildvcs=false \
        -ldflags="-s -w -X github.com/MustardSeedNetworks/stem/internal/version.Version=${VERSION} -X github.com/MustardSeedNetworks/stem/internal/version.Commit=${COMMIT} -X github.com/MustardSeedNetworks/stem/internal/version.BuildTime=${BUILD_DATE} -X github.com/MustardSeedNetworks/stem/internal/version.UIBuildHash=${UI_HASH}" \
        -o /out/stem ./cmd/stem

# -----------------------------------------------------------------------------
# Stage 3: minimal runtime
# -----------------------------------------------------------------------------
FROM debian:bookworm-slim AS runtime
RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        ca-certificates \
        libcap2-bin \
        tini \
    && rm -rf /var/lib/apt/lists/* \
    && groupadd --system stem \
    && useradd --system --gid stem --home-dir /var/lib/stem --shell /usr/sbin/nologin stem \
    && mkdir -p /etc/stem /var/lib/stem /var/log/stem \
    && chown -R stem:stem /etc/stem /var/lib/stem /var/log/stem \
    && chmod 0750 /etc/stem /var/lib/stem /var/log/stem

COPY --from=go-build /out/stem /usr/bin/stem
# Raw-socket capability so the daemon can run as the unprivileged stem user.
RUN setcap 'cap_net_raw,cap_net_admin=+ep' /usr/bin/stem

USER stem
WORKDIR /var/lib/stem
EXPOSE 8444 8043

# OCI labels for traceability.
ARG VERSION=dev
ARG COMMIT=unknown
LABEL org.opencontainers.image.title="stem" \
      org.opencontainers.image.source="https://github.com/MustardSeedNetworks/stem" \
      org.opencontainers.image.licenses="Proprietary" \
      org.opencontainers.image.version="${VERSION}" \
      org.opencontainers.image.revision="${COMMIT}"

# tini reaps zombies and forwards signals so SIGTERM cleanly shuts down the daemon.
ENTRYPOINT ["/usr/bin/tini", "--", "/usr/bin/stem"]
CMD ["web"]
