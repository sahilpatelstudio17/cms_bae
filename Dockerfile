# syntax=docker/dockerfile:1

FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o app ./cmd/server

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/app /usr/local/bin/cms-server

RUN chmod +x /usr/local/bin/cms-server

ENV APP_ENV=production
ENV PORT=8080

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/cms-server"]