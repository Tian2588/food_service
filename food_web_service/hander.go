package main

import (
	"encoding/json"
	"food_web_service/db"
	msgqueue "food_web_service/msg_queue"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type ProduceFoodRequest struct {
	Name   string `json:"name"`
	Number int64  `json:"number"`
}

func ProduceFood(c echo.Context) error {
	food := ProduceFoodRequest{}
	err := c.Bind(&food)
	if err != nil {
		msg := "[ProduceFood] food request: %v, bind error: %v"
		log.Errorf(msg, food, err)
		return c.String(http.StatusBadRequest, "bad request")
	}

	// check and validate the value of input data
	if food.Name == "" {
		msg := "[ProduceFood] the name should not be empty"
		log.Error(msg)
		return c.String(http.StatusBadRequest, msg)
	}
	if food.Number <= 0 {
		msg := "[ProduceFood] the number should not be less than 0"
		log.Error(msg)
		return c.String(http.StatusBadRequest, msg)
	}

	// first store it in SQLite datebase, if stored successfully, then publish it to NATS;
	// if failed or succeeded to publish message, then update its correspond identifier in SQLite
	sqliteDB, err := db.CreateConnection(db.SQLITE_FILE_NAME)
	if err != nil {
		msg := "[ProduceFood] failed to connect SQLite, err: " + err.Error()
		log.Error(msg)
		return c.String(http.StatusInternalServerError, msg)
	}
	defer sqliteDB.Close()

	id, err := db.UpsertFood(sqliteDB, food.Name, food.Number)
	if err != nil {
		msg := "[ProduceFood] failed to store in SQLite, err: " + err.Error()
		log.Error(msg)
		return c.String(http.StatusInternalServerError, msg)
	}

	// nats
	if NATSConn == nil {
		msg := "[ProduceFood] failed to get NATS conn"
		log.Error(msg)
		return c.String(http.StatusInternalServerError, msg)
	}
	msgBytes, err := json.Marshal(food)
	if err != nil {
		msg := "[ProduceFood] food request: %v, marshal err: %v"
		log.Errorf(msg, food, err)
		return c.String(http.StatusInternalServerError, "could not serialize request")
	}
	err = NATSConn.Publish(msgqueue.NATS_TOPIC, msgBytes)
	if err != nil {
		msg := "[ProductFood] failed to publish message to NATS topic: %s, err: %v"
		log.Errorf(msg, msgqueue.NATS_TOPIC, err)
	} else {
		// set identifier in SQLite database
		err = db.UpdateIdentifierByID(sqliteDB, id, db.IDENTIFIER_BE_SENT_NATS)
		if err != nil {
			msg := "[ProduceFood] update identifier in SQLite, id: %s, err: %v"
			log.Errorf(msg, id, err)
			return c.String(http.StatusInternalServerError, "database err")
		}
	}

	return c.JSON(http.StatusCreated, food)
}

type ConsumeFoodResponse struct {
	Name   string `json:"name"`
	Number int64  `json:"number"`
}

func ConsumeFood(c echo.Context) error {
	sqliteDB, err := db.CreateConnection(db.SQLITE_FILE_NAME)
	if err != nil {
		msg := "[ConsumeFood] failed to connect SQLite, err: %v"
		log.Errorf(msg, err)
		return c.String(http.StatusInternalServerError, "failed to connect SQLite")
	}
	defer sqliteDB.Close()

	food, err := db.GetLatestFood(sqliteDB)
	if err != nil {
		msg := "[ConsumeFood] failed to get latest food from database, err :" + err.Error()
		log.Error(msg)
		return c.String(http.StatusInternalServerError, msg)
	}

	ids := make([]int64, len(food))
	foodResponse := make([]*ConsumeFoodResponse, len(food))
	for i, f := range food {
		foodResponse[i] = &ConsumeFoodResponse{
			Name:   f.Name,
			Number: f.Number,
		}
		ids[i] = f.ID
	}
	// Set identifier in the database
	err = db.UpdateIdentifierByIDList(sqliteDB, ids, db.IDENTIFIER_ALREADY_GET)
	if err != nil {
		msg := "[ConsumeFood] failed to update indentifier from database,id list: %v, err: %v"
		log.Errorf(msg, ids, err)
		return c.String(http.StatusInternalServerError, "database err")
	}

	return c.JSON(http.StatusOK, foodResponse)
}
