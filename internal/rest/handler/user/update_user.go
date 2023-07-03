package user

import (
	"encoding/json"
	"fmt"
	"image/color"
	"log"
	"net/http"

	apierror "be-wedding/internal/rest/error"
	"be-wedding/internal/store"

	"be-wedding/internal/rest/response"

	"github.com/fogleman/gg"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	qrcode "github.com/skip2/go-qrcode"
)

type UpdateUserRequest struct {
	Name string `json:"name"`
}

type UpdateUserResponse struct {
	Message string `json:"message"`
	UserID  string `json:"user_id"`
	Name    string `json:"name"`
}

func (handler *userHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := chi.URLParam(r, "id")

	req := UpdateUserRequest{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, apierror.BadRequestError(err.Error()))
		return
	}

	userData := &store.UserData{
		ID:   userID,
		Name: req.Name,
	}

	if err := handler.userStore.Update(ctx, userData); err != nil {
		log.Println("error update new user data: %w", err)
		response.Error(w, apierror.InternalServerError())
		return
	}

	invitationCompleteData, err := handler.invitationStore.FindOneCompleteDataByUserID(ctx, userID)
	if err != nil {
		log.Println(err)
		response.Error(w, apierror.NotFoundError("User id not found"))
		return
	}

	qrImageInitial := fmt.Sprintf("qr-%s.png", uuid.NewString())
	qrImageName := invitationCompleteData.User.QRImage
	initialFilePath := fmt.Sprintf("./static/qr-codes/%s", qrImageInitial)
	finalFilePath := fmt.Sprintf("./static/qr-codes/%s", qrImageName)
	err = qrcode.WriteColorFile(invitationCompleteData.User.ID, qrcode.Medium, 256, color.White, color.RGBA{110, 81, 59, 255}, initialFilePath)
	if err != nil {
		log.Println(err.Error())
		response.Error(w, apierror.BadRequestError(err.Error()))
		return
	}

	const S = 256
	im, err := gg.LoadImage(initialFilePath)
	if err != nil {
		log.Fatal(err)
	}

	dc := gg.NewContext(S, S+20)
	dc.SetRGB(1, 1, 1)
	dc.Clear()
	dc.SetRGB(0, 0, 0)
	if err := dc.LoadFontFace("./static/fonts/Alice-Regular.ttf", 12); err != nil {
		panic(err)
	}

	dc.DrawImage(im, 0, 20)
	dc.DrawStringAnchored("Tiket reservasi pernikahan Afra & Akram", S/2, 10, 0.5, 0.5)
	dc.DrawStringAnchored(fmt.Sprintf("untuk %s", invitationCompleteData.User.Name), S/2, 20, 0.5, 0.5)

	dc.Clip()
	dc.SavePNG(finalFilePath)

	resp := UpdateUserResponse{
		Message: "success",
		UserID:  userData.ID,
		Name:    userData.Name,
	}

	response.Respond(w, http.StatusCreated, resp)
}
