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
	"sync"
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
	mu    sync.Mutex
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
	mu       sync.Mutex
}

var data Data
var adrSrv = "localhost:8010"
var srvStatus = 0 // 0 - нет ответа от сервера 1 - норм 2 - нет связи ни с одним агентов

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
	serverMux.HandleFunc("/updateAgents", data.updateAgents)
	serverMux.HandleFunc("/getAgentList", data.getAgentList)

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

	checkSrvStatus()

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

	if srvStatus == 1 {
		if res {
			w.WriteHeader(200)
		} else {
			w.WriteHeader(400)
		}
	} else {
		w.WriteHeader(500)
	}

	err := data.cashe["newtask"].Execute(w, nil)
	PrintEr(err)
}

// addData добавляем выражения в массив данных
func (d *Data) addData(expression string, res bool) error {
	///previosId := 0
	previosId := len(d.listTask)

	newData := &Task{
		Id: previosId, Expression: expression,
		ValidExp: res, Time: d.timeOper.All,
		Status: 0,
	}

	if _, ok := data.mapTask[expression]; !ok {
		data.mapTask[expression] = newData

		if res {
			//посылаем данные
			err := d.sendSrv(newData)
			for err != nil {
				time.Sleep(20 * time.Millisecond)
				err = d.sendSrv(newData)
			}
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
	d.listTask[id].Result = tsk.Result
	d.listTask[id].Status = 1
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

	switch {
	case plus != "":
		tm, err := strconv.Atoi(plus)
		if err != nil || tm == d.timeOper.Plus {
			return
		}
		d.timeOper.Plus = tm
	case minus != "":
		tm, err := strconv.Atoi(minus)
		if err != nil || tm == d.timeOper.Minus {
			return
		}
		d.timeOper.Minus = tm
	case mul != "":
		tm, err := strconv.Atoi(mul)
		if err != nil || tm == d.timeOper.Mul {
			return
		}
		d.timeOper.Mul = tm
	case div != "":
		tm, err := strconv.Atoi(div)
		if err != nil || tm == d.timeOper.Div {
			return
		}
		d.timeOper.Div = tm
	case all != "":
		tm, err := strconv.Atoi(all)
		if err != nil || tm == d.timeOper.All {
			return
		}
		d.timeOper.All = tm
	}

	if plus != "" || minus != "" || mul != "" || div != "" || all != "" {
		Send(d.timeOper, "http://"+adrSrv+"/updateTime")
		fmt.Println(plus, minus, mul, div, all)
	}
}

// servers список серверов
func (d *Data) servers(w http.ResponseWriter, r *http.Request) {
	err := data.cashe["servers"].Execute(w, d.srvList)
	PrintEr(err)
}

// обновляем список серверов (Агентов)
func (d *Data) updateAgents(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return
	}

	lst := []*Agents{}
	err = json.Unmarshal(data, &lst)
	if err != nil {
		return
	}
	d.srvList = lst
}

// Выдаём список агентов по запросу
func (d *Data) getAgentList(w http.ResponseWriter, r *http.Request) {
	data, err := json.Marshal(d.srvList)
	if err != nil {
		return
	}
	w.Write(data)
}

// функция которая будет запрашивать сервер о его состоянии и менять глобальную переменную
func checkSrvStatus() {
	go func() {
		for {
			time.Sleep(500 * time.Millisecond)
			resp, err := http.Get("http://" + adrSrv + "/statusSrv")
			if err != nil {
				continue
			}
			defer resp.Body.Close()

			data, err := io.ReadAll(resp.Body)
			if err != nil {
				continue
			}
			status, err := strconv.Atoi(string(data))
			if err != nil {
				continue
			}
			srvStatus = status
		}
	}()
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
