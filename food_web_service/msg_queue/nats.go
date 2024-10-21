package msgqueue

import (
	"os"
	"time"

	"github.com/labstack/gommon/log"
	nats "github.com/nats-io/nats.go"
)

const NATS_TOPIC = "food"

func ConnectNATS() (*nats.Conn, error) {
	uri := os.Getenv("NATS_URI")
	var err error
	var natsConn *nats.Conn

	for i := 0; i < 3; i++ {
		natsConn, err = nats.Connect(uri)
		if err == nil {
			break
		}

		log.Infof("waiting before connecting to NATS at: %s", uri)
		time.Sleep(1 * time.Second)
	}
	if err != nil {
		log.Errorf("error establishing connection to NATS: ", err)
		return nil, err
	}
	log.Infof("connected to NATS at: %s", natsConn.ConnectedUrl())
	return natsConn, nil
}
