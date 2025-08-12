FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY common/go.mod common/go.sum ./common/
COPY orders-service/go.mod orders-service/go.sum ./orders-service/

RUN echo "go 1.24" >> go.work
RUN echo "use (" >> go.work
RUN echo "  ./common" >> go.work
RUN echo "  ./orders-service" >> go.work
RUN echo ")" >> go.work

RUN go mod download

COPY common/ ./common/
COPY orders-service/ ./orders-service/
COPY orders-service/config.yaml ./orders-service/config.yaml

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main ./orders-service

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /app/orders-service/config.yaml ./orders-service/

EXPOSE 3002

CMD ["/app/main"]