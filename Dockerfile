FROM golang:1.25

# Install libvips (vipsthumbnail)
RUN apt-get update && apt-get install -y libvips-tools && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build
RUN go build -o /usr/local/bin/estelled ./cmd/estelled
