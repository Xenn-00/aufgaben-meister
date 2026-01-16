package app_errors

import (
	"errors"
	"strings"
	"unicode"

	"github.com/go-playground/validator/v10"
)

func ParseValidationError(err error) []FieldError {
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return nil
	}

	var out []FieldError
	for _, fe := range ve {
		msgKey, params := validationMessageKey(fe)

		out = append(out, FieldError{
			Field:      toSnakeCase(fe.Field()),
			Reason:     fe.Tag(),
			MessageKey: msgKey,
			Params:     params,
		})
	}

	return out
}

func toSnakeCase(s string) string {
	var b strings.Builder
	for i, r := range s {
		if unicode.IsUpper(r) {
			if i > 0 {
				b.WriteByte('_')
			}
			b.WriteRune(unicode.ToLower(r))
		} else {
			b.WriteRune(r)
		}
	}
	return strings.ReplaceAll(b.String(), " ", "_")
}

func validationMessageKey(fe validator.FieldError) (string, map[string]interface{}) {
	switch fe.Tag() {
	case "required":
		return "validation.required", nil
	case "min":
		return "validation.min", map[string]interface{}{
			"min": fe.Param(),
		}
	case "max":
		return "validation.max", map[string]interface{}{
			"max": fe.Param(),
		}
	case "email":
		return "validation.email", nil
	case "typeProject":
		return "validation.type_project", nil
	case "visibility":
		return "validation.visibility", nil
	default:
		return "validation.invalid", nil
	}
}
