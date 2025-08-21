FROM golang:1.24-alpine AS builder

WORKDIR /app


COPY notification-service/go.mod notification-service/go.sum ./notification-service/

RUN echo "go 1.24" >> go.work
RUN echo "use (" >> go.work
RUN echo "  ./notification-service" >> go.work
RUN echo ")" >> go.work

RUN go mod download

COPY notification-service/ ./notification-service/
COPY notification-service/config.yaml ./notification-service/config.yaml

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main ./notification-service/

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /app/notification-service/config.yaml ./notification-service/config.yaml

CMD ["/app/main"]