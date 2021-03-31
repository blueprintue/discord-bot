# syntax=docker/dockerfile:1.2

FROM --platform=$BUILDPLATFORM crazymax/goreleaser-xx:latest AS goreleaser-xx
FROM --platform=$BUILDPLATFORM golang:alpine AS base
COPY --from=goreleaser-xx / /
RUN apk add --no-cache git
WORKDIR /src

FROM base AS build
ARG TARGETPLATFORM
RUN --mount=type=bind,source=.,target=/src,rw \
  goreleaser-xx --debug \
    --name="blueprintue-discord" \
    --dist="/out" \
    --hooks="go mod tidy" \
    --hooks="go mod download" \
    --ldflags="-s -w -X 'main.version={{.Version}}'" \
    --files="LICENSE" \
    --files="README.md"

FROM scratch AS artifact
COPY --from=build /out/*.tar.gz /
COPY --from=build /out/*.zip /

FROM alpine:3.13 AS image
RUN apk --update --no-cache add ca-certificates libressl shadow \
  && addgroup -g 1000 blueprintue-discord \
  && adduser -u 1000 -G blueprintue-discord -s /sbin/nologin -D blueprintue-discord \
  && mkdir -p /var/log/blueprintue-discord \
  && chown blueprintue-discord. /var/log/blueprintue-discord
COPY --from=build /usr/local/bin/blueprintue-discord /usr/local/bin/blueprintue-discord
USER blueprintue-discord
ENTRYPOINT [ "blueprintue-discord" ]