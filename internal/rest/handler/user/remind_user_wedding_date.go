package user

import (
	"log"
	"net/http"

	apierror "be-wedding/internal/rest/error"
	"be-wedding/internal/rest/response"

	"github.com/go-chi/chi/v5"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"google.golang.org/protobuf/proto"
)

type RemindUserWeddingDateResponse struct {
	Message string `json:"message"`
}

func (handler *userHandler) RemindUserWeddingDate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := chi.URLParam(r, "id")

	invitationCompleteData, err := handler.invitationStore.FindOneCompleteDataByUserID(ctx, userID)
	if err != nil {
		log.Println(err)
		response.Error(w, apierror.NotFoundError("User id not found"))
		return
	}

	err = handler.invitationStore.UpdateDateReminder(ctx, invitationCompleteData)
	if err != nil {
		log.Println(err)
		response.Error(w, apierror.InternalServerError())
	}

	reminderMessage := proto.String(`Terima kasih telah meluangkan waktu 
		
Kami akan mengirimkan reminder pada H-7 dan H-1 acara resepsi ☺️`)
	err = handler.waClient.SendMessage(ctx, invitationCompleteData.User.WhatsAppNumber, &waProto.Message{
		Conversation: reminderMessage,
	})
	if err != nil {
		log.Println(err)
		response.Error(w, apierror.BadRequestError(err.Error()))
		return
	}

	resp := RemindUserWeddingDateResponse{
		Message: "reminder will be sent to the provided WhatsApp number",
	}

	response.Respond(w, http.StatusCreated, resp)
}
