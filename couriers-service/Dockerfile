FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY common/go.mod common/go.sum ./common/
COPY couriers-service/go.mod couriers-service/go.sum ./couriers-service/

RUN echo "go 1.24" >> go.work
RUN echo "use (" >> go.work
RUN echo "  ./common" >> go.work
RUN echo "  ./couriers-service" >> go.work
RUN echo ")" >> go.work

RUN go mod download

COPY common/ ./common/
COPY couriers-service/ ./couriers-service/

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main ./couriers-service/

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/main .

EXPOSE 3003

CMD ["/app/main"]