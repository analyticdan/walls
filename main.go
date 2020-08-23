package main

import (
	"log"
	"fmt"
	"sync"
	"net/http"
	"database/sql"

	"golang.org/x/crypto/bcrypt"
	//"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

/*
Implement cookies.
Use mutexes around DB. (Or transactions. Or whatever they taught in CS168/CS162.)
Implement schema for keeping track of messages.
*/

var db *sql.DB

func signupHandler(w http.ResponseWriter, r *http.Request) {
	// On GET, send HTML form.
	if r.Method == http.MethodGet {
		http.ServeFile(w, r, "./resources/signup.html")
		return
	}

	// On POST, check form.
	if r.ParseForm() != nil {
		http.Error(w, "Internal server failure (parsing HTTP form). Please try again.", 500)
		return
	}
	user := r.FormValue("user")
	pass := r.FormValue("pass")
	if len(pass) == 0 {
		http.Error(w, "Please enter a password.", 400)
		return
	} else if len(user) == 0 {
		http.Error(w, "Please enter a username.", 400)
		return
	} else if len(user) >= 256 {
		http.Error(w, "Usernames cannot be longer than 255 characters. Please try again.", 400)
		return
	}

	// Query database to see if user already exists.
	rows, err := db.Query("SELECT * FROM users WHERE username = ?;", user)
	if err != nil {
		http.Error(w, "Internal database failure (searching database for username). Please try again.", 500)
		return
	}
	defer rows.Close()
	if rows.Next() {
		http.Error(w, "Username already in use. Please try again.", 400)
		return
	}

	// Insert new user into database.
	hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Internal server failure (hashing). Please try again.", 500)
		return
	}
	_, err = db.Exec("INSERT INTO users(uid, username, hash) VALUES (NULL, ?, ?);", user, string(hash))
	if err != nil {
		http.Error(w, "Internal database failure (creating profile). Please try again.", 500)
		fmt.Println(err)
		return
	}
	http.Redirect(w, r, "", 307)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	// On GET, send HTML form.
	if r.Method == http.MethodGet {
		http.ServeFile(w, r, "./resources/login.html")
		return
	}

	// On POST, check form.
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Internal server failure. Please try again.", 500)
		return
	}
	user := r.FormValue("user")
	pass := r.FormValue("pass")

	// Query database to see if user exists.
	rows, err := db.Query("SELECT uid, hash FROM users WHERE username = ?;", user)
	if err != nil {
		http.Error(w, "Internal database failure (finding user). Please try again.", 500)
		return
	}
	defer rows.Close()
	if !rows.Next() {
		http.Error(w, "Username not found. Please try again.", 200)
		return
	}
	// Verify password matches.
	var uid, hash string
	err = rows.Scan(&uid, &hash)
	if err != nil {
		http.Error(w, "Internal server/database failure (fetching user records). Please try again.", 500)
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(pass))
	if err != nil {
		http.Error(w, "Invalid password. Please try again.", 401)
		return
	}
	fmt.Fprintf(w, "You're in.")
	//login.... Zzzz...
	//http.Redirect(w, r, "", 307)
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
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS users(uid INTEGER PRIMARY KEY, username VARCHAR(255), hash CHAR(60));")
	if (err != nil) {
		log.Fatal(err)
	}

	// Create cookie table if it doesn't exist.
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS cookies(uid INTEGER, session_id CHAR(36));")
	if (err != nil) {
		log.Fatal(err)
	}

	// Register handlers.
	mux := http.NewServeMux()
	mux.HandleFunc("/signup", signupHandler)
	mux.HandleFunc("/login", loginHandler)
	mux.HandleFunc("/", defaultHandler)

	// Start listening and serving.
	s := &http.Server{Addr: ":8080", Handler: mux}
	log.Fatal(s.ListenAndServe())
}