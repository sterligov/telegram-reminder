FROM golang:1.13 as build

RUN mkdir /app

WORKDIR /app

COPY go.mod /app
COPY go.sum /app

RUN go mod download

COPY . /app

ENV GOOS=linux GOARCH=amd64

RUN go build -o tgreminder .

ENTRYPOINT ["./tgreminder"]
