FROM golang:1.24-alpine AS builder

# Install build dependencies for CGO and SQLite
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Enable CGO for SQLite
ENV CGO_ENABLED=1
RUN go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates sqlite
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main", "server"] 