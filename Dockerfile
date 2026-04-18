FROM golang:1.26.2-alpine3.23

RUN apk add --no-cache git sqlite-dev gcc musl-dev

WORKDIR /app

COPY src/go.mod src/go.sum ./
RUN go mod download

COPY src/. .

RUN CGO_ENABLED=1 go build -o rssbot .

FROM alpine:3.23
COPY --from=0 /app/rssbot /usr/local/bin/rssbot
CMD ["rssbot"]