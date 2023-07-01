package invitation

import (
	"fmt"
	"log"
	"net/http"

	apierror "be-wedding/internal/rest/error"

	"be-wedding/internal/rest/response"

	"github.com/go-chi/chi/v5"
)

func (handler *invitationHandler) GetInvitationCompleteDataByWANumber(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	waNumber := chi.URLParam(r, "waNumber")

	invitationCompleteData, err := handler.invitationStore.FindOneCompleteDataByWANumber(ctx, waNumber)
	if err != nil {
		log.Println(err)
		response.Error(w, apierror.NotFoundError("wa number not found"))
		return
	}

	var qrImageLink string
	if invitationCompleteData.User.QRImage != "" {
		// still in hard-code
		qrImageLink = fmt.Sprintf("https://api.kramili.site/static/%s", invitationCompleteData.User.QRImage)
	}
	resp := GetInvitationCompleteDataResponse{
		Invitation: InvidationData{
			ID:       invitationCompleteData.Invitation.ID,
			Name:     invitationCompleteData.Invitation.Name,
			Type:     invitationCompleteData.Invitation.Type,
			Status:   invitationCompleteData.Invitation.Status,
			Schedule: invitationCompleteData.Invitation.Schedule,
		},
		User: UserData{
			ID:             invitationCompleteData.User.ID,
			Name:           invitationCompleteData.User.Name,
			WhatsAppNumber: invitationCompleteData.User.WhatsAppNumber,
			Status:         invitationCompleteData.User.Status,
			QRImageLink:    qrImageLink,
			PeopleCount:    invitationCompleteData.User.PeopleCount,
		},
	}

	response.Respond(w, http.StatusOK, resp)
}
