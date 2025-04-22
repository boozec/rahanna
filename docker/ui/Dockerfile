# Build stage
FROM golang:alpine AS builder

RUN apk --update add ca-certificates git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o rahanna-ui cmd/ui/main.go

# Run stage
FROM alpine:latest

LABEL version="0.1.0"
WORKDIR /app

COPY --from=builder /app/rahanna-ui .

CMD ["./rahanna-ui"]
