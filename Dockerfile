# Start from the official Golang image
FROM golang:1.22-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code into the container
COPY *.go ./

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o pvc-usage .

# Start a new stage from scratch
FROM alpine:latest  

# Install ca-certificates for HTTPS requests to the Kubernetes API server
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the pre-built binary file from the previous stage
COPY --from=builder /app/pvc-usage .

# Command to run the executable
ENTRYPOINT ["./pvc-usage"]