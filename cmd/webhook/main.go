package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sd-slack-api/internal/models"
	"strings"

	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	router := gin.New()
	router.POST("/", handler)
	http.ListenAndServe("0.0.0.0:8080", router)
}

// "event": {
// 	"type": "message",
// 	"channel": "D024BE91L",
// 	"user": "U2147483697",
// 	"text": "Hello hello can you hear me?",
// 	"ts": "1355517523.000005",
// 	"event_ts": "1355517523.000005",
// 	"channel_type": "im"
// },

func handler(context *gin.Context) {
	type body struct {
		Event struct {
			Channel string `json:"channel"`
			Text    string `json:"text"`
			User    string `json:"user"`
			Ts      string `json:"ts"`
		} `json:"event"`
	}

	var b body

	context.BindJSON(&b)

	fmt.Println(b.Event.Text, b.Event.User)

	context.Status(http.StatusOK)

	if b.Event.User == "U0548DK7W07" {
		return
	}

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
		"prompts", // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		log.Fatalf("failed to declare a queue. Error: %s", err)
	}

	prompt := models.Prompt{
		Text:            strings.Replace(b.Event.Text, "<@U0548DK7W07> ", "", -1),
		Channel:         b.Event.Channel,
		ThreadTimestamp: b.Event.Ts,
	}

	p, _ := json.Marshal(prompt)

	channel.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        p,
		},
	)
}
