FROM golang:1.26.8-alpine AS build
LABEL org.opencontainers.image.source="https://github.com/ric7pt/golang-rest-api"

# Set necessary environmet variables needed for our image
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Set the working directory inside the container
WORKDIR /app

# Copy dependency files first (better layer caching)
COPY go.mod go.sum ./

# Download dependency using go mod
RUN go mod download

# copy source files
COPY . .

# Build the application
WORKDIR /app/cmd/go-rest
RUN go build -o /app/bin/go-rest .

# multi-stage - Run stage 2 (grabs only the built binary)
FROM alpine:latest
WORKDIR /app
COPY --from=build /app/bin/go-rest .

# Listen on all interfaces so the server is reachable from outside the container
ENV HTTP_SERVER_HOST=0.0.0.0

# Command to run when starting the container
CMD ["./go-rest"]