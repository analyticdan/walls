package main

import (
	"strconv"
	"net/http"
	"html/template"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

/*
	Attempts to put (UID, SID) into the cookies table of the database,
	where SID is a 36 char session id string.

	If the insertion into the cookie table was sucessful, this function
	writes the cookies to W.
	Otherwise, this function does nothing and returns a non-nil error.
*/
func setCookies(w http.ResponseWriter, uid int) error {
	// Generate new session id.
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
	http.SetCookie(w, &http.Cookie{
		Name: "uid",
		Value: strconv.Itoa(uid),
		Path: "/",
	})
	http.SetCookie(w, &http.Cookie{
		Name: "sid",
		Value: sid,
		Path: "/",
	})
	return nil
}

/* Fields to fill in the login template. */
type LoginTemplate struct {
	Message string
	DefaultUsername string
}

/* Serves the login page. */
func loginHandler(w http.ResponseWriter, r *http.Request) {
	// Set no-cache header.
	w.Header().Set("Cache-Control", "no-cache")
	// If the client is already logged in, redirect home.
	if validateCookies(r) == nil {
		http.Redirect(w, r, "/", 307)
		return
	}
	// Ready the login page template and the data struct.
	tmpl, err := template.ParseFiles("templates/login.html")
	if err != nil {
		serverError(w, err, "could not parse template 'signup.html'.")
		return
	}
	data := LoginTemplate{}
	// Receive form data on POST (and use it to populate the template).
	if r.Method == http.MethodPost {
		// Parse POST data.
		err = r.ParseForm()
		if err != nil {
			serverError(w, err, "could not parse form 'signup'.")
			return
		}
		username := r.FormValue("username")
		password := r.FormValue("password")
		// Query the database to see if credentials match a signed-up user.
		stmt := "SELECT uid, hash FROM users WHERE username = ?;"
		rows, err := db.Query(stmt, username)
		if err != nil {
			serverError(w, err, "failed searching database for user data.")
			return
		}
		defer rows.Close()
		// By default, assume the credentials are bad since we discard
		// the data struct if the credentials are valid.
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
			// We need rows closed here in order to set cookies later on.
			// It's okay if we double close because Close() is idempotent.
			rows.Close()
			// Check that the hashed password matches the value in the database.
			err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
			if err == nil {
				if setCookies(w, uid) == nil {
					http.Redirect(w, r, "/", 307)
				} else {
					serverError(w, err, "could not set cookies")
				}
			}
		}
	}
	tmpl.Execute(w, data)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	sid, err := getSID(r)
	if err == nil {
		_, _ = db.Exec("DELETE FROM cookies WHERE sid = ?", sid)
	}
	http.Redirect(w, r, "/", 307)
}
