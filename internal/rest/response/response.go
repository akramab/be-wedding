package response

import (
	"encoding/json"
	"net/http"

	err "be-wedding/internal/rest/error"
)

type CommonMessage struct {
	Message string `json:"message"`
}

func Respond(w http.ResponseWriter, code int, payload interface{}) {
	response, _ := json.Marshal(payload)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func RespondSuccess(w http.ResponseWriter) {
	Respond(w, http.StatusOK, CommonMessage{"success"})
}

func Error(w http.ResponseWriter, err err.Error) {
	Respond(w, err.StatusCode, err)
}

func FieldError(w http.ResponseWriter, err err.FieldError) {
	Respond(w, err.StatusCode, err)
}
