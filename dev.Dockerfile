FROM golang:alpine

# Install build dependencies
RUN apk add --no-cache make

# Set the Current Working Directory inside the container
WORKDIR /app

# Install air for live reload on save
RUN go install github.com/cosmtrek/air@latest

# Copy the Go modules and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy everything else to the container
COPY . .

# Build the Go app
RUN make build/api

# Expose port 4000 to the outside world
EXPOSE 4000

# Run the produced binary with air
CMD ["air", "-c", ".air.toml"]
