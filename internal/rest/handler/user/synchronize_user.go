package user

import (
	"encoding/csv"
	"log"
	"net/http"
	"os"
)

type SynchronizeUserRequest struct {
	FileName string `json:"filename"`
}

type SyncronizeUserData struct {
	Name     string
	WaNumber string
}

func (handler *userHandler) SynchronizeUser(w http.ResponseWriter, r *http.Request) {
	// open file
	req := SynchronizeUserRequest{}
	f, err := os.Open(req.FileName)
	if err != nil {
		log.Fatal(err)
	}

	// remember to close the file at the end of the program
	defer f.Close()

	// read csv values using csv.Reader
	csvReader := csv.NewReader(f)
	data, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal(err)
	}

	synchroUserDataList := []SyncronizeUserData{}
	for _, line := range data {
		synchroUserData := SyncronizeUserData{}
		for columnNumber, field := range line {
			if columnNumber == 0 {
				synchroUserData.Name = field
			}

			if columnNumber == 2 {
				synchroUserData.WaNumber = field
			}
		}
		synchroUserDataList = append(synchroUserDataList, synchroUserData)
	}
}
