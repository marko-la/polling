package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type JSONResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, data any) error {
	maxBytes := 1024 * 1024 // 1MB
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)

	dec.DisallowUnknownFields()

	err := dec.Decode(data)

	if err != nil {
		return err
	}

	err = dec.Decode(&struct{}{})

	if err != io.EOF {
		return errors.New("body must contain only a single JSON value")
	}

	return nil

}

func (app *application) writeJSON(w http.ResponseWriter, status int, data any, headers ...http.Header) error {

	out, err := json.Marshal(data)

	if err != nil {
		return err
	}

	if len(headers) > 0 {
		for key, value := range headers[0] {
			w.Header()[key] = value
		}
	}

	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(status)

	_, err = w.Write(out)

	if err != nil {
		return err
	}

	return nil

}

func (app *application) writeError(w http.ResponseWriter, err error, status ...int) error {
	statusCode := http.StatusBadRequest

	if len(status) > 0 {
		statusCode = status[0]
	}

	var payload JSONResponse
	payload.Error = true
	payload.Message = err.Error()

	return app.writeJSON(w, statusCode, payload)
}

func (app *application) writeMessage(w http.ResponseWriter, message string, status ...int) error {
	statusCode := http.StatusOK

	if len(status) > 0 {
		statusCode = status[0]
	}

	var payload JSONResponse
	payload.Error = false
	payload.Message = message

	return app.writeJSON(w, statusCode, payload)
}

func (app *application) checkPollOwnership(w http.ResponseWriter, r *http.Request) error {
	userIDstr, ok := r.Context().Value("userID").(string)

	if !ok {
		return errors.New("missing user")
	}

	userID, err := strconv.Atoi(userIDstr)

	if err != nil {
		return err
	}

	pollIDStr := chi.URLParam(r, "pollID")

	pollID, err := strconv.Atoi(pollIDStr)

	if err != nil {
		return errors.New("invalid poll ID")
	}

	if !app.DB.IsPollOwner(pollID, userID) {
		return errors.New("you are not authorized to update this poll")
	}

	return nil
}
