# Build stage
FROM golang:1.14-alpine as builder

# Create build environment
RUN apk add --no-cache git
WORKDIR /tmp/todos

# Pull dependencies using go mod
COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy the source code and build the executable
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -o ./out/todos ./cmd/todos/

# Final stage
FROM alpine:3.9

LABEL maintainer="benjamin@bengfort.com"
LABEL description="Simple TODO API for personal task tracking"

# Create the execution environment
RUN apk add ca-certificates
RUN adduser -S -D -H -h /app todos
USER todos

# Copy the executable from the build stage
COPY --from=builder /tmp/todos/out/todos /app/todos
WORKDIR /app

# Run the todos server
CMD ["/app/todos", "serve"]
