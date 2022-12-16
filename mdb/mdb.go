package mdb

import (
	"database/sql"
	"log"
	"time"

	"github.com/mattn/go-sqlite3"
)

type EmailEntry struct {
	Id          int64
	Email       string
	ConfirmedAt *time.Time
	OptOut      bool
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

// CRUD OPERATIONS

// CREATE EMAIL

func CreateEmail(db *sql.DB, email string) error {
	query := `INSERT INTO
		emails(email, confirmed_at, opt_out)
		VALUES(?, 0, false)`
	_, err := db.Exec(query, email)
	// check for errors
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

// READ EMAIL

func GetEmail(db *sql.DB, email string) (*EmailEntry, error) {
	query := `SELECT id, email, confirmed_at, opt_out
		FROM emails
		WHERE email = ?`
	// the data returned by the Query method is returned as rows
	rows, err := db.Query(query, email)
	// check for errors
	if err != nil {
		log.Println(err)
		return nil
	}
	// the Query methd keeps the db open so it can keep reading more rows so we need to defer closing it to freexe up the db
	defer rows.Close()

	// iterate through the rows and return an EmailEntry from row
	// rows.Next returns a sql.Rows data type which we pass to our emailEntryFromRow function to create a new EmailEntry
	for rows.Next() {
		return emailEntryFromRow(rows)
	}
	return nil, nil
}

// UPDATE EMAIL

func UpdateEmail(db *sql.DB, entry EmailEntry) error {
	t := entry.ConfirmedAt.Unix()
	query := `INSET INTO
		emails(email, confirmed_at, opt_out)
		VALUES(?, ?, ?)
		ON CONFLICT(email) DO UPDATE SET
			confirmed_at=?
			opt_out=?
		`
	_, err := db.Exec(query, entry.Email, t, entry.OptOut, t, entry.OptOut)
	// check for errors
	if err != nil {
		log.Println(err)
		return nil
	}

	return nil
}

// DELETE EMAIL

func DeleteEmail(db *sql.DB, email string) error {
	// Normally a delete operation would delete a row from db like so:
	// query := `DELETE FROM emails
	// 	WHERE email=?
	// `
	// But since this app is a mailing list, we just want to update the opt_out field to be true
	query := `UPDATE emails
			  SET opt_out=true
			  WHERE email=?`
	_, err := db.Exec(query, email)
	// check for errors
	if err != nil {
		log.Println(err)
		return nil
	}
	return nil
}
