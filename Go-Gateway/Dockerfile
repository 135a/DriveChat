# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies (using a proxy if needed, e.g., GOPROXY=https://goproxy.cn,direct)
RUN go mod download

# Copy the source code
COPY . .

# Build the application with optimized flags:
# -s: strip symbol table
# -w: strip DWARF debug information
# CGO_ENABLED=0: static compilation for scratch/alpine compatibility
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o main ./cmd/main.go

# Final stage
FROM alpine:latest

WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/main .

# Expose the port the app runs on
EXPOSE 8080

# Command to run the executable
CMD ["./main"]
