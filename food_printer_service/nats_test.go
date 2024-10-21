package main

import (
	"fmt"
	"testing"

	natsserver "github.com/nats-io/nats-server/v2/test"
	nats "github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
)

func TestConnectNATS(t *testing.T) {
	srv := natsserver.RunDefaultServer()
	defer srv.Shutdown()

	url := fmt.Sprintf("nats://127.0.0.1:%d", nats.DefaultPort)
	nc, err := nats.Connect(url)
	if err != nil {
		t.Fatalf("failed to create default connection: %v", err)
	}
	defer nc.Close()
	assert.NotNil(t, nc)
}
