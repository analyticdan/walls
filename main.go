package main

import (
	"fmt"
	"log"
	"net/http"
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

/*
Use mutexes around DB. (Or transactions. Or whatever they taught in CS168/CS162.) Many such race conditions.
*/

var db *sql.DB

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	// Only handle "/".
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	// Set no-cache header.
	w.Header().Set("Cache-Control", "no-cache")
	// Serve different pages based on log in status.
	isLoggedIn, err := validateCookies(r)
	if err != nil {
		serverError(w, err, "could not validate cookies.")
		return
	} else if isLoggedIn {
		uid, err := getUID(r)
		if err != nil {
			serverError(w, err, "could not get UID from cookies.")
			return
		}
		username, err := findUsername(uid)
		if err != nil {
			serverError(w, err, "could not match uid to username")
			return
		}
		http.Redirect(w, r, fmt.Sprintf("/wall/%s", username), 307)
	} else {
		http.ServeFile(w, r, "static/index.html")
	}
	return
}

func main() {
	var err error
	// Open connection to database.
	db, err = sql.Open("sqlite3", "database/data.db")
	if (err != nil) {
		log.Fatal(err)
	}
	defer db.Close()
	// Create users table if it doesn't exist.
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS users(uid INTEGER PRIMARY KEY, username VARCHAR(255), hash CHAR(60));")
	if (err != nil) {
		log.Fatal(err)
	}
	// Create cookie table if it doesn't exist.
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS cookies(uid INTEGER, sid CHAR(36) UNIQUE);")
	if (err != nil) {
		log.Fatal(err)
	}
	// Create posts table if it doesn't exist.
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS posts(from_uid INTEGER, to_uid INTEGER, body VARCHAR(255), date DATETIME);")
	if (err != nil) {
		log.Fatal(err)
	}

	// Register handlers.
	mux := http.NewServeMux()
	mux.Handle("/imgs/", http.FileServer(http.Dir(".")))
	mux.HandleFunc("/wall/", wallHandler)
	mux.HandleFunc("/signup", signupHandler)
	mux.HandleFunc("/login", loginHandler)
	mux.HandleFunc("/logout", logoutHandler)
	mux.HandleFunc("/", defaultHandler)
	
	// Start listening and serving.
	s := &http.Server{Addr: ":8080", Handler: mux}
	log.Fatal(s.ListenAndServe())
}
