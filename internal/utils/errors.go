package utils

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrForbidden    = errors.New("forbidden")
	ErrConflict     = errors.New("conflict")
	ErrUnauthorized = errors.New("unauthorized")
)

// SendErrorResponse maps standard sentinel errors to their correct HTTP status codes
func SendErrorResponse(c *gin.Context, err error) {
	if err == nil {
		SendError(c, http.StatusInternalServerError, "unknown error", nil)
		return
	}

	statusCode := http.StatusInternalServerError
	message := err.Error()

	// Unwrap/trace sentinel errors
	if errors.Is(err, ErrNotFound) {
		statusCode = http.StatusNotFound
	} else if errors.Is(err, ErrForbidden) {
		statusCode = http.StatusForbidden
	} else if errors.Is(err, ErrConflict) {
		statusCode = http.StatusConflict
	} else if errors.Is(err, ErrUnauthorized) {
		statusCode = http.StatusUnauthorized
	}

	SendError(c, statusCode, message, nil)
}
