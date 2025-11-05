# Stage 1: Build the Go main application
FROM golang:1.20-alpine AS builder
WORKDIR /app

# If you have multiple files, copy the entire source
COPY . .

# Build the Go application
RUN make release

# Set up the final image
FROM python:3.11-slim

# Install supervisor for process management
RUN apt-get update && apt-get install -y --no-install-recommends \
    supervisor \
 && rm -rf /var/lib/apt/lists/*

# Copy the built Go binary
COPY --from=builder /app/main /usr/local/bin/main

# Create directories for microservices
WORKDIR /app
COPY microservices/ ./microservices

# Install Python dependencies for each microservice
RUN for dir in microservices/*/; do \
      if [ -f "$dir/requirements.txt" ]; then \
        pip install --no-cache-dir -r "$dir/requirements.txt"; \
      fi \
    done

# Copy supervisor configuration
COPY supervisord.conf /etc/supervisor/conf.d/supervisord.conf

# Expose port(s) if needed
EXPOSE 8080

# Set entrypoint
CMD ["/usr/bin/supervisord", "-n"]
