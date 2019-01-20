package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"

	"database/sql"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func main() {
	initDB()

	router := mux.NewRouter()

	router.HandleFunc("/item/{item}/user/{user}", getItemAPI).Methods("GET")
	router.HandleFunc("/item/{item}/answer", postAnswer).Methods("POST")

	router.HandleFunc("/user/{user}/items", getUserItems).Methods("GET")
	router.HandleFunc("/user/{user}", getUserAPI).Methods("GET")

	log.Fatal(http.ListenAndServe(":8080", router))
}

func initDB() {
	var err error
	dbUser := os.Getenv("DBUSER")
	dbPass := os.Getenv("DBPASSWORD")
	db, err = sql.Open("mysql", dbUser+":"+dbPass+"@tcp(mydatabase.ctxb7zts2zyl.eu-central-1.rds.amazonaws.com:3306)/CHASM")
	if err != nil {
		log.Panic(err)
	}
}
