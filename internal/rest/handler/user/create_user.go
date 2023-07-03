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
	"github.com/google/uuid"
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
	newUserData := &store.UserData{
		InvitationID:   invitation.ID,
		InvitationType: invitation.Type,
		WhatsAppNumber: req.WhatsAppNumber,
		QRImageName:    qrImageName,
	}

	if err := handler.userStore.Insert(ctx, newUserData); err != nil {
		log.Println("error insert new user data: %w", err)
		response.Error(w, apierror.BadRequestError(fmt.Sprintf("phone number: %s already exists!", newUserData.WhatsAppNumber)))
		return
	}

	resp := CreateUserResponse{
		Message: "success",
		UserID:  newUserData.ID,
	}

	response.Respond(w, http.StatusCreated, resp)
}
