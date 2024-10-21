package main

import (
	"fmt"
	"testing"

	natsserver "github.com/nats-io/nats-server/v2/test"
	nats "github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
)

func TestPrinter(t *testing.T) {
	srv := natsserver.RunDefaultServer()
	defer srv.Shutdown()

	url := fmt.Sprintf("nats://127.0.0.1:%d", nats.DefaultPort)
	nc, err := nats.Connect(url)
	if err != nil {
		t.Fatalf("failed to create default connection: %v", err)
	}
	defer nc.Close()
	assert.NotNil(t, nc)

	// test data
	testMsg := `{"name":"banaba","number",1}`
	err = nc.Publish(NATS_TOPIC, []byte(testMsg))
	if err != nil {
		t.Errorf("failed to publish msg to NATS: %v", err)
	}
	// Subscribe and check if received msg is equal to published msg
	if _, err := nc.Subscribe(NATS_TOPIC, func(msg *nats.Msg) {
		fmt.Println("Got msg from NATS:", msg.Subject, string(msg.Data))
		assert.Equal(t, testMsg, msg.Data)
	}); err != nil {
		t.Fatal(err)
	}
}
