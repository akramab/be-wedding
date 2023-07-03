package user

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	apierror "be-wedding/internal/rest/error"
	"be-wedding/internal/store"

	"be-wedding/internal/rest/response"

	"github.com/go-chi/chi/v5"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"google.golang.org/protobuf/proto"
)

type CreateUserRSVPRequest struct {
	PeopleCount int64 `json:"people_count"`
}

type CreateUserRSVPResponse struct {
	Message     string `json:"message"`
	UserRSVPID  string `json:"user_rsvp_id"`
	PeopleCount int64  `json:"people_count"`
}

func (handler *userHandler) CreateUserRSVP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := chi.URLParam(r, "id")

	req := CreateUserRSVPRequest{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, apierror.BadRequestError(err.Error()))
		return
	}

	userRSVP := &store.UserRSVPData{
		UserID:      userID,
		PeopleCount: req.PeopleCount,
	}

	if err := handler.userStore.InsertUserRSVP(ctx, userRSVP); err != nil {
		log.Println("error insert new user rsvp data: %w", err)
		response.Error(w, apierror.InternalServerError())
		return
	}

	invitationCompleteData, err := handler.invitationStore.FindOneCompleteDataByUserID(ctx, userID)
	if err != nil {
		log.Println(err)
		response.Error(w, apierror.NotFoundError("User id not found"))
		return
	}

	userRSVPMessage := proto.String(fmt.Sprintf(`Terima kasih telah mengisi konfirmasi kehadiran

Berikut ini rekap rencana kehadiran yang tercatat:

*Nama*			: %s
*Jumlah Orang*	: %d

Berikut ini kami lampirkan pula kode QR sebagai tiket masuk anda`, invitationCompleteData.User.Name, userRSVP.PeopleCount))
	err = handler.waClient.SendMessage(ctx, invitationCompleteData.User.WhatsAppNumber, &waProto.Message{
		Conversation: userRSVPMessage,
	})
	if err != nil {
		log.Println(err)
		response.Error(w, apierror.BadRequestError(err.Error()))
		return
	}

	captionImageMessage := `Tunjukkan code QR saat hendak memasuki venue pada hari H.`
	err = handler.waClient.SendImageMessage(ctx, invitationCompleteData.User.WhatsAppNumber, invitationCompleteData.User.QRImage, captionImageMessage)
	if err != nil {
		log.Println(err)
		response.Error(w, apierror.BadRequestError(err.Error()))
		return
	}

	resp := CreateUserRSVPResponse{
		Message:     "success",
		UserRSVPID:  userRSVP.ID,
		PeopleCount: userRSVP.PeopleCount,
	}

	response.Respond(w, http.StatusCreated, resp)
}
