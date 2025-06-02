package validation

import (
	"driftGo/api/common/errors"
	"net/http"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// ValidateRequest validates a request struct and returns validation errors if any
func ValidateRequest(w http.ResponseWriter, request interface{}) bool {
	if err := validate.Struct(request); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, e := range validationErrors {
				switch e.Tag() {
				case "required":
					errors.ValidationErrorHandler(w, e.Field()+" is required")
				case "email":
					errors.ValidationErrorHandler(w, "Invalid email format")
				case "min":
					errors.ValidationErrorHandler(w, e.Field()+" must be at least "+e.Param())
				case "max":
					errors.ValidationErrorHandler(w, e.Field()+" must be at most "+e.Param())
				case "oneof":
					errors.ValidationErrorHandler(w, e.Field()+" must be one of: "+e.Param())
				default:
					errors.ValidationErrorHandler(w, "Invalid "+e.Field())
				}
				return false
			}
		}
		errors.ValidationErrorHandler(w, "Invalid request format")
		return false
	}
	return true
}
