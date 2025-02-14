FROM golang:1.23.6-alpine as builder

RUN apk update && apk add --no-cache bash

WORKDIR /merch_store_app
COPY . .

COPY migrations/*.sql migrations/
COPY migrations/migrate.sh .
#COPY .env .

RUN go install github.com/pressly/goose/v3/cmd/goose@v3.24.1

RUN chmod +x ./migrate.sh

CMD ["bash", "./migrate.sh"]
