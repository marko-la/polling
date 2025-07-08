package main

import (
	"errors"
	"net/http"
	"polling/internal/models"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

// user routes handlers

func (app *application) Signup(w http.ResponseWriter, r *http.Request) {

	var payload struct {
		Username  string `json:"username"`
		Password  string `json:"password"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}

	err := app.readJSON(w, r, &payload)

	if err != nil {
		app.writeError(w, err)
		return
	}

	if payload.Username == "" || payload.Password == "" || payload.FirstName == "" || payload.LastName == "" {
		app.writeError(w, errors.New("missing one or more required field ['username','password','first_name', 'last_name']"))
		return
	}

	user := models.User{
		Username:  payload.Username,
		FirstName: payload.FirstName,
		LastName:  payload.LastName,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	user.HashPassword(payload.Password)

	err = app.DB.CreateUser(user)

	if err != nil {
		app.writeError(w, err)
		return
	}

	app.writeMessage(w, "User successfuly created")
}

func (app *application) Login(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &payload)

	if err != nil {
		app.writeError(w, err)
		return
	}

	if payload.Username == "" || payload.Password == "" {
		app.writeError(w, errors.New("missing one or more required field ['username','password']"))
		return
	}

	user, err := app.DB.GetUserByUsername(payload.Username)

	if err != nil {
		app.writeError(w, err)
		return
	}

	valid := user.CheckPassword(payload.Password)

	if !valid {
		app.writeError(w, errors.New("password incorrect"), http.StatusUnauthorized)
		return
	}

	app.writeJSON(w, http.StatusOK, user)
}

// poll routes handlers

func (app *application) CreatePoll(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}

	err := app.readJSON(w, r, &payload)

	if err != nil {
		app.writeError(w, err)
		return
	}

	if payload.Title == "" {
		app.writeError(w, errors.New("missing one or more required field ['title']"))
		return
	}

	poll := models.Poll{
		Title:       payload.Title,
		Description: payload.Description,
	}

	created, err := app.DB.CreatePoll(poll)

	if err != nil {
		app.writeError(w, err)
		return
	}

	app.writeJSON(w, http.StatusOK, created)
}

func (app *application) GetAllPolls(w http.ResponseWriter, r *http.Request) {

	polls, err := app.DB.GetAllPolls()

	if err != nil {
		app.writeError(w, err)
		return
	}

	app.writeJSON(w, http.StatusOK, polls)
}

func (app *application) AddPollOptions(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		PollID  int                 `json:"poll_id"`
		Options []models.PollOption `json:"options"`
	}

	err := app.readJSON(w, r, &payload)

	if err != nil {
		app.writeError(w, err)
		return
	}

	err = app.DB.AddPollOptions(payload.PollID, payload.Options)

	if err != nil {
		app.writeError(w, err)
		return
	}

	app.writeMessage(w, "options added successfully!")
}

func (app *application) GetPoll(w http.ResponseWriter, r *http.Request) {
	pollIDStr := chi.URLParam(r, "pollID")

	pollID, err := strconv.Atoi(pollIDStr)
	if err != nil {
		app.writeError(w, errors.New("invalid poll ID"))
		return
	}

	poll, err := app.DB.GetPollByID(pollID)

	if err != nil {
		app.writeError(w, err)
		return
	}

	app.writeJSON(w, http.StatusOK, poll)
}

func (app *application) UpdatePoll(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		PollID      int    `json:"poll_id"`
		Title       string `json:"title"`
		Description string `json:"description"`
	}

	err := app.readJSON(w, r, &payload)

	if err != nil {
		app.writeError(w, err)
		return
	}

	poll := models.Poll{
		Title:       payload.Title,
		Description: payload.Description,
	}

	err = app.DB.UpdatePollByID(payload.PollID, poll)

	if err != nil {
		app.writeError(w, err)
		return
	}

	app.writeMessage(w, "Poll updated")

}

func (app *application) RemovePoll(w http.ResponseWriter, r *http.Request) {
	pollIDStr := chi.URLParam(r, "pollID")

	pollID, err := strconv.Atoi(pollIDStr)
	if err != nil {
		app.writeError(w, errors.New("invalid poll ID"))
		return
	}

	err = app.DB.DeletePollByID(pollID)

	if err != nil {
		app.writeError(w, err)
		return
	}

	app.writeMessage(w, "Poll deleted")
}

func (app *application) UpdatePollOption(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		OptionID int    `json:"option_id"`
		Text     string `json:"text"`
	}

	err := app.readJSON(w, r, &payload)

	if err != nil {
		app.writeError(w, err)
		return
	}

	err = app.DB.UpdateOptionByID(payload.OptionID, payload.Text)

	if err != nil {
		app.writeError(w, err)
		return
	}

	app.writeMessage(w, "Option updated")
}

func (app *application) RemovePollOption(w http.ResponseWriter, r *http.Request) {
	optionIDStr := chi.URLParam(r, "optionID")

	optionID, err := strconv.Atoi(optionIDStr)
	if err != nil {
		app.writeError(w, errors.New("invalid ID"))
		return
	}

	err = app.DB.DeleteOptionByID(optionID)

	if err != nil {
		app.writeError(w, err)
		return
	}

	app.writeMessage(w, "Option deleted")
}

func (app *application) Vote(w http.ResponseWriter, r *http.Request) {

}

func (app *application) GetOptionVotes(w http.ResponseWriter, r *http.Request) {

}
