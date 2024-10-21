package db

import (
	"testing"

	_ "github.com/glebarez/sqlite"
)

func TestCreateConnection(t *testing.T) {
	_, err := CreateConnection("./data/food.db")
	if err != nil {
		t.Errorf("CreateConnection() error = %v", err)
	}
}
