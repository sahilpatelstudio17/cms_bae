FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o app ./cmd/server

FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/app ./app

ENV APP_ENV=production

EXPOSE 8080

CMD ["/app/app"]