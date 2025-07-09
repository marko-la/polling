package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v4"
)

func (app *application) routes() http.Handler {
	mux := chi.NewRouter()

	mux.Use(middleware.Recoverer)
	mux.Use(app.enableCORS)

	mux.Post("/signup", app.Signup)
	mux.Post("/login", app.Login)

	mux.Get("/polls", app.GetAllPolls)
	mux.Get("/polls/{pollID}", app.GetPoll)

	mux.Get("/polls/{pollID}/options/{optionID}/votes", app.GetOptionVotes)

	mux.Route("/", func(r chi.Router) {
		r.Use(app.authRequired)
		r.Post("/polls/create", app.CreatePoll)
		r.Put("/polls/{pollID}", app.UpdatePoll)
		r.Delete("/polls/{pollID}", app.RemovePoll)

		r.Post("/polls/{pollID}/options", app.AddPollOptions)
		r.Put("/polls/{pollID}/options/{optionID}", app.UpdatePollOption)
		r.Delete("/polls/{pollID}/options/{optionID}", app.RemovePollOption)

		r.Put("/polls/{pollID}/options/{optionID}/votes", app.Vote)
		r.Delete("/polls/{pollID}/options/{optionID}/votes", app.Unvote)
	})

	return mux
}

func (j *Auth) GetTokenFromHeaderAndVerify(w http.ResponseWriter, r *http.Request) (string, *Claims, error) {
	w.Header().Add("Vary", "Authorization")

	// get auth header
	authHeader := r.Header.Get("Authorization")

	// sanity check
	if authHeader == "" {
		return "", nil, errors.New("no auth heder")
	}

	// split the header
	headerParts := strings.Split(authHeader, " ")

	// check if the header is in the correct format
	// "Bearer <token>"
	if len(headerParts) != 2 || headerParts[0] != "Bearer" {
		return "", nil, errors.New("invalid auth header format")
	}

	token := headerParts[1]

	// declare an empty claims

	claims := &Claims{}

	// parse the token
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(j.Secret), nil

	})

	if err != nil {
		if strings.HasPrefix(err.Error(), "token is expired by") {
			return "", nil, errors.New("expired token")
		}

		return "", nil, err

	}

	if claims.Issuer != j.Issuer {
		return "", nil, errors.New("invalid issuer")
	}

	return token, claims, nil

}
