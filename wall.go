package main

import (
	"regexp"
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
	// Assert that the request is really for a wall.
	isWall, err := regexp.MatchString("^/wall/[A-Za-z0-9_]+[/]?$", r.URL.Path)
	if err != nil {
		serverError(w, err, "could not match URL against regex.")
		return
	} else if !isWall {
		http.NotFound(w, r)
		return
	}
	// Set no-cache header.
	w.Header().Set("Cache-Control", "no-cache")
	// Ready the login page template.
	tmpl, err := template.ParseFiles("templates/wall.html")
	if err != nil {
		serverError(w, err, "could not parse template 'wall.html'.")
		return
	}
	// Gather data for handling this HTTP request.
	isLoggedIn := validateCookies(r) == nil
	ownerUsername := strings.Split(r.URL.Path, "/")[2]
	ownerUID, err := findUID(ownerUsername)
	if err != nil {
		serverError(w, err, "failed locating wall's uid from username")
		return
	} else if ownerUID == 0 {
		http.ServeFile(w, r, "static/walldne.html")
		return
	}
	data := WallTemplate{LoggedIn: isLoggedIn, Owner: ownerUsername}
	// Process potential new wall posts on POST.
	if isLoggedIn && r.Method == http.MethodPost {
		visitorUID, err := getUID(r)
		if err != nil {
			serverError(w, err, "could not get UID from cookies.")
			return
		}
		err = r.ParseForm()
		if err != nil {
			serverError(w, err, "could not parse HTTP form 'wall'.")
			return
		}
		post := r.FormValue("post")
		if len(post) != 0 && len(post) < 141 {
			stmt := "INSERT INTO posts(from_uid, to_uid, body, date) VALUES (?, ?, ?, datetime('now'));"
			_, err = db.Exec(stmt, visitorUID, ownerUID, post)
			if err != nil {
				serverError(w, err, "could not store post in database.")
				return
			}
		}
	}
	// Populate page with up to 20 of the most recent wall posts.
	stmt := "SELECT from_uid, body FROM posts WHERE to_uid = ? ORDER BY date DESC LIMIT 20;"
	rows, err := db.Query(stmt, ownerUID)
	if err != nil {
		serverError(w, err, "could not fetch posts")
		return
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
