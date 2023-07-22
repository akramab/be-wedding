package user

import (
	apierror "be-wedding/internal/rest/error"
	"be-wedding/internal/rest/response"
	"be-wedding/internal/store"
	"encoding/csv"
	"encoding/json"
	"fmt"
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
	QRImage  string
}

func (handler *userHandler) SynchronizeUser(w http.ResponseWriter, r *http.Request) {
	// open file
	ctx := r.Context()
	req := SynchronizeUserRequest{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, apierror.BadRequestError(err.Error()))
		return
	}

	f, err := os.Open("./static/" + req.FileName)
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
	for idx, line := range data {
		if idx != 0 {
			synchroUserData := SyncronizeUserData{}
			for columnNumber, field := range line {
				if columnNumber == 0 {
					synchroUserData.Name = field
				}

				if columnNumber == 1 {
					synchroUserData.QRImage = fmt.Sprintf("%s.png", field)
				}

				if columnNumber == 2 {
					synchroUserData.WaNumber = field
				}
			}
			synchroUserDataList = append(synchroUserDataList, synchroUserData)
		}
	}

	for _, synchroUser := range synchroUserDataList {
		newUserData := &store.UserData{
			InvitationID:   "synchronize-data",
			InvitationType: "GROUP",
			WhatsAppNumber: synchroUser.WaNumber,
			QRImageName:    synchroUser.QRImage,
		}

		if err := handler.userStore.Insert(ctx, newUserData); err != nil {
			log.Println("error insert new user data: %w", err)
			response.Error(w, apierror.BadRequestError(fmt.Sprintf("phone number: %s already exists!", newUserData.WhatsAppNumber)))
			return
		}

		userData := &store.UserData{
			ID:   newUserData.ID,
			Name: synchroUser.Name,
		}

		if err := handler.userStore.Update(ctx, userData); err != nil {
			log.Println("error update new user data: %w", err)
			response.Error(w, apierror.InternalServerError())
			return
		}

		handler.userStore.InsertUserRSVP(ctx, &store.UserRSVPData{
			UserID:      newUserData.ID,
			PeopleCount: 1,
		})

		// CREATE QR
		// qrImageInitial := fmt.Sprintf("qr-%s.png", uuid.NewString())
		// initialFilePath := fmt.Sprintf("./static/qr-codes/%s", qrImageInitial)
		// finalFilePath := fmt.Sprintf("./static/qr-codes/%s", qrImageName)
		// err = qrcode.WriteColorFile(newUserData.ID, qrcode.Medium, 256, color.White, color.RGBA{110, 81, 59, 255}, initialFilePath)
		// if err != nil {
		// 	log.Println(err.Error())
		// 	response.Error(w, apierror.BadRequestError(err.Error()))
		// 	return
		// }

		// const S = 256
		// im, err := gg.LoadImage(initialFilePath)
		// if err != nil {
		// 	log.Fatal(err)
		// }

		// dc := gg.NewContext(S, S+20)
		// dc.SetRGB(1, 1, 1)
		// dc.Clear()
		// dc.SetRGB(0, 0, 0)
		// if err := dc.LoadFontFace("./static/fonts/Alice-Regular.ttf", 12); err != nil {
		// 	panic(err)
		// }

		// dc.DrawImage(im, 0, 20)
		// dc.DrawStringAnchored("Tiket reservasi pernikahan Afra & Akram", S/2, 10, 0.5, 0.5)
		// dc.DrawStringAnchored(fmt.Sprintf("untuk %s", newUserData.Name), S/2, 20, 0.5, 0.5)

		// dc.Clip()
		// dc.SavePNG(finalFilePath)
	}

	log.Println("DEBUG")
	log.Println(synchroUserDataList)
}
