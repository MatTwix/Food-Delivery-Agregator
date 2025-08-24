FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY common/go.mod common/go.sum ./common/
COPY scheduler-service/go.mod scheduler-service/go.sum ./scheduler-service/

RUN echo "go 1.24" >> go.work
RUN echo "use (" >> go.work
RUN echo "  ./common" >> go.work
RUN echo "  ./scheduler-service" >> go.work
RUN echo ")" >> go.work

RUN go mod download

COPY common/ ./common/
COPY scheduler-service/ ./scheduler-service/
COPY scheduler-service/config.yaml ./scheduler-service/config.yaml

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main ./scheduler-service/

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /app/scheduler-service/config.yaml ./scheduler-service/

EXPOSE 3003

CMD ["/app/main"]