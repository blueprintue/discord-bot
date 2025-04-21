# syntax=docker/dockerfile:1

FROM golang:1.24-alpine AS base
ENV GOFLAGS="-buildvcs=false"
RUN apk add --no-cache gcc linux-headers musl-dev
WORKDIR /src

FROM golangci/golangci-lint:v1.64.8-alpine AS golangci-lint
FROM base AS lint
RUN --mount=type=bind,target=. \
  --mount=type=cache,target=/root/.cache \
  --mount=from=golangci-lint,source=/usr/bin/golangci-lint,target=/usr/bin/golangci-lint \
  golangci-lint run ./...