FROM golang:1.25 AS builder

WORKDIR /app

COPY . .

RUN make setup
RUN make build-ci

FROM alpine:3.22.1

WORKDIR /app

COPY --from=builder /app/main .

EXPOSE 8000

CMD ["./main"]