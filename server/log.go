package server

import "fmt"

func Log(s ...string) {
	fmt.Println("log srv: ", s)
}

func Logg(s ...interface{}) {
	fmt.Println("log srv: ", s)
}
