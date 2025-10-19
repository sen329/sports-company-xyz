# Stage 1: Build the Go application
FROM golang:1.21-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum to leverage Docker's layer caching.
# This step will only be re-run if these files change.
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application as a static binary
# CGO_ENABLED=0 is important for creating a static binary that can run in a minimal container like alpine.
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/main ./main.go

# Stage 2: Create the final, minimal image
FROM alpine:latest

WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /app/main .

# Your application uses a .env file. Copy it into the image.
# For production, it's often better to manage these as environment variables passed to the container.
COPY .env .

# Expose the port the application will run on (ensure this matches the PORT in your .env file)
EXPOSE 8080

# The command to run the application
CMD ["./main"]