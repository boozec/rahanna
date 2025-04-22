FROM golang:alpine AS builder

RUN apk --update add ca-certificates git

WORKDIR /app

COPY . .
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build cmd/api/main.go

# Run the exe file
FROM scratch

LABEL version="0.1.0"
WORKDIR /app

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /app .

EXPOSE 8080

CMD ["./main"]
