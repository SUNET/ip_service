package configuration

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/SUNET/vc/pkg/logger"
	"github.com/stretchr/testify/assert"
)

func TestParse_MissingEnv(t *testing.T) {
	assert.NoError(t, os.Unsetenv("CONFIG_YAML"))

	_, err := Parse(t.Context(), logger.NewSimple("test"))
	assert.Error(t, err)
}

func TestParse_FileMissing(t *testing.T) {
	assert.NoError(t, os.Setenv("CONFIG_YAML", "/no/such/path.yaml"))
	t.Cleanup(func() { _ = os.Unsetenv("CONFIG_YAML") })

	_, err := Parse(t.Context(), logger.NewSimple("test"))
	assert.Error(t, err)
}

func TestParse_PathIsDirectory(t *testing.T) {
	dir := t.TempDir()
	assert.NoError(t, os.Setenv("CONFIG_YAML", dir))
	t.Cleanup(func() { _ = os.Unsetenv("CONFIG_YAML") })

	_, err := Parse(t.Context(), logger.NewSimple("test"))
	assert.Error(t, err)
}

func TestParse_InvalidYAML(t *testing.T) {
	f := filepath.Join(t.TempDir(), "bad.yaml")
	assert.NoError(t, os.WriteFile(f, []byte(":\n:this: is not: valid: [yaml"), 0600))
	assert.NoError(t, os.Setenv("CONFIG_YAML", f))
	t.Cleanup(func() { _ = os.Unsetenv("CONFIG_YAML") })

	_, err := Parse(t.Context(), logger.NewSimple("test"))
	assert.Error(t, err)
}

func TestParse_ValidationFail(t *testing.T) {
	// Valid YAML that unmarshals but does not satisfy required fields.
	f := filepath.Join(t.TempDir(), "empty.yaml")
	assert.NoError(t, os.WriteFile(f, []byte("ip_service:\n  api_server: {}\n"), 0600))
	assert.NoError(t, os.Setenv("CONFIG_YAML", f))
	t.Cleanup(func() { _ = os.Unsetenv("CONFIG_YAML") })

	_, err := Parse(t.Context(), logger.NewSimple("test"))
	assert.Error(t, err)
}
