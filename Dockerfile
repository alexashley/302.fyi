FROM golang:1.16.0-alpine3.13 as builder

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download
COPY main.go config.yaml ./
RUN go build -o 302.fyi

###

FROM alpine:3.13

WORKDIR /usr/src/app

COPY --from=builder /usr/src/app/302.fyi .

ENTRYPOINT ["./302.fyi"]
