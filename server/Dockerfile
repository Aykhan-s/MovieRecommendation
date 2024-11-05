FROM golang:1.23.2-alpine AS builder

WORKDIR /server

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go build -ldflags "-s -w" -o movier

FROM alpine:latest

WORKDIR /server

COPY --from=builder /server/movier /server/movier
COPY --from=builder /server/pkg/templates /server/pkg/templates