FROM golang:1.25.3-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o sfsEdgeStore .

FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/sfsEdgeStore .
COPY --from=builder /app/config.example.json config.json

EXPOSE 8081

VOLUME ["/app/edgex_data", "/app/data_sync_queue", "/app/backups"]

ENV TZ=UTC

CMD ["./sfsEdgeStore"]
