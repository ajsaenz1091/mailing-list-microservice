package mdb

import (
	"database/sql"
	"log"
	"time"
	"github.com/mattn/go-sqlite3"
)

type EmailEntry struct {
	Id int64
	Email string
	ConfirmedAt *time.Time
	OptOut bool
}

func TryCreate(db *sql.Db) {
	query := `	CREATE TABLE emails (
		id				INTEGER PRIMARY KEY,
		email			TEXT UNIQUE,
		confirmed_at	INTEGER,
		opt_out			INTEGER,	
	);`
	_, err := db.Exec(query)
	// check for error
	if err != nil {
		if sqlError, ok := err.(sqlite3.Error); ok {
			// code 1 == "table already exists"
			if sqlError.Code != 1 {
				log.Fatal(sqlError)
			}
		} else {
			log.Fatal(err)
		}
	}
}

// this function creates an EmailEntry structure from a db row
func emailEntryFromRow(row *sql.Rows) (*EmailEntry, error) {
	var id int64
	var email string
	var confirmedAt int64
	var optOut bool

	// read data out of db using row.scan function
	err := row.Scan(&id, &email, &confirmedAt, &optOut) // use pointers to read data into the variables

	// check for errors
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// convert time to apropriate time structure
	t := time.Unix(confirmedAt, 0)
	return &EmailEntry{Id: id, Email: email, ConfirmedAt: &t, OptOut: optOut}, nil
}