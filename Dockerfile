# --- Stage 1: The Build Stage ---
# Use an official Go image as the base.
# The 'alpine' version is smaller than the default.
FROM golang:1.25-alpine AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy the go.mod and go.sum files first to leverage Docker's build cache.
# This step only re-runs if these files change.
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of your application's source code
COPY . .

# Build the Go application.
# -o /app/firehawk creates an executable named 'firehawk'.
# CGO_ENABLED=0 is important for creating a statically-linked binary that
# can run on minimal base images like alpine.
RUN CGO_ENABLED=0 go build -o /app/firehawk .

# --- Stage 2: The Final Stage ---
# Use a minimal, secure base image. Alpine is a great choice.
FROM alpine:latest

# Set the Current Working Directory inside the final container
WORKDIR /app

# Copy only the compiled executable from the 'builder' stage.
# Also copy the .env file for configuration.
COPY --from=builder /app/firehawk .

COPY esi_cache.json .
COPY systems.json .
COPY .env .

# (Optional) If you were using a local cache file, you would copy it here too.
# COPY esi_cache.json .

# Command to run when the container starts.
CMD ["./firehawk"]