// server
// оркестратор
// Сервер, который принимает арифметическое выражение,
// переводит его в набор последовательных задач и обеспечивает
// порядок их выполнения. Далее будем называть его оркестратором.

/*
localhost:8010
*/
package server

import (
	"context"
	"encoding/json"
	"io"
	"last/agent"
	"last/front"
	"net/http"
	"sync"
	"time"
)

type TaskInWork struct {
	task  *front.Task
	agent string // addr agent
}

type Orkestrator struct {
	// по адресу находим нужный агент
	agents     map[string]*agent.Agent //список серверов(агентов ) готовых выполнить задачу
	tasks      []*front.Task           //очередь из задач
	taskInWork []*TaskInWork

	mu sync.Mutex
}

var (
	orkestr   Orkestrator
	addrFront = "localhost:8000"
)

func init() {
	orkestr.agents = make(map[string]*agent.Agent)
}

func StartSrv(ctx context.Context) (func(context.Context) error, error) {
	serverMux := http.NewServeMux()
	serverMux.HandleFunc("/", orkestr.newTask)
	//агент может получить таску
	serverMux.HandleFunc("/getTask", orkestr.getTask)
	serverMux.HandleFunc("/sendTask", orkestr.sendAnswerTask)

	srv := &http.Server{Addr: ":8010", Handler: serverMux}
	go func() {
		err := srv.ListenAndServe()
		agent.PrintEr(err)
	}()
	return srv.Shutdown, nil
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
	o.tasks = append(o.tasks, &tsk)
	o.mu.Unlock()
}

// server - agent
// отдаём задачу агенту
// агент спаминт нас запросами мы ему отдаём таски
// это функция по получению Хёртбита и добавления новых агентов и отдачи таски
// новому агенту
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

	//если действительно новый агент
	if !ok {
		Log("new Agent: ", ag.Loacaladdr)
	}

	var tsk *front.Task
	//если есть таски в очереди берём 1 таску
	o.mu.Lock()
	if len(o.tasks) != 0 {
		tsk = o.tasks[0]
		if len(o.tasks) > 1 {
			o.tasks = o.tasks[1:]
		} else {
			o.tasks = nil
		}
	}
	o.mu.Unlock()

	//если были таски в очереди
	if tsk != nil {
		o.mu.Lock()
		o.taskInWork = append(o.taskInWork, &TaskInWork{tsk, ag.Loacaladdr})
		o.mu.Unlock()

		data, err = json.Marshal(tsk)
		if err != nil {
			agent.PrintEr(err)
			return
		}
		Log("Task Sending")
		n, err := w.Write(data)

		Logg(n, err)

		time.Sleep(time.Microsecond)
	}
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

	Logg("получили ответ на таску", tsk)
	front.Send(&tsk, "http://"+addrFront+"/getAnswer")
}

func (o *Orkestrator) MainOrkestrator() {

}
