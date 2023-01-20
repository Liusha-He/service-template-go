FROM golang:1.15-alpine as builder

RUN mkdir /app

COPY . /app

WORKDIR /app

RUN CGO_ENABLED=0 go build -o mailApp ./cmd/api

RUN chmod +x /app/mailApp

FROM alpine:latest

RUN mkdir /app

COPY --from=builder /app/mailApp /app
COPY templates /templates

CMD ["/app/mailApp"]
