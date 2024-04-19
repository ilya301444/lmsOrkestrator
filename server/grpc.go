package server

import (
	"context"
	"last/agent"
	pb "last/agent/proto"
	"last/front"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Server struct {
	pb.MessegeServer
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) TaskAgent(ctx context.Context, in *pb.Agent) (*pb.Zero, error) {
	var ag agent.Agent
	ag.Id = int(in.Id)
	ag.Loacaladdr = in.Loacaladdr
	ag.Status = int(in.Status)

	orkestr.mu.Lock()
	_, ok := orkestr.agents[ag.Loacaladdr] // что бы отображались только новые агенты
	orkestr.agents[ag.Loacaladdr] = &ag    //записываем нового агента если есть обновляем
	orkestr.mu.Unlock()
	addrAgent := ag.Loacaladdr

	//если действительно новый агент
	if !ok {
		orkestr.agentUpdate()
		Log("new Agent: ", ag.Loacaladdr)
	}

	var tsk *front.Task
	//если есть таски в очереди берём 1 таску
	orkestr.mu.Lock()
	if len(orkestr.Tasks) != 0 {
		tsk = orkestr.Tasks[0]
		if len(orkestr.Tasks) > 1 {
			orkestr.Tasks = orkestr.Tasks[1:]
		} else {
			orkestr.Tasks = nil
		}
	}
	orkestr.mu.Unlock()

	//если были таски в очереди
	if tsk != nil {
		orkestr.mu.Lock()
		orkestr.taskInWork[tsk.Id] = &TaskInWork{tsk, ag.Loacaladdr}
		orkestr.mu.Unlock()

		//front.Send(tsk, "http://localhost"+addrAgent+"/newTask")
		sendTask(addrAgent, tsk)
		Log("Task Sending to agent")
	}
	return &pb.Zero{}, nil
}

func (s *Server) HertBit(ctx context.Context, in *pb.Agent) (*pb.Zero, error) {
	//получили запрос от агента и распасим его данные
	var ag agent.Agent
	ag.Id = int(in.Id)
	ag.Loacaladdr = in.Loacaladdr
	ag.Status = int(in.Status)

	orkestr.mu.Lock()
	orkestr.agents[ag.Loacaladdr] = &ag //записываем нового агента если есть обновляем
	orkestr.mu.Unlock()
	return &pb.Zero{}, nil
}

func (s *Server) AnswerTask(ctx context.Context, in *pb.Task) (*pb.Zero, error) {
	tsk := front.Task{}
	tsk.Id = int(in.Id)
	tsk.Expression = in.Expression
	tsk.ValidExp = in.ValidExp
	tsk.Time = int(in.Time)
	tsk.Status = int(in.Status)
	tsk.Result = int(in.Result)

	id := tsk.Id
	orkestr.mu.Lock()
	delete(orkestr.taskInWork, id)
	orkestr.mu.Unlock()

	Logg("получили ответ на таску", tsk)
	front.Send(&tsk, "http://"+addrFront+"/getAnswer")
	return &pb.Zero{}, nil
}

// сервер не реализует
func (s *Server) UpdateTimeLimit(ctx context.Context, in *pb.TimeLimit) (*pb.Zero, error) {
	return &pb.Zero{}, nil
}

// сервер не реализует
func (s *Server) TaskToAgent(ctx context.Context, in *pb.Task) (*pb.Zero, error) {
	return &pb.Zero{}, nil
}

func SrvGrpcStart(adrSrv string) *grpc.Server {
	//adrSrv := "localhost:7000"
	list, err := net.Listen("tcp", adrSrv)
	Logg(err)
	Log("start grpc server ")

	grpcServer := grpc.NewServer()
	agentServer := NewServer()
	pb.RegisterMessegeServer(grpcServer, agentServer)
	go func() {
		if err := grpcServer.Serve(list); err != nil {
			Logg(err)
		}
		orkestr.saveCondact()
	}()
	return grpcServer
}

// клиентские функции для отправки данных агенту
func sendTask(addrAgent string, tsk *front.Task) error {
	con, err := grpc.Dial(addrAgent, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
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

	_, err = grpcClient.TaskToAgent(ctx, &tskSend)
	if err != nil {
		return err
	}
	return nil
}

func sendTime(addrAgent string, newTm *front.Operation) error {
	con, err := grpc.Dial(addrAgent, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer con.Close()

	grpcClient := pb.NewMessegeClient(con)
	ctx := context.TODO()

	tm := pb.TimeLimit{}
	tm.Plus = int32(newTm.Plus)
	tm.Minus = int32(newTm.Minus)
	tm.Mul = int32(newTm.Mul)
	tm.Div = int32(newTm.Div)
	tm.All = int32(newTm.All)

	_, err = grpcClient.UpdateTimeLimit(ctx, &tm)
	if err != nil {
		return err
	}
	return nil
}
