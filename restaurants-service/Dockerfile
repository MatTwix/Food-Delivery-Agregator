FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY common/go.mod common/go.sum ./common/
COPY restaurants-service/go.mod restaurants-service/go.sum ./restaurants-service/

RUN echo "go 1.24" >> go.work
RUN echo "use (" >> go.work
RUN echo "  ./common" >> go.work
RUN echo "  ./restaurants-service" >> go.work
RUN echo ")" >> go.work

RUN go mod download

COPY common/ ./common/
COPY restaurants-service/ ./restaurants-service/
COPY restaurants-service/config.yaml ./restaurants-service/config.yaml

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main ./restaurants-service/

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /app/restaurants-service/config.yaml ./restaurants-service/

EXPOSE 3001

CMD ["/app/main"]
