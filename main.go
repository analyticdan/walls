package main

import (
	"log"
	"fmt"
	//"time"
	"strconv"
	"net/http"
	"database/sql"
	"html/template"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	_ "github.com/mattn/go-sqlite3"
)

/*
Implement cookies.
Use mutexes around DB. (Or transactions. Or whatever they taught in CS168/CS162.)
Implement schema for keeping track of messages.
*/

var db *sql.DB

/* Serves a server error message to the client. */
func serverError(w http.ResponseWriter, msg string) {
	http.Error(w, fmt.Sprintf("Internal server failure. Please try again.\nInternal error message: %s", msg), 500)
}

/*
	Attempts to put (UID, SID) into the cookies table of the database,
	where SID is a random 32 byte UUID (which takes up 36 bytes, after
	including hyphens).

	If insertion into the cookie table was sucessful, writes cookies to W.
	Otherwise, does nothing and returns a non-nil error.
*/
func setCookies(w http.ResponseWriter, uid int) error {
	// Generate new session id (UUID).
	sessionid, err := uuid.NewRandom()
	if err != nil {
		return err
	}
	sid := sessionid.String()
	// Insert cookies into cookie table.
	_, err = db.Exec("REPLACE INTO cookies(uid, sid) VALUES (?, ?);", uid, sid)
	if err != nil {
		return err
	}
	// Notify client of new cookies.
	http.SetCookie(w, &http.Cookie{Name: "uid", Value: strconv.Itoa(uid)})
	http.SetCookie(w, &http.Cookie{Name: "sid", Value: sid})
	return nil
}

/*
	Returns true and the uid of the current session if the server sucessfully
	confirms that the cookies from R are linked to a valid, current session.
	Returns false and some undefined int on errors and on invalid cookies.
*/
func validateCookies(r *http.Request) (bool, int) {
	// Get uid from cookies.
	uidCookie, err := r.Cookie("uid")
	if err != nil {
		return false, 0
	}
	uid, err := strconv.Atoi(uidCookie.Value)
	if err != nil {
		return false, 0
	}
	// Get sid from cookies.
	sidCookie, err := r.Cookie("sid")
	if err != nil {
		return false, 0
	}
	sid := sidCookie.Value
	// Query database to see if cookies are valid.
	rows, err := db.Query("SELECT uid FROM cookies WHERE uid = ? AND sid = ?;", uid, sid)
	if err != nil {
		return false, 0
	}
	defer rows.Close()
	return rows.Next(), uid
}

/*
	Serves the HTML page for signups.
*/
func signupHandler(w http.ResponseWriter, r *http.Request) {
	// If client is already signed in, redirect home.
	current, _ := validateCookies(r)
	if current {
		http.Redirect(w, r, "/", 307)
	}
	// Serve signup page by using template and filling in data.
	tmpl, err := template.ParseFiles("templates/signup.html")
	if err != nil {
		serverError(w, "Could not parse template: signup.html")
	}
	data := struct{
		Message template.HTML
		DefaultUsername string
	}{}
	// Customize HTML based on form data and do backend (on non-GET requests).
	if r.Method != http.MethodGet {
		err = r.ParseForm()
		if err != nil {
			serverError(w, "Could not parse HTTP form: signup")
			return
		}
		username := r.FormValue("username")
		password := r.FormValue("password")
		if len(username) >= 256 {
			data.Message = "Usernames cannot be longer than 255 characters."
		} else {
			rows, err := db.Query("SELECT * FROM users WHERE username = ?;", username)
			if err != nil {
				serverError(w, "Database failed searching for user")
				return
			}
			defer rows.Close()
			if rows.Next() {
				data.Message = "Username already in use. Please try a different username."
				data.DefaultUsername = username
			} else {
				hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
				if err != nil {
					serverError(w, "Server was unable to encrypt password")
					return
				}
				_, err = db.Exec("INSERT INTO users(uid, username, hash) VALUES (NULL, ?, ?);", username, string(hash))
				if err != nil {
					serverError(w, "Database could not store profile")
					return
				}
				data.Message = "Profile created successfully! Please <a href=\"/login\">login</a>."
			}
		}
	}
	tmpl.Execute(w, data)
}

/*
	Serves the HTML page for logins.
*/
func loginHandler(w http.ResponseWriter, r *http.Request) {
	// If client is already signed in, redirect home.
	current, _ := validateCookies(r)
	if current {
		http.Redirect(w, r, "/", 307)
	}
	// Serve login page by using template and filling in data.
	tmpl, err := template.ParseFiles("templates/login.html")
	if err != nil {
		serverError(w, "Could not parse template: login.html")
	}
	data := struct{
		Message         string
		DefaultUsername string
	}{}
	// Customize HTML based on form data and do backend (on non-GET requests).
	if r.Method != http.MethodGet {
		err = r.ParseForm()
		if err != nil {
			serverError(w, "Could not parse HTTP form: login")
			return
		}
		username := r.FormValue("username")
		password := r.FormValue("password")
		// Check database to see if username exists.
		rows, err := db.Query("SELECT uid, hash FROM users WHERE username = ?;", username)
		if err != nil {
			serverError(w, "Database failed searching for user")
			return
		}
		defer rows.Close()
		if rows.Next() {
			var uid int
			var hash string
			err = rows.Scan(&uid, &hash)
			if err != nil {
				serverError(w, "Database failed fetching user info")
				return
			} else if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil {
				rows.Close()
				err := setCookies(w, uid)
				if err != nil {
					serverError(w, "Database failed setting cookies")
					return
				}
				http.Redirect(w, r, "/", 307)
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
	current, uid := validateCookies(r)
	if r.URL.Path == "/" {
		if current {
			fmt.Fprintf(w, fmt.Sprintf("User id: %d already logged in", uid))
		} else {
			http.ServeFile(w, r, "templates/index.html")
		}
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
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS cookies(uid INTEGER, sid CHAR(36) UNIQUE);")
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