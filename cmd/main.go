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
	"last/agent"
	"last/front"
	"last/server"
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
	port := "8033"
	println(os.Args)
	if len(os.Args) < 2 {
		port = "8025"
	} else {
		port = os.Args[1]
	}

	ctx := context.Background()
	downAgent, err := agent.StartAgent(ctx, port)
	printEr(err)
	downFront, err := front.StartFront()
	printEr(err)
	downServ, downServ2, err := server.StartSrv(ctx)
	printEr(err)

	// При нажатии ctl C происходит подача данных в канал sigChan
	// и выход из прогаммы
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, os.Interrupt)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	fmt.Println("Press ctrl+C for Exit")
	<-sigChan

	go downAgent.GracefulStop()
	go downFront(ctx)
	go downServ(ctx)
	go downServ2.GracefulStop()
	time.Sleep(400 * time.Millisecond)
}
