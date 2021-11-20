package main

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/ebosas/microservices/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/streadway/amqp"
)

// TestInsertToDB tests whether a Rabbit message is inserted
// correctly into the database
func TestInsertToDB(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock connection: %s", err)
	}
	defer db.Close()

	now := time.Now().UnixMilli()
	var tests = []struct {
		message string
		source  string
		time    int64
	}{
		{"Hello", "back", now},
		{"Another test!", "back", now},
		{"1", "front", now},
		{" ", "back", now - 60*60*1000},
	}
	for i, test := range tests {
		mock.ExpectExec("insert into messages").WithArgs(test.message, test.time/1000).WillReturnResult(sqlmock.NewResult(int64(i), 1))

		d := testArguments(t, test.message, test.source, test.time)
		insertToDB(*d, db)

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unfulfilled expectations: %s", err)
		}
	}
}

// testArguments produces arguments for the function being tested
func testArguments(t *testing.T, message, source string, time int64) *amqp.Delivery {
	inputMsg := models.Message{Text: message, Source: source, Time: time}
	messageJson, err := json.Marshal(inputMsg)
	if err != nil {
		t.Fatalf("marshal message: %s", err)
	}
	d := &amqp.Delivery{Body: messageJson}

	return d
}
