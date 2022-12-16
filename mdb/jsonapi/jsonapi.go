package jsonapi

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
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
