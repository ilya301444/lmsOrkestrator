//сохраняем и восстанавливаем состояние данных
package front

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

var nameFileBackup = "./front/frontSave.txt"

//сохраняем состояние
func (d *Data) saveCondact() {
	data, err := json.Marshal(d)
	if err != nil {
		PrintEr(err)
		return
	}
	Log("save condact")

	os.Remove(nameFileBackup)
	f, err := os.OpenFile(nameFileBackup, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		PrintEr(err)
		return
	}

	f.WriteString(string(data))
	f.Close()
}

//восстанавливаем состояние
func restoreCondact() {
	f, err := os.Open(nameFileBackup)
	if err != nil {
		PrintEr(err)
		return
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		PrintEr(err)
		return
	}

	err = json.Unmarshal(data, &data)

	if err != nil {
		PrintEr(err)
		fmt.Println(string(data))
		return
	}
}
