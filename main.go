package main

import (
	"database/sql"
	"log"
	"net/http"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/thegera4/go-htmx-ecommerce/pkg/handlers"
	"github.com/thegera4/go-htmx-ecommerce/pkg/repository"
)

/*** Global Variables ***/
var db *sql.DB // Database instance

// Initialize the database
func initDB() {
	var err error
	db, err = sql.Open("mysql", "root:toor@(127.0.0.1:3306)/ecommerce?parseTime=true")
	if err != nil { log.Fatal(err)}

	if err = db.Ping(); err != nil { log.Fatal(err)}
}

func main() {

	r := mux.NewRouter()

	//Setup MySQL
	initDB()
	defer db.Close()

	// Setup Static folder for static files and images
	fs := http.FileServer(http.Dir("./static"))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", fs))

	repo := repository.NewRepository(db)
	handler := handlers.NewHandler(repo)

	r.HandleFunc("/seed-products", handler.SeedProducts).Methods("POST")

	http.ListenAndServe(":8080", r)
}