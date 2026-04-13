FROM golang:1.24-alpine

WORKDIR /app

COPY . .

RUN go mod tidy
RUN go build -o app ./cmd/api

EXPOSE 8080

CMD ["./app"]