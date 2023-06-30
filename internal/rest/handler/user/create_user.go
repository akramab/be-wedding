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

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"google.golang.org/protobuf/proto"

	qrcode "github.com/skip2/go-qrcode"
)

type CreateNewUserRequest struct {
	WhatsAppNumber string `json:"wa_number"`
}

type CreateUserResponse struct {
	Message string `json:"message"`
	UserID  string `json:"user_id"`
}

func (handler *userHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	invitationID := chi.URLParam(r, "id")

	invitation, err := handler.invitationStore.FindOneByID(ctx, invitationID)
	if err != nil {
		log.Println(err)
		response.Error(w, apierror.NotFoundError("Invitation id not found"))
		return
	}

	if invitation.Status == store.InvitationStatusUsed {
		log.Println("invitation has been used")
		response.Error(w, apierror.BadRequestError("Invitation has been used to register before"))
		return
	}

	req := CreateNewUserRequest{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, apierror.BadRequestError(err.Error()))
		return
	}

	qrImageName := fmt.Sprintf("qr-%s.png", uuid.NewString())
	err = qrcode.WriteColorFile("https://example.org", qrcode.Medium, 256, color.White, color.RGBA{110, 81, 59, 255}, fmt.Sprintf("./static/qr-codes/%s", qrImageName))
	if err != nil {
		log.Println(err.Error())
		response.Error(w, apierror.BadRequestError(err.Error()))
		return
	}

	newUserData := &store.UserData{
		InvitationID:   invitation.ID,
		InvitationType: invitation.Type,
		WhatsAppNumber: req.WhatsAppNumber,
		QRImageName:    qrImageName,
	}

	if err := handler.userStore.Insert(ctx, newUserData); err != nil {
		log.Println("error insert new user data: %w", err)
		response.Error(w, apierror.InternalServerError())
		return
	}

	whatsAppRegisteredMessage := proto.String(`Nomor WhatsApp anda telah sukses terdaftar. 
		
Terima kasih sudah menyempatkan waktu untuk membuka undangan Resepsi kami.`)
	err = handler.waClient.SendMessage(ctx, newUserData.WhatsAppNumber, &waProto.Message{
		Conversation: whatsAppRegisteredMessage,
	})
	if err != nil {
		log.Println(err)
		response.Error(w, apierror.BadRequestError(err.Error()))
		return
	}

	resp := CreateUserResponse{
		Message: "success",
		UserID:  newUserData.ID,
	}

	response.Respond(w, http.StatusCreated, resp)
}
