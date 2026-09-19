# syntax=docker/dockerfile:1.7

FROM golang:1.25-bookworm AS build-base

WORKDIR /src

ENV CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GOFLAGS=-mod=readonly

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

COPY . ./

FROM build-base AS api-build
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

FROM build-base AS migrate-build
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath -ldflags="-s -w" -o /out/migrate ./cmd/migrate

FROM build-base AS worker-build
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath -ldflags="-s -w" -o /out/storage-worker ./cmd/storage-worker

FROM gcr.io/distroless/base-debian12:nonroot AS api

WORKDIR /app
COPY --from=api-build /out/api /app/api

EXPOSE 8080
ENTRYPOINT ["/app/api"]

FROM gcr.io/distroless/base-debian12:nonroot AS migrate

WORKDIR /app
COPY --from=migrate-build /out/migrate /app/migrate
COPY --from=migrate-build /src/migrations /app/migrations

ENTRYPOINT ["/app/migrate"]

FROM gcr.io/distroless/base-debian12:nonroot AS worker

WORKDIR /app
COPY --from=worker-build /out/storage-worker /app/storage-worker

ENTRYPOINT ["/app/storage-worker"]
