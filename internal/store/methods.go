package store

import (
	"context"
)

// Set sets key value in store
func (s *KV) Set(ctx context.Context, k, v string) error {
	return s.File.WriteString(k, v)
}

// Get gets k and return v, or error
func (s *KV) Get(ctx context.Context, k string) string {
	return s.File.ReadString(k)
}
