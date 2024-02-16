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
	"log"
	"net/http"
	"runtime"
	"time"

	expp "github.com/overseven/go-math-expression-parser/parser"
)

type Task struct {
	resultChan chan int
	ValueTask  string
}

type Agent struct {
	Id         int
	Loacaladdr string // адрес агента
	NumTread   int
	Status     int             // < 0  - мёртв 1 - жив если 100 раз мёртв(100с) - то исключаем из агентов
	listTask   map[string]Task //expression Task
}

var (
	serverAddr = "127.0.0.1:8010" //addres orkestratora
)

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

func Log(s string) {
	fmt.Println("log:", s)
}

var agent Agent

// StartAgent запускает основу с которой будет общаться
func StartAgent(ctx context.Context, port string) (func(context.Context) error, error) {
	serveMux := http.NewServeMux()
	serveMux.HandleFunc("/", agent.newTask)
	serveMux.HandleFunc("/heartBit", agent.heartBit)

	agent.Loacaladdr = ":" + port
	err := agent.servrConn()
	for err != nil {
		time.Sleep(200 * time.Millisecond)
		err = agent.servrConn()
		fmt.Println("errror", err)

	}

	srv := &http.Server{Addr: agent.Loacaladdr, Handler: serveMux}
	go func() {
		err := srv.ListenAndServe()
		PrintEr(err)
	}()

	return srv.Shutdown, nil
}

// servrConn подключаемся к серверу, заявляем о новом агенте
func (a *Agent) servrConn() error {
	a.NumTread = runtime.NumCPU() - 1
	a.Status = 1
	dataJsn, err := json.Marshal(a)
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

// NewTask получение новой задачи
func (a *Agent) newTask(w http.ResponseWriter, r *http.Request) {

}

// heartBit проверка состояния со стороны оркестратора
func (a *Agent) heartBit(w http.ResponseWriter, r *http.Request) {
	dataJsn, err := json.Marshal(a)
	if err != nil {
		PrintEr(err)
		return
	}

	w.Write(dataJsn)
}

func (a *Agent) ExecuteTask(str string) int {
	parser := expp.NewParser()
	// parsing
	_, err := parser.Parse(str)
	if err != nil {
		fmt.Println("Error: ", err)
		return 0
	}
	//fmt.Println("\nParsed execution tree:", exp)
	// output: 'Parsed execution tree: ( * 10 ( bar ( 60,6,0.6 ) ) )'
	// execution of the expression
	result, err := parser.Evaluate(map[string]float64{})
	if err != nil {
		fmt.Println("Error: ", err)
	}
	return int(result)
}

/*
func getPort() (int, error) {
	resp, err := http.Get(serverAddr)
	if err != nil {
		PrintEr(err)
		return 0, nil
	}

	data, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		PrintEr(err)
		return 0, err
	}

	num, err := strconv.Atoi(string(data))
	if err != nil {
		PrintEr(err)
		return 0, err
	}
	return num, nil
}
*/
