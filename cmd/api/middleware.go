package main

import (
	"context"
	"net/http"
)

func (app *application) enableCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://*")

		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT,PATCH, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, X-CSRF-Token, Authorization")
			return
		} else {
			h.ServeHTTP(w, r)
		}

	})
}

func (app *application) authRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, claims, err := app.auth.GetTokenFromHeaderAndVerify(w, r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		userID := claims.Subject

		ctx := context.WithValue(r.Context(), "userID", userID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
