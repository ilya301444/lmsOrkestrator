// front
// отображает страницы
/*
localhost:8000
*/

package front

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Task struct { // данные таски которые будут передаваться в html
	Id         int
	Subid      int    // для того что бы понять как потом собирать задачу
	Expression string //будет также уникальным идентификатором выражения
	ValidExp   bool   //валидоно или не валидно выражение
	Time       int    //в секундах
	Status     int    //0  в процессе 1  выполнено
	Result     int
}

// время выполнения операций по умолчаню 200 -All 50 -other
type Operation struct {
	Plus  int
	Minus int
	Mul   int
	Div   int
	All   int
}

// список агентов которые нужны для отображения
// содержит меньше данных чем agent.Agent тк они не нужны для отображения
type Agents struct {
	Name   string
	Status int
}

// массив текущих данных (пока без Sql)
// для отображения
type Data struct {
	cashe    map[string]*template.Template //сохраняем страницы что бы не читать с диска
	listTask []*Task                       //сохраняем данные что бы не обращаться к бд постоянно
	mapTask  map[string]*Task
	srvList  []*Agents
	timeOper Operation
}

var data Data
var adrSrv = "localhost:8010"

func init() {
	data.mapTask = make(map[string]*Task)

	data.timeOper.All = 200
	data.timeOper.Plus = 50
	data.timeOper.Minus = 50
	data.timeOper.Mul = 50
	data.timeOper.Div = 50
}

// StartFront точка старнта фронта
func StartFront(ctx context.Context) (func(context.Context) error, error) {
	serverMux := http.NewServeMux()
	serverMux.HandleFunc("/", data.main)
	serverMux.HandleFunc("/listtask", data.list)
	serverMux.HandleFunc("/newtask", data.new)
	serverMux.HandleFunc("/setting", data.setting)
	serverMux.HandleFunc("/servers", data.servers)
	serverMux.HandleFunc("/getAnswer", data.getAnswer)

	srv := &http.Server{Addr: ":8000", Handler: serverMux}
	go func() {
		err := srv.ListenAndServe()
		PrintEr(err)
	}()

	data.cashe = make(map[string]*template.Template)

	templ, err := template.ParseFiles("front/template/listtask.html")
	FatalEr(err)
	data.cashe["listtask"] = templ
	templ, err = template.ParseFiles("front/template/newtask.html")
	FatalEr(err)
	data.cashe["newtask"] = templ
	templ, err = template.ParseFiles("front/template/servers.html")
	FatalEr(err)
	data.cashe["servers"] = templ
	templ, err = template.ParseFiles("front/template/setting.html")
	FatalEr(err)
	data.cashe["setting"] = templ
	templ, err = template.ParseFiles("front/template/main.html")
	FatalEr(err)
	data.cashe["main"] = templ

	return srv.Shutdown, nil
}

// main функция которая выводит страницу со ссылками на все остальные
func (d *Data) main(w http.ResponseWriter, r *http.Request) {
	err := data.cashe["main"].Execute(w, nil)
	PrintEr(err)
}

// список задач
func (d *Data) list(w http.ResponseWriter, r *http.Request) {
	err := data.cashe["listtask"].Execute(w, d.listTask)
	PrintEr(err)
}

// страница ввода выражения
func (d *Data) new(w http.ResponseWriter, r *http.Request) {
	expression := r.FormValue("exp")
	if expression == "" {
		err := data.cashe["newtask"].Execute(w, nil)
		PrintEr(err)
		return
	}

	res := validExpression(expression)
	data.addData(expression, res)
	if res {
		w.WriteHeader(200)
	} else {
		w.WriteHeader(400)
	}

	err := data.cashe["newtask"].Execute(w, nil)
	PrintEr(err)
}

// addData добавляем выражения в массив данных
func (d *Data) addData(expression string, res bool) error {
	previosId := 0
	if data.listTask != nil {
		previosId = len(d.listTask) - 1
	}

	newData := &Task{
		Id: previosId + 1, Expression: expression,
		ValidExp: res, Time: 200,
		Status: 0,
	}

	if _, ok := data.mapTask[expression]; !ok {
		data.mapTask[expression] = newData
		//посылаем данные
		err := d.sendSrv(newData)
		for err != nil {
			time.Sleep(20 * time.Millisecond)
			err = d.sendSrv(newData)
		}
		data.listTask = append(d.listTask, newData)
		Log(newData)
	}
	return nil
}

// sendSrv посылает данные серверу
func (d *Data) sendSrv(data *Task) error {
	return Send(data, "http://"+adrSrv+"/")
}

// получаем ответ со значением таски
func (d *Data) getAnswer(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		PrintEr(err)
		return
	}

	tsk := Task{}
	err = json.Unmarshal(data, &tsk)
	if err != nil {
		PrintEr(err)
		return
	}

	id := tsk.Id
	d.listTask[id-1].Result = tsk.Result
	d.listTask[id-1].Status = 1
}

// Send посылает даннные через пост в виде json по адресу
func Send(a interface{}, urlAdr string) error {
	dataJsn, err := json.Marshal(a)
	if err != nil {
		return err
	}
	resp, err := http.Post(urlAdr, "application/json", bytes.NewBuffer(dataJsn))
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

// setting страница с настройками выполнения задач
func (d *Data) setting(w http.ResponseWriter, r *http.Request) {
	err := data.cashe["setting"].Execute(w, d.timeOper)
	PrintEr(err)

	plus := r.FormValue("plus")
	minus := r.FormValue("minus")
	mul := r.FormValue("mul")
	div := r.FormValue("div")
	all := r.FormValue("all")

	if plus != "" {
		tm, _ := strconv.Atoi(plus)
		data.timeOper.Plus = tm
	}
	if minus != "" {
		tm, _ := strconv.Atoi(minus)
		data.timeOper.Minus = tm
	}
	if mul != "" {
		tm, _ := strconv.Atoi(mul)
		data.timeOper.Mul = tm
	}
	if div != "" {
		tm, _ := strconv.Atoi(div)
		data.timeOper.Div = tm
	}
	if all != "" {
		tm, _ := strconv.Atoi(all)
		data.timeOper.All = tm
	}

	if plus != "" || minus != "" || mul != "" || div != "" || all != "" {
		fmt.Println(plus, minus, mul, div, all)
	}
}

// servers список серверов
func (d *Data) servers(w http.ResponseWriter, r *http.Request) {
	err := data.cashe["servers"].Execute(w, d.srvList)
	PrintEr(err)
}

// ------ Logg
func Log(data interface{}) {
	fmt.Println("Log front: ", data)
}

func PrintEr(err error) {
	if err != nil {
		fmt.Println("front err: ", err)
	}
}

func FatalEr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
