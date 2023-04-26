FROM golang:1.20

WORKDIR /var/app
COPY ./src /var/app
RUN go mod download
RUN go build -o webhook cmd/webhook/main.go
ENTRYPOINT ./webhook