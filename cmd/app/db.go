package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

var (
	connStr = "user=postgres password=postgres dbname=emaildb sslmode=disable"
)

func connectToDb() *sql.DB {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	return db
}
