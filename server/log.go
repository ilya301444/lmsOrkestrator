package server

import "fmt"

func Log(s ...string) {
	fmt.Println("log srv:", s)
}
