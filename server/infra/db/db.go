package db

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func init() {
	d, err := sql.Open("sqlite3", "server/test/test.db")
	if err != nil {
		log.Fatal(err)
	} else if d == nil {
		log.Fatal("nil db generated")
	}
	if _, err := d.Exec("PRAGMA foreign_keys = ON"); err != nil {
		log.Fatal(err)
	}
	DB = d
}
