package helpers

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// defaultValidator is a process-wide validator reused by Check to avoid
// per-call allocation on hot request paths.
var defaultValidator = newValidator()

func newValidator() *validator.Validate {
	validate := validator.New()
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name, _, _ := strings.Cut(fld.Tag.Get("json"), ",")

		if name == "-" {
			return ""
		}

		return name
	})

	return validate
}

// Check checks for validation error
func Check(s any) error {
	if err := defaultValidator.Struct(s); err != nil {
		return NewErrorFromError(err)
	}

	return nil
}
