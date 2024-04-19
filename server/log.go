package server

import "log"

func Log(s ...string) {
	log.Println("log srv: ", s)
}

func Logg(s ...interface{}) {
	log.Println("log srv: ", s)
}
