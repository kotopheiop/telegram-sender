FROM golang:alpine as builder

WORKDIR /build
ADD go.mod .
ADD go.sum .
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/telegram-sender ./main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates curl

RUN mkdir /app
WORKDIR /app
COPY --from=builder /build/.env /app/.env
COPY --from=builder /app/telegram-sender /app/telegram-sender

CMD ["./telegram-sender"]