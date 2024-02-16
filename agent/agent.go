// agent
// Вычислитель, который может получить от оркестратора задачу, выполнить его и
// вернуть серверу результат. Далее будем называть его агентом.
/*
Подразумевается что все агенты будут храниться на одном Ip адресе
новый порт должен вводить пользователь  при запуске нового агента

колличество доступных потоков -1
на 1м будет крутится наш сервер для получения и хёрт бит


localhost:8020
новых серверов начинаются 8020


общаеься с оркестратором на localhost:8010

*/

package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"last/front"
	"log"
	"net/http"
	"runtime"
	"sync"
	"time"

	expp "github.com/overseven/go-math-expression-parser/parser"
)

type Agent struct {
	Id         int
	Loacaladdr string // адрес агента
	Status     int    // < 0  - мёртв 100 - жив если 100 раз мёртв(100с) - то исключаем из агентов(нужно для сервера)

	operLimit  front.Operation //лимит времени выполнения подзадачи
	mu         sync.RWMutex
	numTread   int
	limitTread int
	stop       chan struct{}
}

var (
	serverAddr = "127.0.0.1:8010" //addres orkestratora
	//настройки интервала отправки запроса на сервер
	timeSleep = 1000 * time.Millisecond
	agent     Agent
)

func init() {
	agent.numTread = runtime.NumCPU() - 1
	agent.limitTread = runtime.NumCPU() - 1
	agent.Status = 100 //после 100 раза (100с) без ответа будет значить что сервис умер

	agent.operLimit.All = 200
	agent.operLimit.Plus = 50
	agent.operLimit.Minus = 50
	agent.operLimit.Mul = 50
	agent.operLimit.Div = 50
}

// функции отладки для агента
func PrintEr(err error) {
	if err != nil {
		fmt.Println("log:", err)
	}
}

func FatalEr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// Log для агента
func Log(s string) {
	fmt.Println("agent log:", s)
}

func Logg(s ...any) {
	fmt.Println("agent log:", s)
}

// StartAgent запускает основу с которой будет общаться
func StartAgent(ctx context.Context, port string) (func(context.Context) error, error) {
	serveMux := http.NewServeMux()
	//удаляем все вычисления
	serveMux.HandleFunc("/reboot", agent.reboot)
	serveMux.HandleFunc("/newTimeLimit", agent.newTimeLimit)
	serveMux.HandleFunc("/newTask", agent.newTask)

	agent.Loacaladdr = ":" + port
	agent.stop = make(chan struct{}, 32)
	agent.servrConn()

	srv := &http.Server{Addr: agent.Loacaladdr, Handler: serveMux}
	go func() {
		err := srv.ListenAndServe()
		agent.mu.Lock()
		agent.Status = 0
		agent.mu.Unlock()

		PrintEr(err)
	}()

	return srv.Shutdown, nil
}

// оркестратор в агенте (запускает потоки, в зависимости от полученных задачь от сервера)
// servrConn подключаемся к серверу
func (a *Agent) servrConn() {
	// горутина которая говорит что данный агент всё ещё жив
	//heartBit
	go func() {
		for {
			time.Sleep(timeSleep)
			a.mu.RLock()
			data := a
			a.mu.RUnlock()

			err := front.Send(data, "http://"+serverAddr+"/getTask")
			PrintEr(err)
		}
	}()

	//горутина по получению задач
	go func() {
		for {
			time.Sleep(timeSleep)
			a.mu.Lock()
			num := a.numTread
			a.mu.Unlock()
			// запрашиваем таску если не заняты все потоки
			if num > 0 {
				a.mu.RLock()
				dataJsn, err := json.Marshal(a)
				a.mu.RUnlock()

				if err != nil {
					PrintEr(err)
				}

				//запросили задачу
				resp, err := http.Post("http://"+serverAddr+"/getTask", "application/json", bytes.NewBuffer(dataJsn))
				if err != nil {
					PrintEr(err)
				}
				resp.Body.Close()
			}
		}
	}()
}

// выдаёт сколько времени надо выполнять задачу в зависимости от знака в выражении
func (a *Agent) getTimeLimit(exp string) int {
	if len(exp) == 3 {
		ch := exp[1]
		switch ch {
		case '+':
			return a.operLimit.Plus
		case '-':
			return a.operLimit.Minus
		case '*':
			return a.operLimit.Mul
		case '/':
			return a.operLimit.Div
		}
	}
	return a.operLimit.All
}

// останавливаем все задачи
func (a *Agent) reboot(w http.ResponseWriter, r *http.Request) {
	for i := 0; i < a.limitTread-a.numTread; i++ {
		a.stop <- struct{}{}
	}
}

// обнавляем лимиты времени выполнения задачи
func (a *Agent) newTimeLimit(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		Logg(err)
		return
	}

	timeOper := front.Operation{}
	err = json.Unmarshal(data, &timeOper)
	if err != nil {
		Logg(err)
		return
	}
	a.operLimit = timeOper
}

// обнавляем лимиты времени выполнения задачи
func (a *Agent) newTask(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		PrintEr(err)
		return
	}
	tsk := &front.Task{}
	err = json.Unmarshal(data, tsk)
	if err != nil {
		PrintEr(err)
		return
	}

	//запускаем задачу на выполнение в отдельной горутине
	go func() {
		a.ExecuteTask(tsk)

		// время после которого будет выполнена задача (считаем это время незначительным по сравнению с секундой)
		timeStop := time.Duration(a.getTimeLimit(tsk.Expression)) * time.Second
		select {
		//если пришёл сигнал на перезагрузку (изменилось время выполнения операций +-*/)
		case <-a.stop:
			return
		case <-time.After(timeStop):
			err := a.sendTask(tsk)
			PrintEr(err)
		}

		a.mu.Lock()
		a.numTread++
		a.mu.Unlock()

		Logg("task is continue! ", tsk)
	}()
}

func (a *Agent) ExecuteTask(tsk *front.Task) {
	a.mu.Lock()
	a.numTread--
	a.mu.Unlock()
	str := tsk.Expression

	parser := expp.NewParser()
	_, err := parser.Parse(str)
	if err != nil {
		PrintEr(err)
	}
	result, err := parser.Evaluate(map[string]float64{})
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}

	Logg(result, str)
	tsk.Status = 1
	tsk.Result = int(result)
}

// agent - server
// отсылаем таску серверу с резуальтатом
func (a *Agent) sendTask(tsk *front.Task) error {
	return front.Send(tsk, "http://"+serverAddr+"/sendTask")
}
