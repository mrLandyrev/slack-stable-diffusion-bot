package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"log"
	"sd-slack-api/internal/models"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/slack-go/slack"
)

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("unable to open connect to RabbitMQ server. Error: %s", err)
	}

	defer func() {
		_ = conn.Close() // Закрываем подключение в случае удачной попытки подключения
	}()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("failed to open a channel. Error: %s", err)
	}

	defer func() {
		_ = ch.Close() // Закрываем подключение в случае удачной попытки подключения
	}()

	q, err := ch.QueueDeclare(
		"images", // name
		false,    // durable
		false,    // delete when unused
		false,    // exclusive
		false,    // no-wait
		nil,      // arguments
	)
	if err != nil {
		log.Fatalf("failed to declare a queue. Error: %s", err)
	}
	messages, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		log.Fatalf("failed to register a consumer. Error: %s", err)
	}
	for message := range messages {
		var send models.Send
		json.Unmarshal(message.Body, &send)
		b64, _ := base64.StdEncoding.DecodeString(send.Image)
		client := slack.New("paster token here")
		client.UploadFile(slack.FileUploadParameters{Reader: bytes.NewReader(b64), Filename: "image.png", Filetype: "image/png", Channels: []string{send.Channel}, ThreadTimestamp: send.ThreadTimestamp})
	}
}
