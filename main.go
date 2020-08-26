package main

import (
	"log"
	"fmt"
	//"sync"
	"net/http"
	"database/sql"
	"html/template"

	"golang.org/x/crypto/bcrypt"
	//"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

/*
Implement cookies.
Use mutexes around DB. (Or transactions. Or whatever they taught in CS168/CS162.)
Implement schema for keeping track of messages.
*/

func serverError(w http.ResponseWriter, code int, msg string) {
	msg = fmt.Sprintf("Internal server failure. Please try again.\n" +
		              "Internal error message:\n%s", msg)
	http.Error(w, msg, code)
}

// Checks if a user with the username USERNAME exists in the database.
func userExists(username string) (exists bool, err error) {
	rows, err := db.Query("SELECT * FROM users WHERE username = ?;", username)
	if err != nil {
		return
	}
	exists = rows.Next()
	rows.Close()
	return
}

var db *sql.DB
func signupHandler(w http.ResponseWriter, r *http.Request) {
	// Fetch signup template.
	tmpl, err := template.ParseFiles("templates/signup.html")
	if err != nil {
		serverError(w, 500, "Could not parse template: signup.html")
	}
	// Initialize data with which to fill out the template.
	data := struct{
		Message         template.HTML
		DefaultUsername string
	}{}
	// On non-GET, try to personalize the page.
	if r.Method != http.MethodGet {
		err = r.ParseForm()
		if err != nil {
			serverError(w, 500, "Could not parse HTTP form: signup")
			return
		}
		username := r.FormValue("username")
		password := r.FormValue("password")
		if len(username) >= 256 {
			data.Message = "Usernames cannot be longer than 255 characters."
		} else {
			// Check database to see if username is taken.
			rows, err := db.Query("SELECT * FROM users WHERE username = ?;", username)
			if err != nil {
				serverError(w, 500, "Database failed searching for user")
				return
			}
			defer rows.Close()
			if rows.Next() {
				data.Message = "Username already in use. Please try a different username."
				data.DefaultUsername = username
			} else {
				// Insert new user into database.
				hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
				if err != nil {
					serverError(w, 500, "Server was unable to encrypt password")
					return
				}
				_, err = db.Exec("INSERT INTO users(uid, username, hash) VALUES (NULL, ?, ?);", username, string(hash))
				if err != nil {
					serverError(w, 500, "Database could not store profile")
					return
				}
				data.Message = "Profile created successfully! Please <a href=\"/login\">login</a>."
			}
		}
	}
	tmpl.Execute(w, data)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/login.html")
	if err != nil {
		serverError(w, 500, "Could not parse template: login.html")
	}
	// Initialize data with which to fill out the template.
	data := struct{
		Message         string
		DefaultUsername string
	}{}
	// On non-GET, try to personalize the page.
	if r.Method != http.MethodGet {
		err = r.ParseForm()
		if err != nil {
			serverError(w, 500, "Could not parse HTTP form: login")
			return
		}
		username := r.FormValue("username")
		password := r.FormValue("password")
		// Check database to see if username exists.
		rows, err := db.Query("SELECT uid, hash FROM users WHERE username = ?;", username)
		if err != nil {
			serverError(w, 500, "Database failed searching for user")
			return
		}
		defer rows.Close()
		// Check database to see if hashes match.
		if rows.Next() {
			var uid, hash string
			err = rows.Scan(&uid, &hash)
			if err != nil {
				serverError(w, 500, "Database failed fetching user info")
				return
			} else if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil {
				http.Redirect(w, r, "/"+ username, 307)
				return
			} else {
				data.Message = "Invalid username/password."
				data.DefaultUsername = username
			}
		} else {
			data.Message = "Invalid username/password."
			data.DefaultUsername = username
		}
	}
	tmpl.Execute(w, data)
}

func defaultHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		http.ServeFile(w, r, "templates/index.html")
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