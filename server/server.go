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
	"bytes"
	"context"
	"encoding/json"
	"io"
	"last/agent"
	"net/http"
	"sync"
	"time"
)

type TaskInWork struct {
	task  string
	agent string // addr agent
}

type Orkestrator struct {
	// по адресу находим нужный агент
	agents     map[string]agent.Agent //список серверов(агентов ) готовых выполнить задачу
	tasks      []string               //очередь из задач
	taskInWork []TaskInWork
	status     int // ждём таску(или другое событие) 0 или берём её из очереди 1 (нужно для синфронизации когда очередь пусата)
	syncTask   chan struct{}
	mu         sync.Mutex
}

var orkestr Orkestrator

func init() {
	orkestr.agents = make(map[string]agent.Agent)
	orkestr.syncTask = make(chan struct{}) // когда приходит новая таска посылаем сигнал чтоб снять оркестратор с ожидания
}

func StartSrv(ctx context.Context) (func(context.Context) error, error) {
	serverMux := http.NewServeMux()
	serverMux.HandleFunc("/", orkestr.newTask)
	serverMux.HandleFunc("/newAgent", orkestr.newAgent)

	srv := &http.Server{Addr: ":8010", Handler: serverMux}
	go func() {
		err := srv.ListenAndServe()
		agent.PrintEr(err)
	}()
	return srv.Shutdown, nil
}

// newTask полученная с фронта
func (o *Orkestrator) newTask(w http.ResponseWriter, r *http.Request) {
	body := r.Body
	data, err := io.ReadAll(body)
	if err != nil {
		agent.PrintEr(err)
		return
	}
	Log("New task:", string(data))
	o.tasks = append(o.tasks, string(data))
}

func (o *Orkestrator) newAgent(w http.ResponseWriter, r *http.Request) {
	body := r.Body
	data, err := io.ReadAll(body)
	if err != nil {
		agent.PrintEr(err)
		return
	}

	var ag agent.Agent
	if err := json.Unmarshal(data, &ag); err != nil {
		agent.PrintEr(err)
		return
	}
	o.agents[ag.Loacaladdr] = ag
	Log("new Agent: ", ag.Loacaladdr)
}

// hertBit смотрим живой ли агент и обновляем инф о нём
func (o *Orkestrator) hertBit() {

}

func (o *Orkestrator) MainOrkestrator() {

	go func() {
		for {
			adrAgent := ""
			//идём поочереди и раздаём задачи агентам
			for {
				if len(o.tasks) != 0 {
					o.status = 1
					//выбираем агент с наибольшим количеством потоков
					max := 0
					for k, v := range o.agents {
						if v.NumTread > max {
							max = v.NumTread
							adrAgent = k
						}
					}
					o.sendTask(adrAgent, o.tasks[0])
					if len(o.tasks) > 1 {
						o.tasks = o.tasks[1:]
					} else {
						o.tasks = nil
					}
				} else {
					break
				}
			}
			o.status = 0
			<-o.syncTask
		}

		time.Sleep(20 * time.Millisecond) // так как время выполнения 1 задачи 1с то 20мс должно хватить для
	}()
}

// посылаем таску агенту
func (o *Orkestrator) sendTask(adrAgent string, task string) {
	dataJsn, err := json.Marshal()
	if err != nil {
		return err
	}

	resp, err := http.Post("http://"+serverAddr+"/newAgent", "application/json", bytes.NewBuffer(dataJsn))
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}
