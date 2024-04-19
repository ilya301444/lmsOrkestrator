// сохраняем и восстанавливаем состояние данных
package front

import (
	"database/sql"
	"encoding/json"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	DB *sql.DB
}

var db DB

type DataDB struct {
	id   int64
	data string
}

func init() {
	var err error
	db.DB, err = sql.Open("sqlite3", "./front/store.db")
	if err != nil {
		log.Fatal(err)
	}

	err = db.CreateTabels()
	if err != nil {
		log.Fatal("create tabel ", err)
	}
}

func (d *DB) CreateTabels() error {
	const (
		agetTb = `create table if not exists agent (
			id integer primary key,
			data text
		);`
		insertNullData = `INSERT or IGNORE into agent values(1, "")`
	)
	_, err := d.DB.Exec(agetTb)
	if err != nil {
		return err
	}

	_, err = d.DB.Exec(insertNullData)
	if err != nil {
		return err
	}

	return nil
}

func (d *DB) LoadData() (string, error) {
	var query string = "SELECT data from agent where id=1"
	row := d.DB.QueryRow(query)
	data := DataDB{}
	err := row.Scan(&data.data)
	if err != nil {
		return "", err
	}
	return data.data, nil
}

func (d *DB) SaveData(dat []byte) error {
	const query = `update agent set data=$1 where id=1`
	_, err := d.DB.Exec(query, string(dat))
	if err != nil {
		return err
	}

	return nil
}

// сохраняем состояние
func (d *Data) saveCondact() {
	d.mu.Lock()
	data, err := json.Marshal(d)
	d.mu.Unlock()
	if err != nil {
		PrintEr(err)
		return
	}
	Log("save condact")

	err = db.SaveData(data)
	if err != nil {
		PrintEr(err)
	}
}

// восстанавливаем состояние
func (d *Data) restoreCondact() {
	dat, err := db.LoadData()
	tmpDat := Data{}
	err = json.Unmarshal([]byte(dat), &tmpDat)
	if err != nil {
		PrintEr(err)
		return
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	d.ListTask = tmpDat.ListTask
	d.MapTask = tmpDat.MapTask
	d.TimeOper = tmpDat.TimeOper
}
