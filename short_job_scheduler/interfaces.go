package main

import (
 "github.com/streadway/amqp"
)

// Communication interface contains the methods that are required
type Communication interface {
 Send(message []byte, queue string) error
 Receive(queue string) (<-chan amqp.Delivery, error)
 Connect() error
}
