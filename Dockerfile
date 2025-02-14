FROM golang:1.23.6-alpine as builder

WORKDIR /merch_store_app

COPY go.mod go.sum ./
RUN go mod tidy

COPY . .

RUN GOOS=linux GOARCH=amd64 go build -o merch_store_linux cmd/store/main.go

FROM alpine:3.20.5

WORKDIR /merch_store_app

COPY --from=builder /merch_store_app/merch_store_linux ./merch_store_linux
COPY --from=builder /merch_store_app/config/config.yaml ./config/config.yaml

CMD ["./merch_store_linux"]
