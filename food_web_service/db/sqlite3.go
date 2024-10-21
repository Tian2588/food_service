package db

import (
	"database/sql"
	"strings"
	"sync"

	_ "github.com/glebarez/sqlite"
	"github.com/labstack/gommon/log"
)

const (
	// define identifier status, by default should be IDENTIFIER_TO_SEND_NATS=0
	IDENTIFIER_BE_SENT_NATS          = 1
	IDENTIFIER_ALREADY_CONSUMED_NATS = 2
	IDENTIFIER_ALREADY_GET           = 3

	// SQLite
	SQLITE_FILE_NAME = "/data/food.db"
)

// CreateConnection creates a database connection for SQLite3.
func CreateConnection(fileName string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", fileName)
	if err != nil {
		return nil, err
	}
	err = CreateTableFoodIfNotExists(db)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func CreateTableFoodIfNotExists(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS food(
	    id INTEGER PRIMARY KEY AUTOINCREMENT,
	    name VARCHAR(50) NOT NULL,
	    number INTEGER NOT NULL,
	    identifier TINYINT  DEFAULT 0
	);
	`
	_, err := db.Exec(query)
	if err != nil {
		log.Errorf("[DB] failed to exec create table food if not exists, err: %v", err)
		return err
	}
	return nil
}

// UpsertFood inserts data into the table 'food'
func UpsertFood(db *sql.DB, name string, num int64) (int64, error) {
	stmt, err := db.Prepare("INSERT INTO food(name, number) VALUES(?, ?)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	result, err := stmt.Exec(name, num)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func UpdateIdentifierByIDList(db *sql.DB, ids []int64, identifier int8) error {
	var wg sync.WaitGroup
	for _, id := range ids {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := UpdateIdentifierByID(db, id, identifier)
			if err != nil {
				log.Errorf("[DB] failed to update indentifier by id: %d, err: %v ", id, err)
			}
		}()
	}
	wg.Wait()
	return nil
}

func UpdateIdentifierByID(db *sql.DB, id int64, identifier int8) error {
	stmt, err := db.Prepare("UPDATE food SET identifier = ? WHERE id IN (?)")
	if err != nil {
		msg := "[DB] failed to UpdateFoodIdentifier, err: %v"
		log.Errorf(msg, err)
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(identifier, id)
	if err != nil {
		msg := "[DB] failed to UpdateFoodIdentifier id:%d identifier:%v, err: %v"
		log.Errorf(msg, id, identifier, err)
		return err
	}
	return nil
}

// GetLatestFood retrieves latest data from the table 'food' since last time
func GetLatestFood(db *sql.DB) ([]*Food, error) {
	var food []*Food
	rows, err := db.Query("SELECT id, name, number, identifier FROM food where identifier != ?", IDENTIFIER_ALREADY_GET)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var name string
		var id, num int64
		var identifier int8
		err = rows.Scan(&id, &name, &num, &identifier)
		if err != nil {
			if strings.Contains(err.Error(), "index 0") {
				continue
			}
			return nil, err
		}
		food = append(food, &Food{
			ID:     id,
			Name:   name,
			Number: num,
		})
	}
	return food, nil
}
