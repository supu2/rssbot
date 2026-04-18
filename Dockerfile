FROM golang:1.26.2-alpine3.23

RUN apk add --no-cache git sqlite-dev gcc musl-dev

WORKDIR /app

COPY src/go.mod src/go.sum ./
RUN go mod download

COPY src/. .

RUN CGO_ENABLED=1 go build -o rss-bot .

FROM alpine:3.23
WORKDIR /app
COPY --from=0 /app/rss-bot .
CMD ["rss-bot"]