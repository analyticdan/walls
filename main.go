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
	w.Header().Set("Cache-Control", "no-store")

	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	if !isLoggedIn(r) {
		http.ServeFile(w, r, "static/index.html")
		return
	}

	uid, err := getUID(r)
	if err != nil {
		serveError(w, err, "could not get uid from cookies.")
		return
	}

	username, err := findUsername(uid)
	if err != nil {
		serveError(w, err, "could not match uid to username")
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/wall/%s", username), 307)
	return
}

func main() {
	var err error

	db, err = sql.Open("sqlite3", "database/data.db")
	if (err != nil) {
		log.Fatal(err)
	}
	defer db.Close()

	stmt := "CREATE TABLE IF NOT EXISTS users(uid INTEGER PRIMARY KEY, username VARCHAR(255), hash CHAR(60));"
	_, err = db.Exec(stmt)
	if (err != nil) {
		log.Fatal(err)
	}

	stmt = "CREATE TABLE IF NOT EXISTS cookies(uid INTEGER, sid CHAR(36) UNIQUE);"
	_, err = db.Exec(stmt)
	if (err != nil) {
		log.Fatal(err)
	}

	stmt = "CREATE TABLE IF NOT EXISTS posts(from_uid INTEGER, to_uid INTEGER, body VARCHAR(255), date DATETIME);"
	_, err = db.Exec(stmt)
	if (err != nil) {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/imgs/", http.FileServer(http.Dir(".")))
	mux.HandleFunc("/wall/", wallHandler)
	mux.HandleFunc("/signup", signupHandler)
	mux.HandleFunc("/login", loginHandler)
	mux.HandleFunc("/logout", logoutHandler)
	mux.HandleFunc("/", defaultHandler)

	s := &http.Server{Addr: ":80", Handler: mux}
	log.Fatal(s.ListenAndServe())
}
