package main

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	if err := godotenv.Load(); err != nil {
		panic(err)
	}

	amqpServerURL := os.Getenv("AMQP_SERVER_URL")

	// Create a new RabbitMQ connection.
	connectRabbitMQ, err := amqp.Dial(amqpServerURL)
	if err != nil {
		panic(err)
	}
	defer connectRabbitMQ.Close()

	// Let's start by opening a channel to our RabbitMQ
	// instance over the connection we have already
	// established.
	channelRabbitMQ, err := connectRabbitMQ.Channel()
	if err != nil {
		panic(err)
	}
	defer channelRabbitMQ.Close()

	// With the instance and declare Queues that we can
	// publish and subscribe to.
	_, err = channelRabbitMQ.QueueDeclare(
		"QueueService1", // queue name
		true,            // durable
		false,           // auto delete
		false,           // exclusive
		false,           // no wait
		nil,             // arguments
	)
	if err != nil {
		panic(err)
	}

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
			"",              // exchange
			"QueueService1", // queue name
			false,           // mandatory
			false,           // immediate
			message,         // message to publish
		); err != nil {
			c.JSON(http.StatusOK, gin.H{"err": err.Error()})
			return
		}

		c.Status(http.StatusOK)
	})

	if err := router.Run(":3000"); err != nil {
		panic(err)
	}
}
