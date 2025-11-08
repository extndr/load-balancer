FROM golang:1.25.1-alpine AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o loadbalancer cmd/app/main.go

FROM alpine:latest
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /build/loadbalancer .
EXPOSE 8080
CMD ["./loadbalancer"]
