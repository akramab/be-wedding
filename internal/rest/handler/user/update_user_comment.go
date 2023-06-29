package user

import (
	"encoding/json"
	"log"
	"net/http"

	apierror "be-wedding/internal/rest/error"
	"be-wedding/internal/store"

	"be-wedding/internal/rest/response"

	"github.com/go-chi/chi/v5"
)

type UpdateUserCommentRequest struct {
	Comment string `json:"comment"`
}

type UpdateUserCommentResponse struct {
	Message   string `json:"message"`
	CommentID string `json:"comment_id"`
	Comment   string `json:"comment"`
}

func (handler *userHandler) UpdateUserComment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commentID := chi.URLParam(r, "id")

	req := UpdateUserCommentRequest{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, apierror.BadRequestError(err.Error()))
		return
	}

	userCommentData := &store.UserCommentData{
		ID:      commentID,
		Comment: req.Comment,
	}

	if err := handler.userStore.UpdateComment(ctx, userCommentData); err != nil {
		log.Println("error update user comment data: %w", err)
		response.Error(w, apierror.InternalServerError())
		return
	}

	resp := UpdateUserCommentResponse{
		Message:   "success",
		CommentID: userCommentData.ID,
		Comment:   userCommentData.Comment,
	}

	response.Respond(w, http.StatusCreated, resp)
}
