# Dockerfile was generated from
# https://github.com/lodthe/dockerfiles/blob/main/go/Dockerfile

FROM golang:1.23.4-alpine3.21 AS builder

# Setup base software for building an app.
RUN apk update && apk add ca-certificates git gcc g++ libc-dev binutils

WORKDIR /opt

# Download dependencies.
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy application source.
COPY solution .

# Build the application.
RUN go build -o bin/application ./cmd

# Prepare executor image.
FROM alpine:3.21 AS runner

RUN apk update && apk add ca-certificates libc6-compat openssh bash && rm -rf /var/cache/apk/*

WORKDIR /opt

COPY config.yaml /opt
COPY --from=builder /opt/bin/application ./

EXPOSE 3000

# Add required static files.
#COPY assets assets

# Run the application.
CMD ["./application"]
