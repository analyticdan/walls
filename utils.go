package main

import (
	"log"
	"strconv"
	"net/http"
)

func serveError(w http.ResponseWriter, err error, msg string) {
	log.Printf("Error: %s\nInternal error message: %s\n", msg, err)
	http.Error(w, "Internal server failure. Please try again.", 500)
}

func isLoggedIn(r *http.Request) bool {
	uid, err := getUID(r)
	if err != nil {
		return false
	}

	sid, err := getSID(r)
	if err != nil {
		return false
	}

	stmt := "SELECT * FROM cookies WHERE uid = ? AND sid = ?;"
	rows, err := db.Query(stmt, uid, sid)
	if err != nil {
		return false
	}
	defer rows.Close()

	return rows.Next()
}

func getUID(r *http.Request) (uid int, err error) {
	uidCookie, err := r.Cookie("uid")
	if err != nil {
		return
	}
	uid, err = strconv.Atoi(uidCookie.Value)
	return
}

func getSID(r *http.Request) (sid string, err error) {
	sidCookie, err := r.Cookie("sid")
	if err != nil {
		return
	}
	sid = sidCookie.Value
	return
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
