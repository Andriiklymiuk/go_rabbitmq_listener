package utils

import (
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type AmqpConnection struct {
	ConnectionUri     string
	QueueName         string
	Conn              *amqp.Connection
	Channel           *amqp.Channel
	Queue             amqp.Queue
	Err               chan error
	OnMessageReceived func(amqp.Delivery)
}

const (
	// When reconnecting to the server after connection failure
	reconnectDelay = 5 * time.Second
)

func (c *AmqpConnection) Connect() error {
	conn, err := amqp.Dial(c.ConnectionUri)
	if err != nil {
		return fmt.Errorf("failed to connect to amqp: %s", err)
	}
	c.Conn = conn

	go func() {
		errConnection := <-c.Conn.NotifyClose(make(chan *amqp.Error))
		go c.EstablishConnection(errConnection)
	}()

	channel, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open a channel: %s", err)
	}
	c.Channel = channel
	return nil
}

func (c *AmqpConnection) EstablishConnection(errConnection *amqp.Error) {
	if errConnection != nil {
		log.Println(YellowColor, "Trying to reconnect, because of", errConnection, BlueColor)
	}
	for !c.ConnectAndSubscribe() {
		log.Println(RedColor, "Failed to connect. Retrying...", WhiteColor)
		time.Sleep(reconnectDelay)
	}
	messages, err := c.SubscribeToMessages()
	if err != nil {
		log.Panicln(RedColor, "Couldn't createAmqpConnection", err, WhiteColor)
	}

	go func() {
		for message := range messages {
			go c.OnMessageReceived(message)
		}
	}()

	fmt.Println(
		BlueColor,
		"Connection to rabbitmq is established, listening to messages...",
		WhiteColor,
	)
}

func (c *AmqpConnection) SubscribeToMessages() (<-chan amqp.Delivery, error) {
	messages, err := c.Channel.Consume(
		c.Queue.Name, // queue
		"",           // consumer
		true,         // auto-ack
		false,        // exclusive
		false,        // no-local
		false,        // no-wait
		nil,          // args
	)

	if err != nil {
		return nil, fmt.Errorf(RedColor, "failed to register a consumer: %s", err, WhiteColor)
	}
	return messages, nil
}

func (c *AmqpConnection) ConnectAndSubscribe() bool {
	err := c.Connect()
	if err != nil {

		fmt.Println(RedColor, err, WhiteColor)
		return false
	}
	queue, err := c.Channel.QueueDeclare(
		c.QueueName, // name
		true,        // durable
		false,       // delete when unused
		false,       // exclusive
		false,       // no-wait
		nil,         // arguments
	)
	if err != nil {
		fmt.Println(RedColor, "failed to declare a queue", err, WhiteColor)
		return false
	}
	c.Queue = queue
	return true
}
