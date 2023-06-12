FROM golang:alpine

# Set the Current Working Directory inside the container
WORKDIR /app

# Install air for live server reload on save
RUN go install github.com/cosmtrek/air@latest

# Copy the Go modules and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy everything else to the container
COPY . .

# Build the Go app
RUN go build -ldflags='-s' -o=./bin/api ./cmd/api

# Expose port 4000 to the outside world
EXPOSE 4000

# Run the produced binary with air
CMD ["air", "-c", ".air.toml"]
