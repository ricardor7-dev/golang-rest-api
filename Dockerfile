FROM golang:1.25.7-alpine AS build
LABEL org.opencontainers.image.source="https://github.com/ric7pt/golang-rest-api"

# Set necessary environmet variables needed for our image
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Move to working directory
COPY . .

# Copy and download dependency using go mod
WORKDIR ./src/
RUN go mod download


# Build the application
WORKDIR ./cmd/go-rest
#RUN go mod download
RUN go build all
RUN go build

WORKDIR ../../../bin/

# Copy binary from build to main folder
RUN cp ../src/cmd/go-rest/go-rest* .

# Command to run when starting the container
CMD ["go-rest"]
