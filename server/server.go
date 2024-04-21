// server
// оркестратор
// Сервер, который принимает арифметическое выражение,
// переводит его в набор последовательных задач и обеспечивает
// порядок их выполнения. Далее будем называть его оркестратором.
// Функции оркестрации выполняют MainOrkestrator и getTask
/*
localhost:7010
localhost:8010 http server
*/
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"last/agent"
	"last/front"
	"net/http"
	"strconv"
	"sync"
	"time"

	"google.golang.org/grpc"
)

type TaskInWork struct {
	task  *front.Task
	agent string // addr agent
}

type Orkestrator struct {
	// по адресу находим нужный агент
	agents     map[string]*agent.Agent //список серверов(агентов ) готовых выполнить задачу
	Tasks      []*front.Task           //очередь из задач
	taskInWork map[int]*TaskInWork     // task.Id TaskInWork
	TimeLimit  int                     //тк  всё вычисляется на агенте и мы только отсылаем задачу мы можем знать
	//передельное время

	mu sync.Mutex
}

var (
	orkestr      Orkestrator
	addrFront    = "localhost:8000"
	serverStatus = 2
	fileBackup   = "./server/serverSave.txt"
)

func init() {
	orkestr.agents = make(map[string]*agent.Agent)
	orkestr.taskInWork = make(map[int]*TaskInWork)
	orkestr.TimeLimit = 200 // по умолчанию
}

// стартуем сервер и бдем слушать запросы
func StartSrv(ctx context.Context) (func(context.Context) error, *grpc.Server, error) {
	serverMux := http.NewServeMux()
	restoreCondact()

	serverMux.HandleFunc("/", orkestr.newTask)

	serverMux.HandleFunc("/statusSrv", orkestr.statusSrv)
	srv := &http.Server{Addr: ":8010", Handler: serverMux}
	go func() {
		err := srv.ListenAndServe()
		orkestr.saveCondact()
		agent.PrintEr(err)
	}()

	orkestr.MainOrkestrator()

	srv2 := SrvGrpcStart("localhost:7010")
	return srv.Shutdown, srv2, nil
}

// func StartSrv(ctx context.Context) (*grpc.Server, error) {
// 	var srv *grpc.Server
// 	restoreCondact()
// 	srv = SrvGrpcStart("localhost:7010")
// 	orkestr.MainOrkestrator()
// 	Log("started server")
// 	return srv, nil
// }

// server - agent
// отдаём задачу агенту
// агент спаминт нас запросами мы ему отдаём таски
// это функция по получению Хёртбита и добавления новых агентов и отдачи таски
// новому агенту
/*
func (o *Orkestrator) getTask(w http.ResponseWriter, r *http.Request) {
	body := r.Body
	data, err := io.ReadAll(body)
	if err != nil {
		agent.PrintEr(err)
		return
	}

	//получили запрос от агента и распасим его данные
	var ag agent.Agent
	if err := json.Unmarshal(data, &ag); err != nil {
		agent.PrintEr(err)
		return
	}
	o.mu.Lock()
	_, ok := o.agents[ag.Loacaladdr] // что бы отображались только новые агенты
	o.agents[ag.Loacaladdr] = &ag    //записываем нового агента если есть обновляем
	o.mu.Unlock()
	addrAgent := ag.Loacaladdr

	//если действительно новый агент
	if !ok {
		o.agentUpdate()
		Log("new Agent: ", ag.Loacaladdr)
	}

	var tsk *front.Task
	//если есть таски в очереди берём 1 таску
	o.mu.Lock()
	if len(o.Tasks) != 0 {
		tsk = o.Tasks[0]
		if len(o.Tasks) > 1 {
			o.Tasks = o.Tasks[1:]
		} else {
			o.Tasks = nil
		}
	}
	o.mu.Unlock()

	//если были таски в очереди
	if tsk != nil {
		o.mu.Lock()
		o.taskInWork[tsk.Id] = &TaskInWork{tsk, ag.Loacaladdr}
		o.mu.Unlock()

		front.Send(tsk, "http://localhost"+addrAgent+"/newTask")
		Log("Task Sending to agent")
	}
}

func (o *Orkestrator) heartBit(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		agent.PrintEr(err)
		return
	}
	//получили запрос от агента и распасим его данные
	var ag agent.Agent
	if err := json.Unmarshal(data, &ag); err != nil {
		agent.PrintEr(err)
		return
	}
	o.mu.Lock()
	o.agents[ag.Loacaladdr] = &ag //записываем нового агента если есть обновляем
	o.mu.Unlock()
}

// получаем таску с ответом
func (o *Orkestrator) sendAnswerTask(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		agent.PrintEr(err)
		return
	}

	tsk := front.Task{}
	err = json.Unmarshal(data, &tsk)
	if err != nil {
		agent.PrintEr(err)
		return
	}

	id := tsk.Id
	o.mu.Lock()
	delete(o.taskInWork, id)
	o.mu.Unlock()

	Logg("получили ответ на таску", tsk)
	front.Send(&tsk, "http://"+addrFront+"/getAnswer")
}

// послыаем агенту структуру с обновлённым временем выполнения задач
func (o *Orkestrator) updateTime(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return
	}

	timeOper := front.Operation{}
	err = json.Unmarshal(data, &timeOper)
	if err != nil {
		return
	}

	o.mu.Lock()
	o.TimeLimit = timeOper.All
	o.mu.Unlock()

	//обновляем время на всех агентах
	for k := range o.agents {
		front.Send(timeOper, "http://localhost"+k+"/newTimeLimit")
	}

	//добавляем таски из листа в рабте в очередь
	o.mu.Lock()
	for k, v := range o.taskInWork {
		o.Tasks = append(o.Tasks, v.task)
		delete(o.taskInWork, k)
	}
	o.mu.Unlock()
}
*/

// следим за тасками которые выполняются чтоб не превысили время
// если  агент отволился то все его задачи сново в очередь отсылаем
func (o *Orkestrator) MainOrkestrator() {
	go func() {
		for {
			//проверяем таски вышло ли время выполнения
			time.Sleep(time.Second)
			o.mu.Lock()
			for _, v := range o.taskInWork {
				v.task.Time--
				if v.task.Time <= 0 {
					Logg("task delete ", v.task.Id)
					delete(o.taskInWork, v.task.Id)
					v.task.Time = o.TimeLimit
					o.Tasks = append(o.Tasks, v.task)
				}
			}
			o.mu.Unlock()

			//проверяем агенты
			//если вышло время выполнения удаляем их все таски и перемещаем в очередь
			o.mu.Lock()
			for _, ag := range o.agents {
				ag.Status--

				if ag.Status <= 0 {
					Log("agent delete", ag.Loacaladdr)
					delete(o.agents, ag.Loacaladdr)

					for _, t := range o.taskInWork {
						if t.agent == ag.Loacaladdr {
							delete(o.taskInWork, t.task.Id)
							o.Tasks = append(o.Tasks, t.task)
						}
					}
					o.agentUpdate()
				}
			}
			o.mu.Unlock()
		}
	}()
}

// front - server
// newTask полученная с фронта
func (o *Orkestrator) newTask(w http.ResponseWriter, r *http.Request) {
	body := r.Body
	data, err := io.ReadAll(body)
	if err != nil {
		agent.PrintEr(err)
		return
	}
	Log("New task:", string(data))

	tsk := front.Task{}
	err = json.Unmarshal(data, &tsk)
	if err != nil {
		agent.PrintEr(err)
		return
	}
	//записываем в очередь
	o.mu.Lock()
	o.Tasks = append(o.Tasks, &tsk)
	o.mu.Unlock()
}

// отсылаем обновлённый список агентов на фронт если что то изменилось
func (o *Orkestrator) agentUpdate() {
	lstAgent := []*front.Agents{}
	fmt.Println("lstAgents", lstAgent)
	agent := &front.Agents{}
	o.mu.Lock()
	for _, v := range o.agents {
		agent.Name = v.Loacaladdr
		agent.Status = v.Status
		lstAgent = append(lstAgent, agent)
	}
	o.mu.Unlock()

	front.Send(lstAgent, "http://"+addrFront+"/updateAgents")
}

// запрашиваем с фронта и посылаем ответ
// изменяем статус сервера и он сообщает об ошибах (есть или нет)
func (o *Orkestrator) statusSrv(w http.ResponseWriter, r *http.Request) {
	o.mu.Lock()
	if len(o.agents) > 0 {
		serverStatus = 1
	}
	o.mu.Unlock()
	data := strconv.Itoa(serverStatus)
	w.Write([]byte(data))
}
