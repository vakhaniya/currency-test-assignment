FROM golang:1.25 AS builder

WORKDIR /app

COPY . .

RUN make setup
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd

FROM alpine:3.22.1

WORKDIR /app

COPY --from=builder /app/main .

EXPOSE 8000

CMD ["./main"]