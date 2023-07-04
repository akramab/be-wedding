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

type RemindUserSendWeddingVideoResponse struct {
	Message string `json:"message"`
}

func (handler *userHandler) RemindUserSendWeddingVideo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := chi.URLParam(r, "id")

	invitationCompleteData, err := handler.invitationStore.FindOneCompleteDataByUserID(ctx, userID)
	if err != nil {
		log.Println(err)
		response.Error(w, apierror.NotFoundError("User id not found"))
		return
	}

	err = handler.invitationStore.UpdateVideoReminder(ctx, invitationCompleteData)
	if err != nil {
		log.Println(err)
		response.Error(w, apierror.InternalServerError())
	}

	reminderMessage := proto.String(`Terima kasih telah bersedia mengirim ucapan
		
*Kami tidak berkenan menerima karangan bunga secara fisik*. Namun, kami sangat menantikan ucapan berupa foto atau video yang akan ditampilkan pada hari pernikahan dengan ketentuan:

 *1) Ketentuan Foto* 
a. Foto dapat berupa poster, ucapan selamat, ataupun jenis foto yang lain
b. Format foto dalam .png, .jpg, atau .pdf
c. Foto dibuat dalam layout landscape
d. Ukuran dimensi foto dibebaskan


 *2) Ketentuan Video* 
a. Video dapat berupa film pendek, vlog, video musik, parodi, ataupun jenis video yang lain
b. Format .mp4, .mkv atau .mov
c. Video dibuat dalam layout landscape
d. Durasi maksimal 1 menit

Foto dan/atau Video dapat dikirimkan melalui nomer WhatsApp ini


*Ketik angka 23 jika anda ingin mengirim foto atau video*`)
	err = handler.waClient.SendMessage(ctx, invitationCompleteData.User.WhatsAppNumber, &waProto.Message{
		Conversation: reminderMessage,
	})
	if err != nil {
		log.Println(err)
		response.Error(w, apierror.BadRequestError(err.Error()))
		return
	}

	resp := RemindUserSendWeddingVideoResponse{
		Message: "reminder and detailed info about the video will be sent to the provided WhatsApp number",
	}

	response.Respond(w, http.StatusCreated, resp)
}
