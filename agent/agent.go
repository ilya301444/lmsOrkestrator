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
	"strings"
	"sync"
	"time"

	expp "github.com/overseven/go-math-expression-parser/parser"
	"google.golang.org/grpc"
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
	stopAgent  chan struct{}
}

var (
	serverAddr = "127.0.0.1:7010" //addres orkestratora
	//настройки интервала отправки запроса на сервер
	timeSleep = 1000 * time.Millisecond
	agent     Agent
)

func init() {
	agent.numTread = runtime.NumCPU() - 1
	agent.limitTread = runtime.NumCPU() - 1

	// agent.numTread = 5
	// agent.limitTread = 5
	agent.stop = make(chan struct{}, 3)
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
		log.Println("log agent:", err)
	}
}

func FatalEr(err error) {
	if err != nil {
		log.Fatal("agent fatall err: ", err)
	}
}

// Log для агента
func Log(s string) {
	log.Println("agent log:", s)
}

func Logg(s ...interface{}) {
	log.Println("agent log:", s)
}

// StartAgent запускает основу с которой будет общаться
// func StartAgent(ctx context.Context, port string) (func(context.Context) error, error) {

// 	serveMux := http.NewServeMux()
// 	serveMux.HandleFunc("/newTimeLimit", agent.newTimeLimit)
// 	serveMux.HandleFunc("/newTask", agent.newTask)

// 	agent.Loacaladdr = ":" + port
// 	agent.stop = make(chan struct{}, 32)
// 	agent.servrConn()

// 	srv := &http.Server{Addr: agent.Loacaladdr, Handler: serveMux}
// 	go func() {
// 		_ = srv.ListenAndServe()
// 		agent.stopAgent <- struct{}{}
// 	}()

// 	return srv.Shutdown, nil
// }

func StartAgent(ctx context.Context, port string) (*grpc.Server, error) {
	var srv *grpc.Server
	agent.Loacaladdr = ":" + port
	srv = SrvGrpcStart(agent.Loacaladdr)
	return srv, nil
}

// типа оркестратор в агенте (запускает потоки, как только пришла задача от сервера от сервера)
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

			err := front.Send(data, "http://"+serverAddr+"/heartBit")
			PrintEr(err)

			select {
			case <-a.stopAgent:
				Log("agent 2 close")
				return
			default:
			}
		}
	}()

	//горутина по получению задач постоянно отправляем запрос если есть потоки и
	//нам приходят задачи через пост в другую функцию
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
				_, err = http.Post("http://"+serverAddr+"/getTask", "application/json", bytes.NewBuffer(dataJsn))
				if err != nil {
					PrintEr(err)
				}
			}

			select {
			case <-a.stopAgent:
				Log("agent 1 close")
				return
			default:
			}
		}
	}()
}

// выдаёт сколько времени надо выполнять задачу в зависимости от знака в выражении
func (a *Agent) getTimeLimit(exp string) int {
	plus := strings.Count(exp, "+")
	minus := strings.Count(exp, "-")
	mul := strings.Count(exp, "*")
	div := strings.Count(exp, "/")
	//если не только 1 знак в выражении
	if plus+minus+mul+div > 0 {
		return plus*a.operLimit.Plus + minus*a.operLimit.Minus + mul*a.operLimit.Mul + div*a.operLimit.Div
	}
	return a.operLimit.All
}

// останавливаем все задачи
func (a *Agent) reboot() {
	Log("rebooting")
	a.mu.Lock()
	count := a.limitTread - a.numTread
	a.mu.Unlock()
	for i := 0; i < count; i++ {
		Log("send stop")
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

	Log("update time limit")
	a.mu.Lock()
	a.operLimit = timeOper
	a.mu.Unlock()
	a.reboot() // перезагружаем и удаляем все таски
}

// получаем задачу и запускаем горутину для её выполнения
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
			Log("stoped task")
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

// ExecuteTask решает пример ответ записывает в структуру
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
