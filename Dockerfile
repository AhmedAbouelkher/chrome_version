FROM golang:1.18.4-alpine3.16

RUN apk add --update curl && \
    rm -rf /var/cache/apk/*

WORKDIR /usr/src/app

COPY .env.example ./.env
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o /usr/local/bin/exec ./main.go

CMD ["exec"]