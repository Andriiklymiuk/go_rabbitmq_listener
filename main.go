package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rabbitmq/amqp091-go"

	"andriiklymiuk/go_rabbitmq_listener/utils"
)

type MessageData struct {
	Id        string    `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	Message   string    `json:"message"`
}

type QueueMessage struct {
	Data MessageData `json:"data"`
}

func main() {
	envConfig, err := utils.LoadConnectionConfig()
	if err != nil {
		log.Panicln("Couldn't load env variables", err)
	}

	server := &http.Server{Addr: fmt.Sprintf(":%d", envConfig.ServerPort)}
	http.HandleFunc("/status", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Success"))
	})

	go func() {
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server error: %v", err)
		}
		log.Println("Stopped serving new connections.")
	}()

	var connectionType string

	if envConfig.Host == "localhost" {
		connectionType = "amqp"
	} else {
		connectionType = "amqps"
	}

	connectionUri := fmt.Sprintf(
		"%s://%s:%s@%s:%d",
		connectionType,
		envConfig.User,
		envConfig.Password,
		envConfig.Host,
		envConfig.Port,
	)

	var queueConnection utils.AmqpConnection
	queueConnection.ConnectionUri = connectionUri
	queueConnection.QueueName = "QUEUE_SERVICE"
	queueConnection.OnMessageReceived = func(rawMessage amqp091.Delivery) {
		go func() {
			var message QueueMessage
			err := json.Unmarshal(rawMessage.Body, &message)
			if err != nil {
				fmt.Println(utils.RedColor, err, utils.WhiteColor)
				return
			}
			fmt.Println("You received message: ", message.Data.Message)
		}()
	}

	queueConnection.EstablishConnection(nil)

	gracefulShutdown := make(chan os.Signal, 1)
	signal.Notify(gracefulShutdown, syscall.SIGINT, os.Interrupt, syscall.SIGTERM)

	<-gracefulShutdown

	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownRelease()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("HTTP shutdown error: %v", err)
	}

	fmt.Println("\n🚀 Bye, see you next time 🚀!")
}
