package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"food_web_service/db"
	msgqueue "food_web_service/msg_queue"
	"testing"

	"github.com/labstack/gommon/log"
	natsserver "github.com/nats-io/nats-server/v2/test"
	nats "github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
)

func TestProduceFood(t *testing.T) {
	food := ProduceFoodRequest{
		Name:   "apple",
		Number: 1,
	}
	// connect database and store data
	sqliteDB, err := db.CreateConnection("./data/food.db")
	if err != nil {
		t.Fatal("failed to connect SQLite, err: " + err.Error())
	}
	defer sqliteDB.Close()
	id, err := db.UpsertFood(sqliteDB, food.Name, food.Number)
	if err != nil {
		t.Error("failed to store in SQLite, err: " + err.Error())
	}
	// connect nats
	srv := natsserver.RunDefaultServer()
	defer srv.Shutdown()
	url := fmt.Sprintf("nats://127.0.0.1:%d", nats.DefaultPort)
	nc, err := nats.Connect(url)
	if err != nil {
		t.Fatalf("failed to create default connection: %v", err)
	}
	defer nc.Close()
	assert.NotNil(t, nc)
	// publish msg to nats
	msgBytes, err := json.Marshal(food)
	if err != nil {
		t.Errorf("marshal err: %v", err)
	}
	err = nc.Publish(msgqueue.NATS_TOPIC, msgBytes)
	if err != nil {
		t.Errorf("failed to publish message to NATS err: %v", err)
	} else {
		// set identifier in SQLite database
		err = db.UpdateIdentifierByID(sqliteDB, id, db.IDENTIFIER_BE_SENT_NATS)
		if err != nil {
			t.Errorf("failed to update identifier in SQLite, err: %v", err)
		}
	}
	// check if subscribed data is the same with the published
	if _, err := nc.Subscribe(msgqueue.NATS_TOPIC, func(msg *nats.Msg) {
		assert.Equal(t, string(msgBytes), msg.Data)
	}); err != nil {
		t.Fatal(err)
	}
}

func TestConsumeFood(t *testing.T) {
	sqliteDB, err := db.CreateConnection("./data/food.db")
	if err != nil {
		t.Errorf("create database connection error = %v", err)
	}
	defer sqliteDB.Close()

	// test1: should return initial data
	foodName := "apple"
	foodNum := int64(1)
	err = db.CreateTableFoodIfNotExists(sqliteDB)
	assert.Nil(t, err)
	err = initSQLiteTableData(sqliteDB, foodName, foodNum)
	if err != nil {
		t.Errorf("initialize data for database error = %v", err)
	}
	// retrieve the database
	food, err := db.GetLatestFood(sqliteDB)
	if err != nil {
		t.Errorf("failed to get latest food from database error = %v", err)
	}
	ids := make([]int64, len(food))
	for i, f := range food {
		// check if latest retrieved food is the same with inital data
		assert.Equal(t, foodName, f.Name)
		assert.Equal(t, foodNum, f.Number)
		ids[i] = f.ID
	}
	// set identifier in the database
	err = db.UpdateIdentifierByIDList(sqliteDB, ids, db.IDENTIFIER_ALREADY_GET)
	if err != nil {
		t.Errorf("failed to update indentifier from database, id list: %v, error = %v", ids, err)
	}

	// test2: should return empty
	food, err = db.GetLatestFood(sqliteDB)
	if err != nil {
		t.Errorf("failed to get latest food from database error = %v", err)
	}
	assert.Equal(t, 0, len(food))
}

func initSQLiteTableData(db *sql.DB, name string, num int64) error {
	query := ` INSERT INTO food(name, number) VALUES(?, ?);`
	_, err := db.Exec(query, name, num)
	if err != nil {
		log.Errorf("[Test] failed to exec create table food, err: %v", err)
		return err
	}
	log.Info("[Test] successfully to create table food if not exists")
	return nil
}
