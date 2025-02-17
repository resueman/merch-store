FROM golang:1.23.6-alpine as builder

RUN apk update && apk add --no-cache bash

WORKDIR /test_migrations

COPY test/integration/migrations/*.sql migrations/
COPY test/integration/migrations/migrate.sh .

RUN go install github.com/pressly/goose/v3/cmd/goose@v3.24.1

RUN chmod +x ./migrate.sh

CMD ["bash", "./migrate.sh"]
