package main

// Программа делится на 3 части:
// 1 отображает фронт и обрабатывает запросы
// 2 server(оркестратор) получает запросы от первой части
// 	следит за временем из выполнения и хранит их в бд
// 3 agent выполняет задачи и передаёт обратно server

// данными части обмениваются через http

import (
	"context"
	"fmt"
	"last/front"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func printEr(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

func main() {

	ctx := context.Background()
	downFront, err := front.StartFront(ctx)
	printEr(err)

	fmt.Println(0)
	// При нажатии ctl C происходит подача данных в канал sigChan
	// и выход из прогаммы
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	fmt.Println("Press ctrl+C for Exit")
	<-sigChan

	downFront(ctx)

	time.Sleep(time.Microsecond)
}
