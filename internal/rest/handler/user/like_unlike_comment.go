package user

import (
	"encoding/json"
	"log"
	"net/http"

	apierror "be-wedding/internal/rest/error"

	"be-wedding/internal/rest/response"

	"github.com/go-chi/chi/v5"
)

type LikeUnlikeCommentRequest struct {
	CommentID string `json:"comment_id"`
}

type LikeUnlikeCommentResponse struct {
	Message   string `json:"message"`
	CommentID string `json:"comment_id"`
	IsLiked   bool   `json:"is_liked"`
}

func (handler *userHandler) LikeUnlikeComment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := chi.URLParam(r, "id")

	req := LikeUnlikeCommentRequest{}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, apierror.BadRequestError(err.Error()))
		return
	}

	isLiked, err := handler.userStore.LikeUnlikeComment(ctx, userID, req.CommentID)
	if err != nil {
		log.Println("error like or unlike user comment: %w", err)
		response.Error(w, apierror.InternalServerError())
		return
	}

	resp := LikeUnlikeCommentResponse{
		Message:   "success",
		CommentID: req.CommentID,
		IsLiked:   isLiked,
	}

	response.Respond(w, http.StatusCreated, resp)
}
