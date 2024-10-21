# Introduction about this repo

There are two services: food_web_service and food_printer_service. Also a SQLite database and a NATS message queue are involved.
## food_web_service
food_web_service has 2 endpoints: POST to produce food (which has 2 fields: name and number) and GET to consume the latest food which has not been consumed.

- POST: /v1/food    
- GET : /v1/food

It is a simple web service exposed port `1323` using [Echo framework](https://github.com/labstack/echo) with its SQLite database, writes and reads food in the SQLite database and pubilsh the food to [NATS](https://nats.io/) when producing. When do POST, this service parses the request and validates,then connects to the SQLite database, stores the food in a SQLite database and if stored successfully, then publishes the food message via a connected NATS with the topic named `food`, then updates the identifier of the food. When do GET, this service connects to the SQLite database, retrieves the food whose identifier is not IDENTIFIER_ALREADY_GET and then update the identifier to IDENTIFIER_ALREADY_GET.

```
  // define identifier of the food
  IDENTIFIER_TO_SEND_NATS          = 0
  IDENTIFIER_BE_SENT_NATS          = 1
  IDENTIFIER_ALREADY_CONSUMED_NATS = 2
  IDENTIFIER_ALREADY_GET           = 3
```

## food_web_service
food_printer_service starts after NATS, listening to the specified uri and port of NATS, when running, it subscribes to the NATS topic named `food`. If any message is received, it prints the message to the console.

## SQLite database and NATS
The correct base image is chosen, like "alpine:latest" for SQLite and "nats:0.8.0" for NATS. Also other related parameters are configured in the `compose.yaml`, such as, "depends_on" and "ports" or "expose" in order to let them work well with each other.

# Building and running the services

When you're ready, start your services by running:
`docker compose up --build`.

Your food_web_service will be available at `http://localhost:1323`.

The following cmds can be executed to operate food_web_service:

- For `POST` to produce food:     
`curl -X POST -H "Content-Type: application/json" -d '{"name":"apple", "number":3}' http://localhost:1323/v1/food
{"name":"apple","number":3}`
- For `GET` to consume food which is already produced but has not been consumed:      
` curl -X GET -H "Content-Type: application/json" http://localhost:1323/v1/food`

# Requests performed and snapshots of running the services 
1. try to consume, no food is returned
2. try to produce 5 banana
3. try to produce 2 apples
4. try to get food, 5 banana and 2 apples are returned
5. try to produce 3 apples
6. try to get food, 3 apples are returned
7. try to get food, no food is returned

![alt text](<req_data.png>) 

![alt text](<print_data.png>)

# Some Tips about the issues ever met 
1. When to implement SQLite database, should use no cgo package, like `"github.com/glebarez/sqlite"`. Do not use cgo package, like "github.com/mattn/go-sqlite3", otherwise, you will get the following message and could also get the same message and still could not run the service even after with execution 'export CGO_ENABLED=1':
`please export CGO_ENABLED=1
Because: Binary was compiled with 'CGO_ENABLED=0', sqlite3 requires cgo to work.`

2. When to connect to SQLite database, make sure to have the right format of the database file, like `food.db`, do not use `food.sqlite`, otherwise you may be able to connect to the database but perhaps read not correctly from the database. 

3. When to create a table if not exists in the SQLite database, found that command `sqlite3 /data/food.db < /data/init.sql` under `web-server` in the `compose.yaml` does not work well, so write a function named CreateTableFoodIfNotExists to realize it.

4. When to connect to NATS, make sure to set the environment `NATS_URI` in the corresponding services and use package `"github.com/nats-io/nats.go"` instead of "github.com/nats-io/nats" because some related messages just come out during building the services. 
`github.com/nats-io/go-nats: github.com/nats-io/go-nats@v1.8.1: parsing go.mod:
	module declares its path as: github.com/nats-io/nats.go
	        but was required as: github.com/nats-io/go-nats`

5. Make sure to set the correct value of parameters in the compose.yaml, only then can you set up your services correctly.


# Note
Here is just a simple implementation about the coding challenge. Some unit tests are written instead of mocking because mocking need extra time writing, a little time-consuming. Avoid overengineering, so here just to realize the basic business logic. If services are to deployed to prod, many TODOs need to be considered, like latency, authentication, authorization, encryption, data consistency and so on. 
