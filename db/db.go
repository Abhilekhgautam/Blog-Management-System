package db

import (
	"database/sql"
	"fmt"
	"github.com/go-sql-driver/mysql"
	"log"
	"os"
)

var Db *sql.DB

//ConnectDB - connect the database with provided configuration
func ConnectDB() {
	// configuration for mysql..
	cfg := mysql.Config{
		User:                 os.Getenv("User"),
		Passwd:               os.Getenv("Pass"),
		Net:                  "tcp",
		Addr:                 "127.0.0.1:3306",
		DBName:               "blog",
		AllowNativePasswords: true,
	}

	var err error
	// open  a database handle.
	Db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}

	pingErr := Db.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
	}
	fmt.Println("Connected to the database successfully")
}
