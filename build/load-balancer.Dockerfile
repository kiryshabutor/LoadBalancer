FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .

RUN go build -o load-balancer ./cmd/lb

FROM alpine:3.19

RUN apk --no-cache add ca-certificates

RUN adduser -D -g '' appuser

WORKDIR /app

COPY --from=builder /app/load-balancer /app/load-balancer

USER appuser

EXPOSE 8080

CMD ["./load-balancer"]
