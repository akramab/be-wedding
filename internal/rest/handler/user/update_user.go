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

	userCompletedDataMessage := proto.String(fmt.Sprintf(`Terima kasih %s. 
		
Sampai berjumpa di hari-H Resepsi.`, userData.Name))
	err = handler.waClient.SendMessage(ctx, invitationCompleteData.User.WhatsAppNumber, &waProto.Message{
		Conversation: userCompletedDataMessage,
	})
	if err != nil {
		log.Println(err)
		response.Error(w, apierror.BadRequestError(err.Error()))
		return
	}

	resp := UpdateUserResponse{
		Message: "success",
		UserID:  userData.ID,
		Name:    userData.Name,
	}

	response.Respond(w, http.StatusCreated, resp)
}
