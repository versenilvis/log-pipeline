FROM golang:alpine AS builder

ARG SERVICE_PATH

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/service ${SERVICE_PATH}

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/bin/service .

ENTRYPOINT ["./service"]
