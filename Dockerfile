# Start with the official Go image to build the binary
FROM golang:alpine AS builder

# Install gcc and g++ for building C/C++ dependencies
RUN apk add --no-cache gcc g++

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the go.mod and go.sum files
COPY go.mod go.sum ./

# Download the dependencies
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go app
RUN go build -o zenbpm cmd/zenbpm/main.go

# Start a new stage from scratch
FROM alpine:latest  

# Set the Current Working Directory inside the container
WORKDIR /root/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/zenbpm .

# Expose port 8080 and 4001 to the outside world
EXPOSE 8080 4001

# Command to run the executable
CMD ["./zenbpm"]
