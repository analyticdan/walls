# walls
walls is a toy social network written in Go... and HTML... But mostly Go. (There is a lot more front-end work to go before this is anywhere near good.)

walls is currently in deep alpha development.

walls' main draw is the ability to write on other users' "walls." You can create an account and invites your friends to write on your wall! Perhaps you might write on your friends' wall as well!

Assuming you have go and sqlite3 installed on your local machine, if you want to run walls on your local network, you might have to first go get (pun intended!) some of the dependencies, namely the bcrypt package, the uuid package, and the sqlite3 driver package that are used in this project. Then, simply run `go run *.go` from inside the repo's directory. You can then access walls from `localhost:8080` in a web browser or from your other favorite method of browsing the net.

Please note that in previous versions of this README, the text here was much funnier (in my opinion), but I'm far too satisfied with my code here that I don't really feel the need to be funny. However, as always, please don't sue me if this software messes something up for you. This software is presented as is, and so on and so forth. I still need to figure out a way to give this a BSD copyright or something. Until then, please don't sue me.

Ciao.
