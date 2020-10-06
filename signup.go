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

func signupHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")

	if isLoggedIn(r) {
		http.Redirect(w, r, "/", 307)
		return
	}

	tmpl, err := template.ParseFiles("templates/signup.html")
	if err != nil {
		serveError(w, err, "could not parse template 'signup.html'.")
		return
	}
	data := SignupTemplate{}

	if r.Method == http.MethodPost {
		err = r.ParseForm()
		if err != nil {
			serveError(w, err, "could not parse HTTP form 'signup'.")
			return
		}

		username := r.FormValue("username")
		password := r.FormValue("password")

		matches, err := regexp.MatchString("^[A-Za-z0-9_]*$", username)
		if err != nil {
			serveError(w, err, "could not check username against regex.")
			return
		} else if !matches {
			data.Message = "Usernames must only consist of " +
				"letters, numbers, and underscores. " +
				"Please try a different username."
			data.DefaultUsername = username
		} else if len(username) == 0 || len(username) >= 256 {
			data.Message = "Usernames must be between " +
				"1 and 255 characters in length (inclusive). " +
				"Please try a different username."
		} else {
			uid, err := findUID(username)
			if err != nil {
				serveError(w, err, "failed finding user data from username.")
				return
			} else if uid != 0 {
				data.Message = "That username already in use. " +
					"Please try a different username."
				data.DefaultUsername = username
			} else {
				hash, err := bcrypt.GenerateFromPassword([]byte(password), 0)
				if err != nil {
					serveError(w, err, "could not encrypt password")
					return
				}

				stmt := "INSERT INTO users(uid, username, hash) VALUES (NULL, ?, ?);"
				_, err = db.Exec(stmt, username, string(hash))
				if err != nil {
					serveError(w, err, "could not store user data in database")
					return
				}

				data.Message = "Profile created successfully! " +
					"Please <a href=\"/login\">login</a>."
			}
		}
	}
	
	tmpl.Execute(w, data)
}
