FROM golang:1.24-alpine AS builder

RUN mkdir /app

ADD . /app/

# Set the working directory inside the container
WORKDIR /app

# Copy the go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code from your host to your image filesystem.
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/main.go

# Use a minimal alpine image for the final stage
FROM alpine:latest

# Set the working directory inside the container
WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/main .

ARG COMMIT
ENV SERVER_SERVICE_COMMIT=${COMMIT}
ARG BUILD_TIME
ENV SERVER_SERVICE_BUILD_TIME=${BUILD_TIME}
ARG RELEASE
ENV SERVER_SERVICE_RELEASE=${RELEASE}
ARG PROJECT
ENV SERVER_SERVICE_PROJECT=${PROJECT}

ARG SERVER_SERVICE_HOST=${SERVER_SERVICE_HOST}
ENV SERVER_SERVICE_HOST=${SERVER_SERVICE_HOST}

ENV PORT=8085
EXPOSE ${PORT}

# Run the binary
CMD ["./main"]
