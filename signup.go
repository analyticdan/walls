package main

import (
	"regexp"
	"net/http"
	"html/template"

	"golang.org/x/crypto/bcrypt"
)

/* Fields to fill in signup template. */
type SignupTemplate struct {
	Message template.HTML
	DefaultUsername string
}

/* Serves the signup page. */
func signupHandler(w http.ResponseWriter, r *http.Request) {
	// If client is already logged in, redirect home.
	loggedIn, _ := validateCookies(r)
	if loggedIn {
		http.Redirect(w, r, "/", 307)
		return
	}
	// Ready the login page template and the data struct.
	tmpl, err := template.ParseFiles("templates/signup.html")
	if err != nil {
		serverError(w, err, "could not parse template 'signup.html'.")
		return
	}
	data := SignupTemplate{}
	// Receive form data on POST (and use it to populate the template).
	if r.Method == http.MethodPost {
		// Parse POST data.
		err = r.ParseForm()
		if err != nil {
			serverError(w, err, "could not parse HTTP form 'signup'.")
			return
		}
		username := r.FormValue("username")
		password := r.FormValue("password")
		// Check that the username is alphanumeric + underscores.
		// Also check that the username fits in the database [1 and 256).
		// Password can be any characters whatsoever.
		ok, err := regexp.MatchString("^[A-Za-z0-9_]*$", username)
		if err != nil {
			serverError(w, err, "could not check username against regex.")
			return
		} else if !ok {
			data.Message = "Usernames must only consist of " +
				"letters, numbers, and underscores. " +
				"Please try a different username."
			data.DefaultUsername = username
		} else if len(username) == 0 || len(username) >= 256 {
			data.Message = "Usernames must be between " +
				"1 and 255 characters in length (inclusive). " +
				"Please try a different username."
		} else {
			// Check if the username is already in the database.
			uid, err := findUID(username)
			if err != nil {
				serverError(w, err, "failed finding user data from username.")
				return
			} else if uid != 0 {
				data.Message = "That username already in use. " +
					"Please try a different username."
				data.DefaultUsername = username
			} else {
				// User info is okay. Try to insert into the database.
				hash, err := bcrypt.GenerateFromPassword([]byte(password), 0)
				if err != nil {
					serverError(w, err, "could not encrypt password")
					return
				}
				stmt := "INSERT INTO users(uid, username, hash) VALUES (NULL, ?, ?);"
				_, err = db.Exec(stmt, username, string(hash))
				if err != nil {
					serverError(w, err, "could not store user data in database")
					return
				}
				data.Message = "Profile created successfully! " +
					"Please <a href=\"/login\">login</a>."
			}
		}
	}
	tmpl.Execute(w, data)
}
