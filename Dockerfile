FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . /app

RUN go build -o main ./cmd/main.go

# Use a smaller, production-ready base image
FROM alpine:latest

# Install necessary dependencies including timezone data
RUN apk update && apk add --no-cache \
    ca-certificates \
    openssl \
    bind-tools \
    tzdata

WORKDIR /app

# Copy the compiled Go binary from the builder stage
COPY --from=builder /app/main ./

# Copy the .env file into the container (optional if required by your app)
COPY .env .env

# Set the correct timezone (optional)
ENV TZ=Asia/Tashkent

EXPOSE 9090

CMD ["./main"]
