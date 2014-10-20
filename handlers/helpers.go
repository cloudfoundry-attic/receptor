package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/cloudfoundry-incubator/receptor"
)

func writeErrorResponse(w http.ResponseWriter, statusCode int, err error) error {
	return writeJSONResponse(w, statusCode, receptor.NewErrorResponse(err))
}

func writeJSONResponse(w http.ResponseWriter, statusCode int, jsonObj interface{}) error {
	jsonBytes, err := json.Marshal(jsonObj)
	if err != nil {
		panic("Unable to encode JSON: " + err.Error())
	}

	w.Header().Set("Content-Length", strconv.Itoa(len(jsonBytes)))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	_, err = w.Write(jsonBytes)
	return err
}
