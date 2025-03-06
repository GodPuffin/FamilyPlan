FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o familyplan .

# Use a smaller image for the final container
FROM alpine:latest

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/familyplan .

# Copy templates and static files
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static

# Set environment variables to prevent HTTPS redirects
ENV DISABLE_HTTPS=true

# Expose the port
EXPOSE 8090

# Run the application with explicit binding to all interfaces and HTTPS disabled
CMD ["./familyplan", "serve", "--http=0.0.0.0:8090"] 