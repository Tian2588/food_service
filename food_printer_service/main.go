package main

import (
	"log"
	"net/http"

	llog "github.com/labstack/gommon/log"
)

func main() {
	NATSConn, err := ConnectNATS()
	if err != nil {
		llog.Error("[Printer] failed to get NATS conn, err: %s" + err.Error())
		return
	}
	llog.Info("Printer Servie subscribed to 'food' for processing messages...")
	Printer(NATSConn)

	if err := http.ListenAndServe(":8181", nil); err != nil {
		log.Fatal(err)
	}
}
