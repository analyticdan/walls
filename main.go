package main

import (
	"fmt"
	"log"
	"regexp"
	"net/http"
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

/*
Add logout.
Standardize HTML header (no-cache).
Use mutexes around DB. (Or transactions. Or whatever they taught in CS168/CS162.)
*/

var db *sql.DB

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	isLoggedIn, uid := validateCookies(r)
	if r.URL.Path == "/" {
		if isLoggedIn {
			username, err := findUsername(uid)
			if err != nil {
				serverError(w, err, "could not match uid to username")
			} else {
				http.Redirect(w, r, fmt.Sprintf("/%s", username), 307)
			}
		} else {
			w.Header().Set("Cache-Control", "no-cache")
			http.ServeFile(w, r, "static/index.html")
		}
		return
	}
	// Handle wall.
	isWall, err := regexp.MatchString("^/[A-Za-z0-9_]+[/]?$", r.URL.Path)
	if err != nil {
		serverError(w, err, "could not match URL against regex.")
	} else if isWall {
		wallHandler(w, r)
	} else {
		http.NotFound(w, r)
	}
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
	mux.HandleFunc("/signup/", signupHandler)
	mux.HandleFunc("/login/", loginHandler)
	mux.HandleFunc("/", defaultHandler)
	// Start listening and serving.
	s := &http.Server{Addr: ":8080", Handler: mux}
	log.Fatal(s.ListenAndServe())
}
