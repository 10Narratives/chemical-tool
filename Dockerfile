FROM golang:1.23

WORKDIR /usr/src/chemical-tool

COPY . .

RUN go mod tidy