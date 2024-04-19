package agent

import (
	"context"
	"fmt"
	pb "last/agent/proto"
	"last/front"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Server struct {
	pb.MessegeServer
}

func NewServer() *Server {
	return &Server{}
}

// не реализуется в агенте
func (s *Server) TaskAgent(ctx context.Context, in *pb.Agent) (*pb.Zero, error) {
	return &pb.Zero{}, nil
}

// не реализуется в агенте
func (s *Server) HertBit(ctx context.Context, in *pb.Agent) (*pb.Zero, error) {
	return &pb.Zero{}, nil
}

// не реализуется в агенте
func (s *Server) AnswerTask(ctx context.Context, in *pb.Task) (*pb.Zero, error) {
	return &pb.Zero{}, nil
}

func (s *Server) UpdateTimeLimit(ctx context.Context, in *pb.TimeLimit) (*pb.Zero, error) {
	timeOper := front.Operation{}
	timeOper.Plus = int(in.Plus)
	timeOper.Minus = int(in.Minus)
	timeOper.Mul = int(in.Mul)
	timeOper.Div = int(in.Div)
	timeOper.All = int(in.All)

	Log("update time limit")
	agent.mu.Lock()
	agent.operLimit = timeOper
	agent.mu.Unlock()
	agent.reboot() // перезагружаем и удаляем все таски
	return &pb.Zero{}, nil
}

// принимаем таску
func (s *Server) TaskToAgent(ctx context.Context, in *pb.Task) (*pb.Zero, error) {
	tsk := &front.Task{}
	tsk.Id = int(in.Id)
	tsk.Expression = in.Expression
	tsk.ValidExp = in.ValidExp
	tsk.Time = int(in.Time)
	tsk.Status = int(in.Status)
	tsk.Result = int(in.Result)

	//запускаем задачу на выполнение в отдельной горутине
	go func() {
		agent.ExecuteTask(tsk)
		// время после которого будет выполнена задача (считаем это время незначительным по сравнению с секундой)
		timeStop := time.Duration(agent.getTimeLimit(tsk.Expression)) * time.Second
		select {
		//если пришёл сигнал на перезагрузку (изменилось время выполнения операций +-*/)
		case <-agent.stop:
			Log("stoped task")
		case <-time.After(timeStop):
			sendAnswer(tsk)
		}

		agent.mu.Lock()
		agent.numTread++
		agent.mu.Unlock()

		Logg("task is continue! ", tsk)
	}()
	return &pb.Zero{}, nil
}

func SrvGrpcStart(adrSrv string) *grpc.Server {
	//adrSrv := "localhost:7000"
	list, err := net.Listen("tcp", adrSrv)
	FatalEr(err)
	Log("start grpc server ")

	grpcServer := grpc.NewServer()
	agentServer := NewServer()
	pb.RegisterMessegeServer(grpcServer, agentServer)
	go func() {
		if err := grpcServer.Serve(list); err != nil {
			FatalEr(err)
		}
		fmt.Println("agent srv stoped")
		agent.stopAgent <- struct{}{}
	}()

	startHertBit()
	startGetTask()
	return grpcServer
}

// клиентские функции отсылают данные
func startHertBit() {
	// горутина которая говорит что данный агент всё ещё жив
	//heartBit
	go func() {
		for {
			time.Sleep(timeSleep)

			con, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
			FatalEr(err)
			grpcClient := pb.NewMessegeClient(con)
			ctx := context.TODO()

			agent.mu.RLock()
			data := agent
			agent.mu.RUnlock()

			sendAg := pb.Agent{}
			sendAg.Id = int32(data.Id)
			sendAg.Loacaladdr = data.Loacaladdr
			sendAg.Status = int32(data.Status)

			_, err = grpcClient.HertBit(ctx, &sendAg)
			con.Close()
			PrintEr(err)

			select {
			case <-agent.stopAgent:
				Log("agent 2 close")
				return
			default:
			}
		}
	}()
}

func startGetTask() {
	//горутина по получению задач постоянно отправляем запрос если есть потоки и
	//нам приходят задачи через пост в другую функцию
	go func() {
		for {
			time.Sleep(timeSleep)

			con, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
			FatalEr(err)
			grpcClient := pb.NewMessegeClient(con)
			ctx := context.TODO()

			agent.mu.Lock()
			num := agent.numTread
			agent.mu.Unlock()
			// запрашиваем таску если не заняты все потоки
			if num > 0 {
				agent.mu.RLock()
				data := agent
				agent.mu.RUnlock()

				sendAg := pb.Agent{}
				sendAg.Id = int32(data.Id)
				sendAg.Loacaladdr = data.Loacaladdr
				sendAg.Status = int32(data.Status)

				if err != nil {
					PrintEr(err)
				}

				//запросили задачу
				_, err = grpcClient.TaskAgent(ctx, &sendAg)
				if err != nil {
					PrintEr(err)
				}
			}
			con.Close()

			select {
			case <-agent.stopAgent:
				Log("agent 1 close")
				return
			default:
			}
		}
	}()
}

func sendAnswer(tsk *front.Task) {
	con, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	PrintEr(err)
	defer con.Close()

	grpcClient := pb.NewMessegeClient(con)
	ctx := context.TODO()

	tskSend := pb.Task{}
	tskSend.Id = int32(tsk.Id)
	tskSend.Expression = tsk.Expression
	tskSend.ValidExp = tsk.ValidExp
	tskSend.Time = int32(tsk.Time)
	tskSend.Status = int32(tsk.Status)
	tskSend.Result = int64(tsk.Result)

	_, err = grpcClient.AnswerTask(ctx, &tskSend)
	PrintEr(err)
}
