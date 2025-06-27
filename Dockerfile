FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o api_gateway ./cmd

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/api_gateway .

EXPOSE 8080

CMD ["./api_gateway"]