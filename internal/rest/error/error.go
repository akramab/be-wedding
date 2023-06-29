package error

import (
	"fmt"
	"net/http"
	"strings"
)

type Error struct {
	StatusCode int    `json:"-"`
	Message    string `json:"message"`
}

func (e *Error) Error() string {
	return fmt.Sprintf("%d: %s", e.StatusCode, e.Message)
}

func NotFoundError(message string) Error {
	return Error{
		StatusCode: http.StatusNotFound,
		Message:    message,
	}
}

func ForbiddenError(message string) Error {
	return Error{
		StatusCode: http.StatusForbidden,
		Message:    message,
	}
}

func UnauthorizedError(message string) Error {
	return Error{
		StatusCode: http.StatusUnauthorized,
		Message:    message,
	}
}

func BadRequestError(message string) Error {
	return Error{
		StatusCode: http.StatusBadRequest,
		Message:    message,
	}
}

func InternalServerError() Error {
	return Error{
		StatusCode: http.StatusInternalServerError,
		Message:    "internal server error",
	}
}

type InvalidField struct {
	Name    string `json:"name"`
	Message string `json:"message"`
}

type FieldError struct {
	StatusCode int            `json:"-"`
	Fields     []InvalidField `json:"fields"`
}

func (e *FieldError) Error() string {
	fieldNames := []string{}
	for _, field := range e.Fields {
		fieldNames = append(fieldNames, field.Name)
	}

	return fmt.Sprintf("%d: %s", e.StatusCode, strings.Join(fieldNames, ", "))
}

func (err FieldError) WithField(fieldName string, message string) FieldError {
	newErr := err
	newErr.Fields = append(newErr.Fields, InvalidField{
		Name:    fieldName,
		Message: message,
	})

	return newErr
}

func NewFieldError() FieldError {
	return FieldError{
		StatusCode: http.StatusUnprocessableEntity,
	}
}
