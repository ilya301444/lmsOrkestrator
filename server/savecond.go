// сохраняем даннные в БД используем функции  из last/front
package server

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"last/front"
)

var db front.DB

func init() {
	var err error
	db.DB, err = sql.Open("sqlite3", "./server/store.db")
	if err != nil {
		log.Fatal(err)
	}

	err = db.CreateTabels()
	if err != nil {
		log.Fatal("create tabel ", err)
	}
}

// работа с оркестратором происходит в файле server.go
// сохраняем состояние
func (o *Orkestrator) saveCondact() {
	//добавляем таски из листа в рабте в очередь
	o.mu.Lock()
	for k, v := range o.taskInWork {
		o.Tasks = append(o.Tasks, v.task)
		delete(o.taskInWork, k)
	}
	o.mu.Unlock()

	data, err := json.Marshal(o)
	if err != nil {
		Logg("err marshaling : ", err)
		return
	}

	err = db.SaveData(data)
	if err != nil {
		Logg(err)
	}

	Log("save condact")
	for k := range o.agents {
		http.Get("http://localhost" + k + "/reboot")
	}
}

// восстанавливаем состояние
func restoreCondact() {
	data, err := db.LoadData()
	if err != nil {
		Logg("load data from tabel: ", err)
		return
	}

	err = json.Unmarshal([]byte(data), &orkestr)
	if err != nil {
		Logg(err)
		return
	}
}
