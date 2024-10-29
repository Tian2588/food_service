package main

import (
	msgqueue "food_web_service/msg_queue"
	"log"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	nats "github.com/nats-io/nats.go"
)

var (
	NATSConn *nats.Conn
)

func main() {
	// Echo instance
	e := echo.New()
	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Skipper:      middleware.DefaultSkipper,
		ErrorMessage: "custom timeout error message returns to client",
		OnTimeoutRouteErrorHandler: func(err error, c echo.Context) {
			log.Println(c.Path())
		},
		Timeout: 10 * time.Second,
	}))

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	// NATS
	NATSConn, _ = msgqueue.ConnectNATS()

	// Routes
	e.POST("/v1/food", ProduceFood)
	e.GET("/v1/food", ConsumeFood)

	// Start server
	e.Logger.Fatal(e.Start(":1323"))
}
