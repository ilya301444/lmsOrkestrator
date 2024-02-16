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
	"html/template"
	"last/agent"
	"net/http"
	"time"
)

type Task struct { // данные таски которые будут передаваться в html
	Id            int
	ExpControlSum string
	Expression    string
	ValidExp      bool      //валидоно или не валидно выражение
	Time          int       //в секундах
	TimeEnd       time.Time //время окончания выполнение задачи timeNow + Time*time.Second
	Status        int       //0 - в процессе 1 - выполнено
	Result        int
}

type Server struct {
	Name   string
	Status int
}

// массив текущих данных (пока без Sql)
type Data struct {
	cashe    map[string]*template.Template //сохраняем страницы что бы не читать с диска
	listTask []*Task                       //сохраняем данные что бы не обращаться к бд постоянно
	srvList  []*Server
}

var data Data
var adrSrv = "localhost:8010"

// StartFront точка старнта фронта
func StartFront(ctx context.Context) (func(context.Context) error, error) {
	serverMux := http.NewServeMux()
	serverMux.HandleFunc("/", data.main)
	serverMux.HandleFunc("/calculator", data.calculator)
	serverMux.HandleFunc("/setting", data.setting)
	serverMux.HandleFunc("/servers", data.servers)
	//получаем данные со страниц
	serverMux.HandleFunc("/calculatorAdd", data.calculatorAdd)
	serverMux.HandleFunc("/settingChange", data.settingChange)

	srv := &http.Server{Addr: ":8000", Handler: serverMux}
	go func() {
		err := srv.ListenAndServe()
		agent.PrintEr(err)
	}()

	data.cashe = make(map[string]*template.Template)

	templ, err := template.ParseFiles("front/template/calculator.html")
	agent.FatalEr(err)
	data.cashe["calculator"] = templ
	templ, err = template.ParseFiles("front/template/servers.html")
	agent.FatalEr(err)
	data.cashe["setting"] = templ
	templ, err = template.ParseFiles("front/template/setting.html")
	agent.FatalEr(err)
	data.cashe["servers"] = templ
	templ, err = template.ParseFiles("front/template/main.html")
	agent.FatalEr(err)
	data.cashe["main"] = templ

	return srv.Shutdown, nil
}

// main функция которая выводит страницу со ссылками на все остальные
func (d *Data) main(w http.ResponseWriter, r *http.Request) {
	err := data.cashe["main"].Execute(w, nil)
	agent.PrintEr(err)
}

// calculator страница ввода выражения
func (d *Data) calculator(w http.ResponseWriter, r *http.Request) {
	err := data.cashe["calculator"].Execute(w, d.listTask)
	agent.PrintEr(err)
}

// calculatorAdd сюда перенаправляемся для получения данных со страниы ввода
func (d *Data) calculatorAdd(w http.ResponseWriter, r *http.Request) {
	expression := r.FormValue("exp")
	res := validExpression(expression)

	data.addData(expression, res)
	if res {
		http.Redirect(w, r, "/calculator", http.StatusOK)
	} else {
		http.Redirect(w, r, "/calculator", http.StatusBadRequest)
	}
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
		TimeEnd: time.Now(), Status: 0,
	}
	//посылаем данные
	err := d.sendSrv(newData)
	for err != nil {
		time.Sleep(20 * time.Millisecond)
		err = d.sendSrv(newData)
	}

	data.listTask = append(d.listTask, newData)
	return nil
}

// sendSrv посылает данные серверу
func (d *Data) sendSrv(data *Task) error {
	dataJsn, err := json.Marshal(data)
	if err != nil {
		return err
	}

	resp, err := http.Post("http://"+adrSrv+"/", "application/json", bytes.NewBuffer(dataJsn))
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

// setting страница с настройками выполнения задач
func (d *Data) setting(w http.ResponseWriter, r *http.Request) {
	err := data.cashe["setting"].Execute(w, d.listTask)
	agent.PrintEr(err)
}

// settingChange перенаправляемся сюда для ввода изменённых настроек окончания выполнеия задачи
func (d *Data) settingChange(w http.ResponseWriter, r *http.Request) {

}

// servers список серверов
func (d *Data) servers(w http.ResponseWriter, r *http.Request) {
	err := data.cashe["servers"].Execute(w, d.srvList)
	agent.PrintEr(err)
}
