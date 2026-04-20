# syntax=docker/dockerfile:1

FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o app ./cmd/server

FROM golang:1.24-alpine

WORKDIR /app

COPY --from=builder /app/app /app/app
RUN cp /app/app /usr/local/bin/cms-server && chmod +x /app/app /usr/local/bin/cms-server

ENV APP_ENV=production
ENV PORT=8080

EXPOSE 8080

ENTRYPOINT ["/app/app"]