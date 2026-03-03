FROM golang:1.25.7-alpine AS build
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
RUN go mod download

# copy source files
COPY . .

# Download dependency using go mod
RUN go mod download

# Build the application
WORKDIR /app/cmd/go-rest
#RUN go mod download
RUN go build -o /app/bin/go-rest .

# Run stage
FROM alpine:latest
WORKDIR /app
COPY --from=build /app/bin/go-rest .

# Command to run when starting the container
CMD ["go-rest"]