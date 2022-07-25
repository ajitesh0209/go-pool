FROM golang:1.18-alpine

WORKDIR /app

COPY go.mod ./
COPY *.go ./
RUN go mod tidy