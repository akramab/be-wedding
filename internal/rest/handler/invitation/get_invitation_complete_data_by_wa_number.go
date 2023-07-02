package invitation

import (
	"fmt"
	"log"
	"net/http"

	"be-wedding/internal/rest/response"

	"github.com/go-chi/chi/v5"
)

func (handler *invitationHandler) GetInvitationCompleteDataByWANumber(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	waNumber := chi.URLParam(r, "waNumber")

	invitationCompleteData, err := handler.invitationStore.FindOneCompleteDataByWANumber(ctx, waNumber)
	if err != nil {
		log.Println(err)
		resp := GetInvitationCompleteDataResponse{
			Message: "fail",
		}
		response.Respond(w, http.StatusOK, resp)
		return
	}

	var qrImageLink string
	if invitationCompleteData.User.QRImage != "" {
		// still in hard-code
		qrImageLink = fmt.Sprintf("https://api.kramili.site/static/%s", invitationCompleteData.User.QRImage)
	}

	var (
		videoReminder int64
		dateReminder  int64
	)

	if invitationCompleteData.User.IsDateReminderSent {
		dateReminder = 2
	} else if invitationCompleteData.User.Name != "" {
		dateReminder = 1
	}

	if invitationCompleteData.User.IsVideoReminderSent {
		videoReminder = 2
	} else if invitationCompleteData.User.Name != "" {
		videoReminder = 1
	}

	resp := GetInvitationCompleteDataResponse{
		Message: "success",
		Invitation: InvidationData{
			ID:       invitationCompleteData.Invitation.ID,
			Name:     invitationCompleteData.Invitation.Name,
			Type:     invitationCompleteData.Invitation.Type,
			Status:   invitationCompleteData.Invitation.Status,
			Schedule: invitationCompleteData.Invitation.Schedule,
		},
		User: UserData{
			ID:                  invitationCompleteData.User.ID,
			Name:                invitationCompleteData.User.Name,
			WhatsAppNumber:      invitationCompleteData.User.WhatsAppNumber,
			Status:              invitationCompleteData.User.Status,
			QRImageLink:         qrImageLink,
			PeopleCount:         invitationCompleteData.User.PeopleCount,
			IsDateReminderSent:  dateReminder,
			IsVideoReminderSent: videoReminder,
		},
	}

	response.Respond(w, http.StatusOK, resp)
}
