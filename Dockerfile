# Build (docker compose / deploy)
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /retech-auth-api ./cmd/api

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /retech-auth-api /usr/local/bin/retech-auth-api
COPY public ./public
ENV PORT=8080
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/retech-auth-api"]
