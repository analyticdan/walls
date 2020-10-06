package main

import (
	"strconv"
	"net/http"
	"html/template"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

/* Fields to fill in login template. */
type LoginTemplate struct {
	Message string
	DefaultUsername string
}

func setCookies(w http.ResponseWriter, uid int) {
	sessionid, err := uuid.NewRandom()
	if err != nil {
		return
	}
	sid := sessionid.String()

	_, err = db.Exec("REPLACE INTO cookies(uid, sid) VALUES (?, ?);", uid, sid)
	if err != nil {
		return
	}

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
}

/* Serves the login page. */
func loginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")

	if isLoggedIn(r) {
		http.Redirect(w, r, "/", 307)
		return
	}

	tmpl, err := template.ParseFiles("templates/login.html")
	if err != nil {
		serveError(w, err, "could not parse template 'signup.html'.")
		return
	}
	data := LoginTemplate{}

	if r.Method == http.MethodPost {
		err = r.ParseForm()
		if err != nil {
			serveError(w, err, "could not parse form 'signup'.")
			return
		}

		password := r.FormValue("password")
		username := r.FormValue("username")
		data.DefaultUsername = username

		stmt := "SELECT uid, hash FROM users WHERE username = ?;"
		rows, err := db.Query(stmt, username)
		if err != nil {
			serveError(w, err, "failed searching database for user data.")
			return
		}
		defer rows.Close()
		
		if rows.Next() {
			var uid int
			var hash string

			err = rows.Scan(&uid, &hash)
			if err != nil {
				serveError(w, err, "failed scanning database for user data")
				return
			}
			rows.Close()

			err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
			if err == nil {
				setCookies(w, uid)
				http.Redirect(w, r, "/", 307)
				return
			} else {
				data.Message = "Invalid username/password."
			}
		}
	}
	
	tmpl.Execute(w, data)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	sid, err := getSID(r)
	if err == nil {
		db.Exec("DELETE FROM cookies WHERE sid = ?", sid)
	}
	http.Redirect(w, r, "/", 307)
}
