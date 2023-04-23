package handler

import (
	"database/sql"
	"encoding/json"
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

	conn, err := db.Conn(r.Context())
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	switch r.Method {
	case "POST":
		var data map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		label, ok := data["label"].(string)
		if !ok || len(label) > 100 {
			http.Error(w, "Invalid label", http.StatusBadRequest)
			return
		}

		dataJSON, err := json.Marshal(data["data"])
		if err != nil {
			http.Error(w, "Invalid data", http.StatusBadRequest)
			return
		}

		conn, err := db.Conn(r.Context())
		if err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			return
		}
		defer conn.Close()

		query := "INSERT INTO link (label, data) VALUES (?, ?)"
		if _, err := conn.ExecContext(r.Context(), query, label, dataJSON); err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Data saved successfully")
	case "GET":
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "Missing id parameter", http.StatusBadRequest)
			return
		}

		conn, err := db.Conn(r.Context())
		if err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			return
		}
		defer conn.Close()

		var label string
		var dataJSON []byte
		query := "SELECT label, data FROM link WHERE id = ?"
		row := conn.QueryRowContext(r.Context(), query, id)
		if err := row.Scan(&label, &dataJSON); err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Link not found", http.StatusNotFound)
			} else {
				http.Error(w, "Something went wrong", http.StatusInternalServerError)
			}
			return
		}

		var data interface{}
		if err := json.Unmarshal(dataJSON, &data); err != nil {
			http.Error(w, "Invalid data", http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"id":    id,
			"label": label,
			"data":  data,
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			return
		}
	case "PUT":
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "Missing id parameter", http.StatusBadRequest)
			return
		}

		var data map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		label, ok := data["label"].(string)
		if !ok || len(label) > 100 {
			http.Error(w, "Invalid label", http.StatusBadRequest)
			return
		}

		dataJSON, err := json.Marshal(data["data"])
		if err != nil {
			http.Error(w, "Invalid data", http.StatusBadRequest)
			return
		}

		query := "UPDATE link SET label = ?, data = ?, updated_at = NOW() WHERE id = ?"
		if _, err := conn.ExecContext(r.Context(), query, label, dataJSON, id); err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Data updated successfully")
	case "DELETE":
		id := r.URL.Query().Get("id")
		query := "DELETE FROM link WHERE id = ?"
		if _, err := conn.ExecContext(r.Context(), query, id); err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Data deleted successfully")
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Header().Set("Allow", "PUT, DELETE")
		fmt.Fprintf(w, "Method %s Not Allowed", r.Method)
	}
}
