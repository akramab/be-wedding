package user

import (
	"be-wedding/internal/rest/response"
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"google.golang.org/protobuf/proto"
)

type ValidateUserQRRsvpResponse struct {
	Name        string `json:"name"`
	PeopleCount int    `json:"people_count"`
	VIP         bool   `json:"vip"`
}

const (
	GetCurrentAdmin1 = "CURRENT_ADMIN_1"
	GetCurrentAdmin2 = "CURRENT_ADMIN_2"

	StringSeparator = ","
)

func (handler *userHandler) ValidateUserQRRsvp(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	invitationCode := chi.URLParam(r, "invitation_code")
	atCode := r.URL.Query().Get("code")

	var adminList []string
	var resp ValidateUserQRRsvpResponse
	if atCode == "1" {
		adminListString, _ := handler.redisCache.Get(ctx, GetCurrentAdmin1).Result()
		if adminListString == "" {
			adminListString = strings.Join([]string{"+62812155249"}, StringSeparator) // random number
		}

		adminList = strings.Split(adminListString, ",")
	} else {
		adminListString, _ := handler.redisCache.Get(ctx, GetCurrentAdmin2).Result()
		if adminListString == "" {
			adminListString = strings.Join([]string{"+62812155249"}, StringSeparator) // random number
		}

		adminList = strings.Split(adminListString, ",")
	}

	if invitationCode == "0c467423-e324-4786-a9d9-0c77eb267407" {
		resp = ValidateUserQRRsvpResponse{
			Name:        "Bapak Ari",
			PeopleCount: 2,
			VIP:         true,
		}
	}

	if invitationCode == "69fee15d-2c45-48ea-982e-4ce6327298fc" {
		resp = ValidateUserQRRsvpResponse{
			Name:        "Bapak Lukman",
			PeopleCount: 10,
			VIP:         false,
		}
	} else {
		resp = ValidateUserQRRsvpResponse{
			Name:        "Tamu Tidak Diundang",
			PeopleCount: -99999,
			VIP:         false,
		}
	}

	textForAdmin := `Konfirmasi Kehadiran Berhasil!

*Berikut data tamu undangan*

Nama: %s
Jumlah Konfirmasi (orang): 	%d
VIP: %s`
	for _, admin := range adminList {
		handler.waClient.SendMessage(context.Background(), admin, &waProto.Message{
			Conversation: proto.String(fmt.Sprintf(textForAdmin,
				resp.Name,
				resp.PeopleCount,
				strconv.FormatBool(resp.VIP),
			)),
		})
	}

	response.Respond(w, http.StatusOK, resp)
	return
}
