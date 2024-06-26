# Builder Stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy and download dependencies separately to leverage Docker caching
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy the rest of the application code
COPY . .

# Build the application
RUN go build -o /app/user-svc

# Final Stage
FROM alpine:latest

# Copy the built binary from the builder stage
COPY --from=builder /app/user-svc .

# Explicitly copy the configuration file into the container
# COPY ./config ./config
COPY config.yaml ./
COPY infrastructure/db/migrations ./

# Set executable permissions for the binary
# RUN chmod +x user-svc

# Expose port
EXPOSE 5000

# Command to run the application
CMD ["./user-svc"]
