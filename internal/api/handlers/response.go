package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response represents a standard API response
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

// SuccessResponse sends a successful response
func SuccessResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
	})
}

// SuccessResponseWithMessage sends a successful response with a message
func SuccessResponseWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// CreatedResponse sends a 201 Created response
func CreatedResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Success: true,
		Data:    data,
	})
}

// ErrorResponse sends an error response
func ErrorResponse(c *gin.Context, statusCode int, err error) {
	c.JSON(statusCode, Response{
		Success: false,
		Error:   err.Error(),
	})
}

// ErrorResponseWithMessage sends an error response with a custom message
func ErrorResponseWithMessage(c *gin.Context, statusCode int, message string) {
	c.JSON(statusCode, Response{
		Success: false,
		Error:   message,
	})
}

// BadRequestError sends a 400 Bad Request response
func BadRequestError(c *gin.Context, err error) {
	ErrorResponse(c, http.StatusBadRequest, err)
}

// NotFoundError sends a 404 Not Found response
func NotFoundError(c *gin.Context, message string) {
	ErrorResponseWithMessage(c, http.StatusNotFound, message)
}

// InternalServerError sends a 500 Internal Server Error response
func InternalServerError(c *gin.Context, err error) {
	ErrorResponse(c, http.StatusInternalServerError, err)
}
