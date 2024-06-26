package helpers

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/moogar0880/problems"
)

// Error is a struct that represents an error
type Error struct {
	Title   string      `json:"title" `
	Details interface{} `json:"details" xml:"details"`
}

// Error is the string representation of the Error struct
func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Details == nil {
		return fmt.Sprintf("Error: [%s]", e.Title)
	}
	return fmt.Sprintf("Error: [%s] %+v", e.Title, e.Details)
}

// NewError creates a new Error
func NewError(id string) *Error {
	return &Error{Title: id}
}

// NewErrorDetails creates a new Error with details
func NewErrorDetails(id string, details interface{}) *Error {
	return &Error{Title: id, Details: details}
}

// NewErrorFromError creates a new Error from an error
func NewErrorFromError(err error) *Error {
	if err == nil {
		return nil
	}
	if pbErr, ok := err.(*Error); ok {
		return pbErr
	}
	if jsonUnmarshalTypeError, ok := err.(*json.UnmarshalTypeError); ok {
		return &Error{Title: "json_type_error", Details: formatJSONUnmarshalTypeError(jsonUnmarshalTypeError)}
	}
	if jsonSyntaxError, ok := err.(*json.SyntaxError); ok {
		return &Error{Title: "json_syntax_error", Details: map[string]any{"position": jsonSyntaxError.Offset, "error": jsonSyntaxError.Error()}}
	}
	if validatorErr, ok := err.(validator.ValidationErrors); ok {
		return &Error{Title: "validation_error", Details: formatValidationErrors(validatorErr)}
	}

	return NewErrorDetails("internal_server_error", err.Error())
}

func formatValidationErrors(err validator.ValidationErrors) []map[string]any {
	v := make([]map[string]any, 0)
	for _, e := range err {
		splits := strings.SplitN(e.Namespace(), ".", 2)
		v = append(v, map[string]any{
			"field":           splits[1],
			"struct":          splits[0],
			"type":            e.Kind().String(),
			"validation":      e.Tag(),
			"validationParam": e.Param(),
			"value":           e.Value(),
		})
	}
	return v
}

func formatJSONUnmarshalTypeError(err *json.UnmarshalTypeError) []map[string]any {
	return []map[string]any{
		{
			"field":    err.Field,
			"expected": err.Type.Kind().String(),
			"actual":   err.Value,
		},
	}
}

// Problem404 creates a new 404 problem
func Problem404() (*problems.DefaultProblem, error) {
	notFound := problems.NewDetailedProblem(404, "Not a valid endpoint")
	if err := problems.ValidateProblem(notFound); err != nil {
		return nil, err
	}

	return notFound, nil
}
