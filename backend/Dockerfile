# syntax=docker/dockerfile:1

FROM golang:1.22-alpine AS build
WORKDIR /src

RUN apk add --no-cache ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/gopayments .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates wget

COPY --from=build /out/gopayments /usr/local/bin/gopayments

ENV LISTEN_ADDR=:8080
EXPOSE 8080

HEALTHCHECK --interval=15s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- "http://127.0.0.1:8080/api/stats" >/dev/null || exit 1

USER nobody
ENTRYPOINT ["/usr/local/bin/gopayments"]
