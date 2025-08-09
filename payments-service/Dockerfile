FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY payments-service/go.mod payments-service/go.sum ./payments-service/

RUN echo "go 1.24" >> go.work
RUN echo "use (" >> go.work
RUN echo "  ./payments-service" >> go.work
RUN echo ")" >> go.work

RUN go mod download

COPY payments-service/ ./payments-service/

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main ./payments-service/

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/main .

CMD ["/app/main"]