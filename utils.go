package main

import (
	"log"
	"strconv"
	"net/http"
)

/* Serves a server error message to the client. */
func serverError(w http.ResponseWriter, err error, msg string) {
	log.Printf("Error: %s\nInternal error message: %s\n", msg, err)
	http.Error(w, "Internal server failure. Please try again.", 500)
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

/*
	Returns the username of the user with the given UID,
	if there were no errors in the process.
*/
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

/*
	Returns the uid of the user with the given USERNAME,
	if there were no errors in the process.
*/
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
