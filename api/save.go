package handler

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
)

func NewHandler(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("mysql", os.Getenv("DSN"))
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		fmt.Fprintf(w, "Ping error")
		return
	}
	fmt.Fprintf(w, "<h1>Connected!</h1>")
}
