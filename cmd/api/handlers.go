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

	// create a jwt user

	u := jwtUser{
		ID: user.ID,

		FirstName: user.FirstName,
		LastName:  user.LastName,
	}

	// generate tokens
	tokens, err := app.auth.GenerateTokenPair(&u)
	if err != nil {
		app.writeError(w, err)
		return
	}

	refreshCookie := app.auth.GetRefreshCookie(tokens.RefreshToken)

	http.SetCookie(w, refreshCookie)

	res := struct {
		User   jwtUser    `json:"user"`
		Tokens TokenPairs `json:"tokens"`
	}{
		User:   u,
		Tokens: tokens,
	}

	app.writeJSON(w, http.StatusOK, res)
}

// poll routes handlers

func (app *application) CreatePoll(w http.ResponseWriter, r *http.Request) {
	userIDstr, ok := r.Context().Value("userID").(string)

	if !ok {
		app.writeError(w, errors.New("missing user"))
		return
	}

	userID, err := strconv.Atoi(userIDstr)

	if err != nil {
		app.writeError(w, err)
		return
	}

	var payload struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}

	err = app.readJSON(w, r, &payload)

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
		UserID:      userID,
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
	// read url parameters
	pollIDStr := chi.URLParam(r, "pollID")

	pollID, err := strconv.Atoi(pollIDStr)

	if err != nil {
		app.writeError(w, errors.New("invalid poll ID"))
		return
	}

	err = app.checkPollOwnership(w, r)

	if err != nil {
		app.writeError(w, err, http.StatusUnauthorized)
		return
	}

	var payload struct {
		Options []models.PollOption `json:"options"`
	}

	err = app.readJSON(w, r, &payload)

	if err != nil {
		app.writeError(w, err)
		return
	}

	err = app.DB.AddPollOptions(pollID, payload.Options)

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
	// read url parameters
	pollIDStr := chi.URLParam(r, "pollID")

	pollID, err := strconv.Atoi(pollIDStr)

	if err != nil {
		app.writeError(w, errors.New("invalid poll ID"))
		return
	}

	// check if user is authorized to update poll
	err = app.checkPollOwnership(w, r)

	if err != nil {
		app.writeError(w, err, http.StatusUnauthorized)
		return
	}

	// read data
	var payload struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	}

	err = app.readJSON(w, r, &payload)

	if err != nil {
		app.writeError(w, err)
		return
	}

	// update poll
	poll := models.Poll{
		Title:       payload.Title,
		Description: payload.Description,
	}

	err = app.DB.UpdatePollByID(pollID, poll)

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

	err = app.checkPollOwnership(w, r)

	if err != nil {
		app.writeError(w, err, http.StatusUnauthorized)
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
	optionIDStr := chi.URLParam(r, "optionID")

	optionID, err := strconv.Atoi(optionIDStr)
	if err != nil {
		app.writeError(w, errors.New("invalid ID"))
		return
	}

	err = app.checkPollOwnership(w, r)

	if err != nil {
		app.writeError(w, err, http.StatusUnauthorized)
		return
	}

	var payload struct {
		Text string `json:"text"`
	}

	err = app.readJSON(w, r, &payload)

	if err != nil {
		app.writeError(w, err)
		return
	}

	err = app.DB.UpdateOptionByID(optionID, payload.Text)

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

	err = app.checkPollOwnership(w, r)

	if err != nil {
		app.writeError(w, err, http.StatusUnauthorized)
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
	userIDstr, ok := r.Context().Value("userID").(string)

	if !ok {
		app.writeError(w, errors.New("missing user"))
		return
	}

	userID, err := strconv.Atoi(userIDstr)

	if err != nil {
		app.writeError(w, err)
		return
	}

	pollIDStr := chi.URLParam(r, "pollID")

	pollID, err := strconv.Atoi(pollIDStr)
	if err != nil {
		app.writeError(w, errors.New("invalid poll ID"))
		return
	}

	optionIDStr := chi.URLParam(r, "optionID")

	optionID, err := strconv.Atoi(optionIDStr)
	if err != nil {
		app.writeError(w, errors.New("invalid option ID"))
		return
	}

	err = app.DB.Vote(pollID, optionID, userID)

	if err != nil {
		app.writeError(w, err)
		return
	}

	app.writeMessage(w, "Voted successfully")

}

func (app *application) Unvote(w http.ResponseWriter, r *http.Request) {
	userIDstr, ok := r.Context().Value("userID").(string)

	if !ok {
		app.writeError(w, errors.New("missing user"))
		return
	}

	userID, err := strconv.Atoi(userIDstr)

	if err != nil {
		app.writeError(w, err)
		return
	}

	optionIDStr := chi.URLParam(r, "optionID")

	optionID, err := strconv.Atoi(optionIDStr)
	if err != nil {
		app.writeError(w, errors.New("invalid ID"))
		return
	}

	err = app.DB.Unvote(optionID, userID)

	if err != nil {
		app.writeError(w, err)
		return
	}

	app.writeMessage(w, "Unvoted successfully")
}

func (app *application) GetOptionVotes(w http.ResponseWriter, r *http.Request) {
	optionIDStr := chi.URLParam(r, "optionID")

	optionID, err := strconv.Atoi(optionIDStr)
	if err != nil {
		app.writeError(w, errors.New("invalid ID"))
		return
	}

	votes, err := app.DB.GetOptionVotes(optionID)

	if err != nil {
		app.writeError(w, err)
		return
	}

	app.writeJSON(w, http.StatusOK, votes)
}
