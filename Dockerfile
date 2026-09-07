# ==============================================================================
# Stage 1: Builder
# ==============================================================================
FROM golang:1.23-alpine AS builder

# Allow Go to automatically manage toolchain version requirements.
ENV GOTOOLCHAIN=auto

WORKDIR /app

# Copy dependency manifests and download modules.
# This layer is cached and only re-run when go.mod or go.sum changes.
COPY go.mod go.sum ./

# Use BuildKit's mount cache to avoid re-downloading modules on every build.
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download -x

# Copy the full source tree.
COPY . .

# Compile a statically-linked, stripped binary.
#   CGO_ENABLED=0  — no C dependencies, fully portable
#   -w             — strip DWARF debug info
#   -s             — strip symbol table
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" -o /app/gotickets ./cmd/main.go


# ==============================================================================
# Stage 2: Runner
# A distroless-equivalent minimal image — no shell, no package manager,
# no attack surface. Only the binary runs inside.
# ==============================================================================
FROM alpine:3.20

# Install runtime dependencies:
#   ca-certificates — required for outbound HTTPS calls
#   tzdata          — required for timezone-aware time formatting
RUN apk --no-cache add ca-certificates tzdata

# Run as a non-root user to follow the principle of least privilege.
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copy only the compiled binary from the builder stage.
COPY --from=builder --chown=appuser:appgroup /app/gotickets .

USER appuser

EXPOSE 5525

ENTRYPOINT ["./gotickets"]
