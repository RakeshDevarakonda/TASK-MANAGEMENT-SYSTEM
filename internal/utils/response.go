package utils

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type ValidationErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type PaginationMetadata struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

type CursorMetadata struct {
	Limit      int     `json:"limit"`
	NextCursor *string `json:"next_cursor"`
}

type ResponseEnvelope struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Meta    interface{} `json:"meta,omitempty"`
}

// SendSuccess sends a standardized success JSON response
func SendSuccess(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, ResponseEnvelope{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// SendPaginatedSuccess sends a standardized paginated success JSON response
func SendPaginatedSuccess(c *gin.Context, statusCode int, message string, data interface{}, page int, perPage int, total int) {
	totalPages := 0
	if perPage > 0 {
		totalPages = (total + perPage - 1) / perPage
	}
	c.JSON(statusCode, ResponseEnvelope{
		Success: true,
		Message: message,
		Data:    data,
		Meta: &PaginationMetadata{
			Page:       page,
			PerPage:    perPage,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

// SendCursorSuccess sends a standardized cursor paginated success JSON response
func SendCursorSuccess(c *gin.Context, statusCode int, message string, data interface{}, limit int, nextCursor string) {
	var next *string
	if nextCursor != "" {
		next = &nextCursor
	}
	c.JSON(statusCode, ResponseEnvelope{
		Success: true,
		Message: message,
		Data:    data,
		Meta: &CursorMetadata{
			Limit:      limit,
			NextCursor: next,
		},
	})
}

// SendError sends a standardized error JSON response
func SendError(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, ResponseEnvelope{
		Success: false,
		Message: message,
		Data:    data,
	})
}

// SendValidationError parses validator errors and sends a standardized bad request response with details
func SendValidationError(c *gin.Context, err error) {
	var details []ValidationErrorDetail

	if errs, ok := err.(validator.ValidationErrors); ok {
		for _, f := range errs {
			field := strings.ToLower(f.Field())
			var msg string

			switch f.Tag() {
			case "required":
				msg = field + " is required"
			case "email":
				msg = "invalid email format"
			case "min":
				msg = field + " must be at least " + f.Param() + " characters"
			default:
				msg = field + " validation failed on " + f.Tag()
			}

			details = append(details, ValidationErrorDetail{
				Field:   field,
				Message: msg,
			})
		}
	} else {
		// Handle non-validator JSON binding errors (e.g. unknown fields or bad syntax)
		errStr := err.Error()
		if strings.HasPrefix(errStr, "json: unknown field ") {
			fieldName := strings.TrimPrefix(errStr, "json: unknown field ")
			fieldName = strings.Trim(fieldName, `"`)
			msg := "field '" + fieldName + "' is not allowed"
			
			details = append(details, ValidationErrorDetail{
				Field:   fieldName,
				Message: msg,
			})
		} else if strings.Contains(errStr, "syntax error") || strings.Contains(errStr, "invalid character") {
			details = append(details, ValidationErrorDetail{
				Field:   "body",
				Message: "invalid JSON payload",
			})
		}
	}

	var mainMessage string
	if len(details) > 0 {
		mainMessage = details[0].Message
	} else {
		mainMessage = err.Error()
	}

	c.JSON(http.StatusBadRequest, ResponseEnvelope{
		Success: false,
		Message: mainMessage,
		Data:    details,
	})
}
