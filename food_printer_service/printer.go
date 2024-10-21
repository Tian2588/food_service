package main

import (
	"fmt"
	"log"

	nats "github.com/nats-io/nats.go"
)

// Printer subscribes to NATS, if the message from topic is received, and print it to the console.
func Printer(NATSConn *nats.Conn) {
	// Subscribe
	if _, err := NATSConn.Subscribe(NATS_TOPIC, func(msg *nats.Msg) {
		fmt.Println("Got msg from NATS:", msg.Subject, string(msg.Data))
		// wg.Done()
	}); err != nil {
		log.Fatal(err)
	}
}
