package helpers

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/moogar0880/problems"
	"github.com/stretchr/testify/assert"
)

func TestError_Error(t *testing.T) {
	var nilErr *Error
	assert.Equal(t, "", nilErr.Error())

	e := &Error{Title: "boom"}
	assert.Equal(t, "Error: [boom]", e.Error())

	e2 := &Error{Title: "boom", Details: map[string]any{"field": "ip"}}
	assert.Contains(t, e2.Error(), "boom")
	assert.Contains(t, e2.Error(), "field")
}

func TestNewErrorHelpers(t *testing.T) {
	e := NewError("id")
	assert.Equal(t, "id", e.Title)
	assert.Nil(t, e.Details)

	e2 := NewErrorDetails("id", "x")
	assert.Equal(t, "id", e2.Title)
	assert.Equal(t, "x", e2.Details)
}

func TestNewErrorFromError_Nil(t *testing.T) {
	assert.Nil(t, NewErrorFromError(nil))
}

func TestNewErrorFromError_PassthroughOwnError(t *testing.T) {
	src := &Error{Title: "own"}
	got := NewErrorFromError(src)
	assert.Same(t, src, got)
}

func TestNewErrorFromError_JSONSyntax(t *testing.T) {
	// Trigger a *json.SyntaxError with invalid JSON.
	var v map[string]any
	err := json.Unmarshal([]byte("{invalid"), &v)
	assert.Error(t, err)

	got := NewErrorFromError(err)
	assert.Equal(t, "json_syntax_error", got.Title)
	details, ok := got.Details.(map[string]any)
	assert.True(t, ok)
	assert.Contains(t, details, "position")
	assert.Contains(t, details, "error")
}

func TestNewErrorFromError_JSONUnmarshalType(t *testing.T) {
	// Trigger a *json.UnmarshalTypeError.
	type target struct {
		N int `json:"n"`
	}
	var v target
	err := json.Unmarshal([]byte(`{"n":"not-a-number"}`), &v)
	assert.Error(t, err)

	got := NewErrorFromError(err)
	assert.Equal(t, "json_type_error", got.Title)
	details, ok := got.Details.([]map[string]any)
	assert.True(t, ok)
	assert.Len(t, details, 1)
	assert.Equal(t, "n", details[0]["field"])
}

func TestNewErrorFromError_ValidationError(t *testing.T) {
	type inp struct {
		IP string `json:"ip" validate:"required,ip"`
	}
	err := Check(inp{IP: "not-an-ip"})
	assert.Error(t, err)

	// Check wraps the validator error into helpers.Error already.
	e, ok := err.(*Error)
	assert.True(t, ok)
	assert.Equal(t, "validation_error", e.Title)

	details, ok := e.Details.([]map[string]any)
	assert.True(t, ok)
	assert.NotEmpty(t, details)
	assert.Equal(t, "ip", details[0]["field"])
	assert.Equal(t, "ip", details[0]["validation"])
}

func TestNewErrorFromError_Generic(t *testing.T) {
	got := NewErrorFromError(errors.New("boom"))
	assert.Equal(t, "internal_server_error", got.Title)
	assert.Equal(t, "boom", got.Details)
}

func TestCheck_OK(t *testing.T) {
	type inp struct {
		IP string `json:"ip" validate:"required,ip"`
	}
	err := Check(inp{IP: "127.0.0.1"})
	assert.NoError(t, err)
}

func TestCheck_JSONDashTag(t *testing.T) {
	type inp struct {
		Secret string `json:"-" validate:"required"`
	}
	// When json tag is "-" the field name lookup returns empty; validator
	// still validates against the Go field name. Confirm it produces an error.
	err := Check(inp{})
	assert.Error(t, err)
}

func TestProblem404(t *testing.T) {
	p := Problem404()
	assert.NotNil(t, p)
	// The problems package should give us a 404 status.
	pb, ok := any(p).(*problems.Problem)
	assert.True(t, ok)
	assert.Equal(t, 404, pb.Status)
}

func TestErrors_Sentinels(t *testing.T) {
	assert.Equal(t, "not a valid endpoint", ErrNotValidEndpoint.Error())
	assert.Equal(t, "missing DB file", ErrMissingDBFile.Error())
	assert.Equal(t, "ip not found", ErrIpNotFound.Error())
}
