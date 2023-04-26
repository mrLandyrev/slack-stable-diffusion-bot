package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sd-slack-api/internal/models"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Fatalf("unable to open connect to RabbitMQ server. Error: %s", err)
	}
	defer conn.Close()

	channel, err := conn.Channel()
	if err != nil {
		log.Fatalf("failed to open channel. Error: %s", err)
	}
	defer channel.Close()

	q, err := channel.QueueDeclare(
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

	promptsQueue, _ := channel.QueueDeclare(
		"prompts",
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)

	promptsMessages, _ := channel.Consume(
		promptsQueue.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)

	for promptsMessage := range promptsMessages {
		var prompt models.Prompt
		json.Unmarshal(promptsMessage.Body, &prompt)
		fmt.Println(prompt.Text)
		serilizedRequest, _ := json.Marshal(struct {
			Prompt string `json:"prompt"`
			Steps  int    `json:"steps"`
		}{
			Prompt: prompt.Text,
			Steps:  20,
		})
		r, _ := http.NewRequest(http.MethodPost, "http://landyrev.site/sdapi/v1/txt2img", bytes.NewReader(serilizedRequest))
		response, _ := http.DefaultClient.Do(r)
		a := struct {
			Images []string `json:"images"`
		}{
			Images: []string{},
		}

		b, _ := io.ReadAll(response.Body)
		json.Unmarshal(b, &a)

		send := models.Send{
			Image:           a.Images[0],
			Channel:         prompt.Channel,
			ThreadTimestamp: prompt.ThreadTimestamp,
		}

		serializedSend, _ := json.Marshal(send)

		channel.Publish(
			"",     // exchange
			q.Name, // routing key
			false,  // mandatory
			false,  // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        serializedSend,
			},
		)
	}
}
