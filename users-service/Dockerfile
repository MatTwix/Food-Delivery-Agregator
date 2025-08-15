FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY common/go.mod common/go.sum ./common/
COPY users-service/go.mod users-service/go.sum ./users-service/

RUN echo "go 1.24" >> go.work
RUN echo "use (" >> go.work
RUN echo "  ./common" >> go.work
RUN echo "  ./users-service" >> go.work
RUN echo ")" >> go.work

RUN go mod download

COPY common/ ./common/
COPY users-service/ ./users-service/
COPY users-service/config.yaml ./users-service/config.yaml

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main ./users-service

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/main .
COPY --from=builder /app/users-service/config.yaml ./users-service/

EXPOSE 3004

CMD ["/app/main"]