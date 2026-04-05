package response

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type ListResponse[T any] struct {
	Data []T            `json:"data"`
	Meta PaginationMeta `json:"meta"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

type ValidationErrorResponse struct {
	Message string            `json:"message"`
	Details []ValidationField `json:"details"`
}

type ValidationField struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// FieldErr is an error type for field-level errors.
// Service returns this, handler detects and renders automatically via HandleError.
type FieldErr struct {
	Fields []ValidationField
}

func (e *FieldErr) Error() string {
	if len(e.Fields) == 1 {
		return fmt.Sprintf("%s: %s", e.Fields[0].Field, e.Fields[0].Message)
	}
	return "validation failed"
}

// NewFieldErr creates a FieldErr with one field.
func NewFieldErr(field string, message string) *FieldErr {
	return &FieldErr{Fields: []ValidationField{{Field: field, Message: message}}}
}

// NewFieldErrs creates a FieldErr with multiple fields.
func NewFieldErrs(fields ...ValidationField) *FieldErr {
	return &FieldErr{Fields: fields}
}

func Success[T any](c *gin.Context, data T) {
	c.JSON(200, data)
}

func List[T any](c *gin.Context, data []T, meta PaginationMeta) {
	if data == nil {
		data = []T{}
	}
	c.JSON(200, ListResponse[T]{Data: data, Meta: meta})
}

func BadRequest(c *gin.Context, message string) {
	c.JSON(400, ErrorResponse{Message: message})
}

func Unauthorized(c *gin.Context, message string) {
	c.JSON(401, ErrorResponse{Message: message})
}

func Forbidden(c *gin.Context, message string) {
	c.JSON(403, ErrorResponse{Message: message})
}

func NotFound(c *gin.Context, message string) {
	c.JSON(404, ErrorResponse{Message: message})
}

// HandleError detects error type and renders appropriate response.
// Use this in handler for service errors.
func HandleError(c *gin.Context, err error) {
	var fe *FieldErr
	if errors.As(err, &fe) {
		c.JSON(400, ValidationErrorResponse{
			Message: "validation failed",
			Details: fe.Fields,
		})
		return
	}
	c.JSON(400, ErrorResponse{Message: err.Error()})
}

func ValidationError(c *gin.Context, err error) {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		details := make([]ValidationField, len(ve))
		for i, fe := range ve {
			details[i] = ValidationField{
				Field:   fe.Field(),
				Message: validationMessage(fe),
			}
		}
		c.JSON(400, ValidationErrorResponse{
			Message: "validation failed",
			Details: details,
		})
		return
	}

	c.JSON(400, ErrorResponse{Message: err.Error()})
}

func validationMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fe.Field() + " is required"
	case "email":
		return fe.Field() + " must be a valid email"
	case "min":
		return fe.Field() + " must be at least " + fe.Param() + " characters"
	case "max":
		return fe.Field() + " must be at most " + fe.Param() + " characters"
	case "uuid":
		return fe.Field() + " must be a valid UUID"
	case "strong_password":
		return fe.Field() + " must be at least 8 characters with 1 uppercase, 1 lowercase, and 1 special character"
	default:
		return fe.Field() + " is invalid"
	}
}
