package main

import (
	"log"
	"fmt"
	"net/http"
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

/*
Implement cookies.
Use mutexes around DB. (Or transactions. Or whatever they taught in CS168/CS162.)
For goodness sakes, please stop storing passwords as plaintext!@!
Implement schema for keeping track of messages.
*/

var db *sql.DB

func loginHandler(w http.ResponseWriter, r *http.Request) {
	// On GET, send HTML form.
	if r.Method == http.MethodGet {
		http.ServeFile(w, r, "resources/login.html")
		return
	}

	// On POST, check form.
	if r.ParseForm() != nil {
		http.Error(w, "Internal server failure. Please try again.", 500)
		return
	}
	user := r.FormValue("user")
	pass := r.FormValue("pass")

	// Query database to see if user exists.
	rows, err := db.Query("SELECT password FROM users WHERE username = ?;", user)
	if err != nil {
		http.Error(w, "Internal database failure. Please try again.", 500)
		return
	}
	defer rows.Close()
	
	// Check if user exists.
	if rows.Next() {
		var s string
		if rows.Scan(&s) != nil {
			http.Error(w, "Internal server/database failure. Please try again.", 500)
		} else if (s == pass) {
			fmt.Fprintf(w, "You're in.")
		} else {
			fmt.Fprintf(w, "Incorrect password.")
		}
		return
	}
	http.Error(w, "User not found. Please try again.", 200)
}

func signupHandler(w http.ResponseWriter, r *http.Request) {
	// On GET, send HTML form.
	if r.Method == http.MethodGet {
		http.ServeFile(w, r, "resources/signup.html")
		return
	}

	// On POST, check form.
	if r.ParseForm() != nil {
		http.Error(w, "Internal server failure. Please try again.", 500)
		return
	}
	user := r.FormValue("user")
	pass := r.FormValue("pass")

	// Check that the username doesn't already exist.
	rows, err := db.Query("SELECT password FROM users WHERE username = ?;", user)
	if err != nil {
		http.Error(w, "Internal database failure. Please try again.", 500)
		return
	} else if rows.Next() {
		http.Error(w, "Username taken. Please try again.", 200)
		return
	}

	// Insert into table.
	_, err = db.Exec("INSERT INTO users(username, password) VALUES (?, ?);", user, pass)
	/* Need to hash and salt passwords. This is terribly insecure, I know. */

	// Automatically log in.
	fmt.Fprintf(w, "You're in.")
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		http.ServeFile(w, r, "resources/index.html")
	} else {
		fmt.Fprintf(w, r.URL.Path)
		/* Serve profiles. Don't just print the path. */
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
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS users(username, password);")
	if (err != nil) {
		log.Fatal(err)
	}

	/* Implement sessions and cookies. Probably here. */

	// Register handlers.
	mux := http.NewServeMux()
	mux.HandleFunc("/login", loginHandler)
	mux.HandleFunc("/signup", signupHandler)
	mux.HandleFunc("/", defaultHandler)

	// Start listening and serving.
	s := &http.Server{
			Addr: ":8080",
			Handler: mux,
	}
	log.Fatal(s.ListenAndServe())
}