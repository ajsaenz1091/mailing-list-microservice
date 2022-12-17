package jsonapi

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"mailinglist/mdb"
	"net/http"
)

func setJsonHeader(w http.ResponseWriter) {
	w.Header().Set("content-type", "application/json; charset=utf-8")
}

// convert json data from the wire into a go structure
func fromJson[T any](body io.Reader, target T) {
	// create new buffer
	buf := new(bytes.Buffer)
	// read data into buffer
	buf.ReadFrom(body)
	// use Unmarshall to convert bytes into a target structure (any structure type T)
	json.Unmarshal(buf.Bytes(), &target)
}

// function to return json data\
func returnJson[T any](w http.ResponseWriter, withData func() (T, error)) {
	setJsonHeader(w)

	data, serverErr := withData()
	// check for errors
	if serverErr != nil {
		w.WriteHeader(500)
		// marshal converts structures into json data
		serverErrJson, err := json.Marshal(&serverErr)
		if err != nil {
			log.Print(err)
			return
		}
		w.Write(serverErrJson)
		return
	}
	dataJson, err := json.Marshal(&data)
	if err != nil {
		log.Print(err)
		w.WriteHeader(500)
		return
	}

	w.Write(dataJson)
}

// function to return an error
func returnErr(w http.ResponseWriter, err error, code int) {
	returnJson(w, func() (interface{}, error) {
		errorMessage := struct {
			Err string
		}{
			Err: err.Error(),
		}
		w.WriteHeader(code)
		return errorMessage, nil
	})
}

// API HANDLERS

// CREATE

func CreateEmail(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "POST" {
			return
		}

		entry := mdb.EmailEntry{}
		fromJson(req.Body, &entry)

		if err := mdb.CreateEmail(db, entry.Email); err != nil {
			returnErr(w, err, 400)
			return
		}

		returnJson(w, func() (interface{}, error) {
			log.Printf("JSON CreateEmail: %v\n", entry.Email)
			return mdb.GetEmail(db, entry.Email)
		})
	})
}

// GET

func GetEmail(db *sql.DB) http.Hander {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "GET" {
			return
		}

		entry := mdb.EmailEntry{}
		fromJson(req.Body, &entry)

		returnJson(w, func() (interface{}, error) {
			log.Printf("JSON GetEmail: %v\n", entry.Email)
			return mdb.GetEmail(db, entry.Email)
		})
	})
}

// GET BATCH

func GetEmailBatch(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "GET" {
			return
		}

		queryParams := mdb.GetEmailBatchQueryParams{}
		fromJson(req.Body, &queryParams)

		if queryParams.Count <= 0 || queryParams.Page <= 0 {
			returnErr(w, errors.New("Page and Count fields are required and must be > 0"), 400)
			return
		}

		returnJson(w, func() (interface{}, error) {
			log.Printf("JSON GetEmailBatch: %v\n", queryParams)
			return mdb.GetEmailBatch(db, queryParams)
		})
	})
}

// UPDATE

func UpdateEmail(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "PUT" {
			return
		}

		entry := mdb.EmailEntry{}
		fromJson(req.Body, &entry)

		if err := mdb.UpdateEmail(db, entry); err != nil {
			returnErr(w, err, 400)
			return
		}

		returnJson(w, func() (interface{}, error) {
			log.Printf("JSON UpdateEmail: %v\n", entry)
			return mdb.GetEmail(db, entry.Email)
		})
	})
}

// DELETE

func DeleteEmail(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "POST" {
			return
		}

		entry := mdb.EmailEntry{}
		fromJson(req.Body, &entry)

		if err := mdb.DeleteEmail(db, entry.Email); err != nil {
			returnErr(w, err, 400)
			return
		}

		returnJson(w, func() (interface{}, error) {
			log.Printf("JSON DeleteEmail: %v\n", entry)
			return mdb.GetEmail(db, entry.Email)
		})
	})
}

// this function will serve up each handler
func Serve(db *sql.DB, bind string) {
	// the Handle method creates a handler at the specified address. the second parameter is the handler function to be triggered
	http.Handle("/email/create", CreateEmail(db))
	http.Handle("/email/get", GetEmail(db))
	http.Handle("/email/get_batch", GetEmailBatch(db))
	http.Handle("/email/update", UpdateEmail(db))
	http.Handle("/email/delete", DeleteEmail(db))
	err := http.ListenAndServe(bind, nil)
	if err != nil {
		// If an error is returned, Fatalf terinates the application
		log.Fatalf("JSON server error: %v", err)
	}
}
