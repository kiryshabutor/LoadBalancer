FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN go build -o backend ./cmd/backend-mock

FROM alpine:3.19

RUN apk --no-cache add ca-certificates

RUN adduser -D -g '' appuser

WORKDIR /app

COPY --from=builder /app/backend /app/backend

USER appuser

EXPOSE 8081 8082 8083

CMD ["./backend"]