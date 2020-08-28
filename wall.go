package main

import (
	"strings"
	"net/http"
	"html/template"
)

// Fields to populate a post within a wall template.
type Post struct {
	From string
	Body string
	Date string
}

// Fields to populate the wall template.
type WallTemplate struct {
	LoggedIn bool
	Owner string
	Posts []Post
}

/* Serves a user's wall. */
func wallHandler(w http.ResponseWriter, r *http.Request) {
	// Ready the login page template.
	tmpl, err := template.ParseFiles("templates/wall.html")
	if err != nil {
		serverError(w, err, "could not parse template 'wall.html'.")
		return
	}
	// Gather data for and initialize the data struct.
	loggedIn, visitorUID := validateCookies(r)
	ownerUsername := strings.Trim(r.URL.Path, "/")
	ownerUID, err := findUID(ownerUsername)
	if err != nil {
		serverError(w, err, "failed locating wall's uid from username")
		return
	} else if ownerUID == 0 {
		// This user Does Not Exist. Serve the static page saying such.
		http.ServeFile(w, r, "static/walldne.html")
	}
	data := WallTemplate{LoggedIn: loggedIn, Owner: ownerUsername}
	// Receive form data on POST (and store new posts, if necessary).
	if loggedIn && r.Method == http.MethodPost {
		// Parse post.
		err = r.ParseForm()
		if err != nil {
			serverError(w, err, "could not parse HTTP form 'wall'.")
			return
		}
		// It is okay to store this unescaped because the sql and http/template
		// libraries will escape it for us.
		post := r.FormValue("post")
		// Only store posts of 140 characters or less (and 1 or more).
		if len(post) >= 1 && len(post) < 141 {
			stmt := "INSERT INTO posts(from_uid, to_uid, body, date) VALUES (?, ?, ?, datetime('now'));"
			_, err = db.Exec(stmt, visitorUID, ownerUID, post)
			if err != nil {
				serverError(w, err, "could not store post in database.")
				return
			}
		}
	}
	// Populate data (and thereby web page) with up to 20
	// of the wall owner's most recent wall posts.
	stmt := "SELECT from_uid, body FROM posts WHERE to_uid = ? ORDER BY date DESC LIMIT ?;"
	rows, err := db.Query(stmt, ownerUID, 20)
	if err != nil {
		serverError(w, err, "could not fetch posts")
	}
	defer rows.Close()
	for rows.Next() {
		var fromUID int
		var body string
		err = rows.Scan(&fromUID, &body)
		if err != nil {
			serverError(w, err, "failed scanning database for posts")
			return
		}
		fromUsername, err := findUsername(fromUID)
		if err != nil {
			serverError(w, err, "could not find poster's username from uid")
			return
		}
		data.Posts = append(data.Posts, Post{From: fromUsername, Body: body})
	}
	tmpl.Execute(w, data)
}
