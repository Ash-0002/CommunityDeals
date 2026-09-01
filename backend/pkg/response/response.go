package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIResponse is the standard response envelope for all endpoints.
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
}

// APIError carries structured error detail.
type APIError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// OK sends a 200 response with data.
func OK(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Created sends a 201 response with data.
func Created(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusCreated, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// BadRequest sends a 400 response.
func BadRequest(c *gin.Context, code, message string, details interface{}) {
	c.JSON(http.StatusBadRequest, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

// Unauthorized sends a 401 response.
func Unauthorized(c *gin.Context, code, message string) {
	c.JSON(http.StatusUnauthorized, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
		},
	})
}

// Forbidden sends a 403 response.
func Forbidden(c *gin.Context, code, message string) {
	c.JSON(http.StatusForbidden, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
		},
	})
}

// NotFound sends a 404 response.
func NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    "NOT_FOUND",
			Message: message,
		},
	})
}

// Conflict sends a 409 response.
func Conflict(c *gin.Context, code, message string) {
	c.JSON(http.StatusConflict, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
		},
	})
}

// TooManyRequests sends a 429 response.
func TooManyRequests(c *gin.Context, message string) {
	c.JSON(http.StatusTooManyRequests, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    "RATE_LIMIT_EXCEEDED",
			Message: message,
		},
	})
}

// InternalError sends a 500 response. Never expose internal error details to clients.
func InternalError(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    "INTERNAL_SERVER_ERROR",
			Message: "An unexpected error occurred. Please try again.",
		},
	})
}
