package main

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"net/http"
	"database/sql"
	"html/template"

	"golang.org/x/crypto/bcrypt"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

/*
Add logout.
Standardize HTML header.
Use mutexes around DB. (Or transactions. Or whatever they taught in CS168/CS162.)
*/

var db *sql.DB

/* Serves a server error message to the client. */
func serverError(w http.ResponseWriter, err error, msg string) {
	log.Printf("Error: %s\nInternal error message: %s\n", msg, err)
	http.Error(w, "Internal server failure. Please try again.", 500)
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
	http.SetCookie(w, &http.Cookie{Name: "uid", Value: strconv.Itoa(uid), Path: "/"})
	http.SetCookie(w, &http.Cookie{Name: "sid", Value: sid, Path: "/"})
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
	rows, err := db.Query("SELECT * FROM cookies WHERE uid = ? AND sid = ?;", uid, sid)
	if err != nil {
		return false, 0
	}
	defer rows.Close()
	return rows.Next(), uid
}

/* Fields to fill in signup template. */
type SignupTemplate struct {
	Message template.HTML
	DefaultUsername string
}

/* Serves the HTML page for signups. */
func signupHandler(w http.ResponseWriter, r *http.Request) {
	// If client is already signed in, redirect home.
	isLoggedIn, _ := validateCookies(r)
	if isLoggedIn {
		http.Redirect(w, r, "/", 307)
		return
	}
	// Serve signup page by using template and filling in data.
	tmpl, err := template.ParseFiles("templates/signup.html")
	if err != nil {
		serverError(w, err, "could not parse template 'signup.html'.")
		return
	}
	data := SignupTemplate{}
	// Receive form data on POST (and use it to populate the template).
	if r.Method == http.MethodPost {
		err = r.ParseForm()
		if err != nil {
			serverError(w, err, "could not parse HTTP form 'signup'.")
			return
		}
		username := r.FormValue("username")
		password := r.FormValue("password")
		ok, err := regexp.MatchString("^[A-Za-z0-9_]*$", username)
		if err != nil {
			serverError(w, err, "could not check username against regex.")
			return
		} else if !ok {
			data.Message = "Usernames must only consist of " +
				"letters, numbers, and underscores.<br>" +
				"Please try a different username."
			data.DefaultUsername = username
		} else if len(username) == 0 || len(username) >= 256 {
			data.Message = "Usernames must be between " +
				"1 and 255 characters in length (inclusive).<br>" +
				"Please try a different username."
		} else {
			rows, err := db.Query("SELECT * FROM users WHERE username = ?;", username)
			if err != nil {
				serverError(w, err, "failed searching database for user data.")
				return
			}
			defer rows.Close()
			if rows.Next() {
				data.Message = "That username already in use. " +
					"Please try a different username."
				data.DefaultUsername = username
			} else {
				hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
				if err != nil {
					serverError(w, err, "could not encrypt password")
					return
				}
				_, err = db.Exec("INSERT INTO users(uid, username, hash) VALUES (NULL, ?, ?);", username, string(hash))
				if err != nil {
					serverError(w, err, "could not store user data in database")
					return
				}
				data.Message = "Profile created successfully!<br>" +
					"Please <a href=\"/login\">login</a>."
			}
		}
	}
	tmpl.Execute(w, data)
}

/* Fields to fill in login template. */
type LoginTemplate struct {
	Message string
	DefaultUsername string
}

/* Serves the HTML page for logins. */
func loginHandler(w http.ResponseWriter, r *http.Request) {
	// If client is already signed in, redirect home.
	isLoggedIn, _ := validateCookies(r)
	if isLoggedIn {
		http.Redirect(w, r, "/", 307)
		return
	}
	// Serve login page by using template and filling in data.
	tmpl, err := template.ParseFiles("templates/login.html")
	if err != nil {
		serverError(w, err, "could not parse template 'signup.html'.")
		return
	}
	data := LoginTemplate{}
	// Receive form data on POST (and use it to populate the template).
	if r.Method == http.MethodPost {
		err = r.ParseForm()
		if err != nil {
			serverError(w, err, "could not parse HTTP form 'signup'.")
			return
		}
		username := r.FormValue("username")
		password := r.FormValue("password")
		rows, err := db.Query("SELECT uid, hash FROM users WHERE username = ?;", username)
		if err != nil {
			serverError(w, err, "failed searching database for user data.")
			return
		}
		defer rows.Close()
		data.Message = "Invalid username/password."
		data.DefaultUsername = username
		if rows.Next() {
			var uid int
			var hash string
			err = rows.Scan(&uid, &hash)
			if err != nil {
				serverError(w, err, "failed scanning database for user data")
				return
			}
			err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
			if err == nil {
				rows.Close()
				err := setCookies(w, uid)
				if err != nil {
					serverError(w, err, "could not set cookies")
					return
				}
				http.Redirect(w, r, "/", 307)
				return
			}
		}
	}
	tmpl.Execute(w, data)
}

type Post struct {
	From string
	Body string
	Date string
}

type WallTemplate struct {
	LoggedIn bool
	Owner string
	Posts []Post
}

/* Serves HTML page for walls. */
func wallHandler(w http.ResponseWriter, r *http.Request) {
	// Serve login page by using template and filling in data.
	tmpl, err := template.ParseFiles("templates/wall.html")
	if err != nil {
		serverError(w, err, "could not parse template 'wall.html'.")
		return
	}
	// Gather necessary data to populate template.
	isLoggedIn, visitorUID := validateCookies(r)
	ownerUsername := strings.Trim(r.URL.Path, "/")
	ownerUID, err := findUID(ownerUsername)
	if err != nil {
		serverError(w, err, "could not find wall's uid from username")
		return
	} else if ownerUID == 0 {
		http.ServeFile(w, r, "static/walldne.html")
	}
	// Initialize data struct with which to fill out template.
	data := WallTemplate{LoggedIn: isLoggedIn, Owner: ownerUsername}
	// Receive form data on POST (and store new posts, if necessary).
	if isLoggedIn && r.Method == http.MethodPost {
		err = r.ParseForm()
		if err != nil {
			serverError(w, err, "could not parse HTTP form 'wall'.")
			return
		}
		post := r.FormValue("post")
		if len(post) >= 1 && len(post) < 141 {
			_, err = db.Exec("INSERT INTO posts(from_uid, to_uid, body, date) VALUES (?, ?, ?, datetime('now'));", visitorUID, ownerUID, post)
			if err != nil {
				serverError(w, err, "could not store post in database.")
				return
			}
		}
	}
	// Populate data with wall posts.
	rows, err := db.Query("SELECT from_uid, body, date FROM posts WHERE to_uid = ? ORDER BY date DESC LIMIT ?;", ownerUID, 20)
	if err != nil {
		serverError(w, err, "could not fetch posts")
	}
	defer rows.Close()
	for rows.Next() {
		var fromUID int
		var body string
		var date string
		err = rows.Scan(&fromUID, &body, &date)
		if err != nil {
			serverError(w, err, "failed scanning database for posts")
			return
		}
		fromUsername, err := findUsername(fromUID)
		if err != nil {
			serverError(w, err, "could not find poster's username from uid")
			return
		}
		data.Posts = append(data.Posts, Post{From: fromUsername, Body: body, Date: date})
	}
	tmpl.Execute(w, data)
}

func findUsername(uid int) (username string, err error) {
	rows, err := db.Query("SELECT username FROM users WHERE uid = ? LIMIT 1;", uid)
	if err != nil {
		return
	}
	defer rows.Close()
	if rows.Next() {
		err = rows.Scan(&username)
	}
	return
}

func findUID(username string) (uid int, err error) {
	rows, err := db.Query("SELECT uid FROM users WHERE username = ? LIMIT 1;", username)
	if err != nil {
		return
	}
	defer rows.Close()
	if rows.Next() {
		err = rows.Scan(&uid)
	}
	return
}

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