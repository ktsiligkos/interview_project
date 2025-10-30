# syntax=docker/dockerfile:1

ARG GO_VERSION=1.22.12

FROM golang:${GO_VERSION}-alpine AS build

WORKDIR /workspace

RUN apk add --no-cache ca-certificates tzdata git

COPY go.mod go.sum ./
ENV GOTOOLCHAIN=auto
RUN go mod download

COPY . .

ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} \
	go build -ldflags="-s -w" -o /workspace/bin/xm-api ./cmd/api

FROM gcr.io/distroless/base-debian12 AS final

WORKDIR /app

COPY --from=build /workspace/bin/xm-api /usr/bin/xm-api
COPY --from=build /workspace/pkg/config ./pkg/config
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo

ENV HTTP_ADDR=:8081
EXPOSE 8081

ENTRYPOINT ["/usr/bin/xm-api"]
