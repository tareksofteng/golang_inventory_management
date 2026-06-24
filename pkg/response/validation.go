package response

import (
	"errors"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// RegisterValidatorJSONTags makes validation error keys use the json field
// name (e.g. "category_id") instead of the Go struct field name ("CategoryID").
// Call it ONCE at startup. Without it the error map would expose Go field names
// to API clients, which is leaky and inconsistent with the rest of the JSON.
func RegisterValidatorJSONTags() {
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return
	}
	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

// ValidationErrors turns the validator's error into a flat field->message map,
// e.g. {"name": "This field is required", "email": "Must be a valid email
// address"}. If err is not a validation error, it returns an empty map.
func ValidationErrors(err error) map[string]string {
	fields := make(map[string]string)

	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		for _, fe := range ve {
			fields[fe.Field()] = messageForTag(fe)
		}
	}
	return fields
}

// messageForTag returns a human-friendly message for each validation rule.
// Extend this switch as you add new binding rules.
func messageForTag(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Must be a valid email address"
	case "min":
		return "Must be at least " + fe.Param() + " character(s)"
	case "max":
		return "Must not exceed " + fe.Param() + " character(s)"
	case "gt":
		return "Must be greater than " + fe.Param()
	case "gte":
		return "Must be greater than or equal to " + fe.Param()
	case "lte":
		return "Must be less than or equal to " + fe.Param()
	default:
		return "Invalid value"
	}
}
