package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"polling/internal/repository"
	"polling/internal/repository/dbrepo"
)

const port = 8080

type application struct {
	Domain string
	DB     repository.Repository
	DSN    string
}

func main() {
	var app application

	flag.StringVar(&app.DSN, "dsn", "host=localhost port=5432 user=postgres password=postgres dbname=polling sslmode=disable timezone=UTC connect_timeout=5", "Postgres connection string")

	flag.Parse()

	conn, err := app.connectToDB()
	if err != nil {
		log.Fatal(err)
	}

	app.DB = &dbrepo.DBRepo{DB: conn}
	defer app.DB.Connection().Close()

	log.Println("Server starting on port: ", port)
	err = http.ListenAndServe(fmt.Sprintf(":%d", port), app.routes())

	if err != nil {
		log.Fatal("Failed to start the server: ", err)
	}

}
