package user

import (
	apierror "be-wedding/internal/rest/error"
	"be-wedding/internal/rest/response"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
)

type M map[string]interface{}

func (handler *userHandler) DownloadQRCode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userID := chi.URLParam(r, "id")
	basePath, _ := os.Getwd()

	invitationCompleteData, err := handler.invitationStore.FindOneCompleteDataByUserID(ctx, userID)
	if err != nil {
		log.Println(err)
		response.Error(w, apierror.NotFoundError("User id not found"))
		return
	}

	fileLocation := filepath.Join(basePath, "./static/qr-codes/"+invitationCompleteData.User.QRImage)

	f, err := os.Open(fileLocation)
	if f != nil {
		defer f.Close()
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println("mau ngetes bang")
	log.Println(f.Name())
	contentDisposition := fmt.Sprintf("attachment; filename=%s", "qr_"+strings.ReplaceAll(strings.ToLower(invitationCompleteData.User.Name), " ", "_"))
	w.Header().Set("Content-Disposition", contentDisposition)

	if _, err := io.Copy(w, f); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
