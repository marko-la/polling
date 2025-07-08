package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *application) routes() http.Handler {
	mux := chi.NewRouter()

	mux.Post("/signup", app.Signup)
	mux.Post("/login", app.Login)

	mux.Get("/polls", app.GetAllPolls)

	mux.Get("/poll/{pollID}", app.GetPoll)
	mux.Post("/poll/create", app.CreatePoll)
	mux.Post("/poll/update", app.UpdatePoll)
	mux.Delete("/poll/{pollID}", app.RemovePoll)

	mux.Post("/poll/add-options", app.AddPollOptions)
	mux.Post("/poll/option/update", app.UpdatePollOption)
	mux.Delete("/poll/option/{optionID}", app.RemovePollOption)

	return mux
}
