FROM golang:alpine as builder

WORKDIR /app

COPY . .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o rinha-api


FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/rinha-api .

CMD ["./rinha-api"]