package handler

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func init() {
	var err error
	db, err = sql.Open("mysql", os.Getenv("DSN"))
	if err != nil {
		panic(err)
	}

	db.SetMaxOpenConns(10)                 // set maximum number of open connections
	db.SetMaxIdleConns(5)                  // set maximum number of idle connections
	db.SetConnMaxLifetime(time.Minute * 1) // set maximum connection lifetime
}

func NewHandler(w http.ResponseWriter, r *http.Request) {
	// Get a connection from the pool
	conn, err := db.Conn(r.Context())
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	if err := conn.PingContext(r.Context()); err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "<h1>Connected!</h1>")
}
