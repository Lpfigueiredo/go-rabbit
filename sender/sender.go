package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func main() {
	err := godotenv.Load()
	failOnError(err, "Failed to load envs")

	amqpServerURL := os.Getenv("AMQP_SERVER_URL")

	// Create a new RabbitMQ connection.
	connectRabbitMQ, err := amqp.Dial(amqpServerURL)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer connectRabbitMQ.Close()

	// Let's start by opening a channel to our RabbitMQ
	// instance over the connection we have already
	// established.
	channelRabbitMQ, err := connectRabbitMQ.Channel()
	failOnError(err, "Failed to open a channel")
	defer channelRabbitMQ.Close()

	// With the instance and declare Queues that we can
	// publish and subscribe to.
	q, err := channelRabbitMQ.QueueDeclare(
		"QueueService1", // queue name
		true,            // durable
		false,           // auto delete
		false,           // exclusive
		false,           // no wait
		nil,             // arguments
	)
	failOnError(err, "Failed to declare a queue")

	router := gin.Default()

	router.GET("/send", func(c *gin.Context) {
		message := amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(c.Query("msg")),
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Attempt to publish a message to the queue.
		if err := channelRabbitMQ.PublishWithContext(
			ctx,
			"",      // exchange
			q.Name,  // queue name
			false,   // mandatory
			false,   // immediate
			message, // message to publish
		); err != nil {
			c.JSON(http.StatusOK, gin.H{"err": err.Error()})
			return
		}

		c.Status(http.StatusOK)
	})

	e := router.Run(":3000")
	failOnError(e, "Failed to init router")
}
