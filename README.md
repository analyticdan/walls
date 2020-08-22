# walls
walls is a toy social network written in Go... and HTML... But mostly Go. The front-end is just a means to an end! I swear!!!

walls is currently in deep, deep alpha development.

walls' main draw is the ability to write on other users' "walls." You can create an account and invites your friends to write on your wall! Perhaps you might write on your friends' wall as well! But let's not get too ahead of ourselves.

Aside from gaping security holes that must be fixed (which I might never get around to), a basic login system must be implemented so that the server knows which clients are logged in so that it can give accurate attribution to wall posts.
Also there are probably race conditions due to my refusal of using best practices when accessing the database. I blame the lack of sleep and extreme enthusiasm I had when I started this. You can feel free to blame it on whatever you'd like!
Finally, there is no actual wall and you cannot write on it. Neither can your friends write on your wall. Neither can you write on your friends' walls... You can scream into the ether if you'd like, I don't mind. Just do it quietly, please. People are sleeping.
In conclusion, all that's done so far is a shoddy sign-up/login system. Feel free to do with this code as you'd please. I'll be adding more functionality as soon as I can get some rest and think of more ideas.

If you want to run walls on your local network, clone this repo and run `go get main.go` to ensure you have the files necessary to connect the walls server application to your local sqlite3 installation. (I'll assume you know how to download Go and whatnot. There are some great tutorials out there if you don't. Absolutely wonderful.) You'll probably want to configure some of the constants in the `main` function of `main.go` to run off of port 80 and your IP. You'll also want to handle port forwarding so that all inbound traffic to your IP on port 80 gets to the walls server. Then, run `go run main.go` and watch as the Zuckerberg level lawsuits pile in (in no small part due to all the security vulnerabilities).

Please note that if you use this software and something bad happens, the writer of this software (hopefully) should not be held liable for damages. In fact, I have enough damages to deal with on my own from this software... Where should I address my complaints to?! Just kidding. But really, please don't sue me.

Ciao.
