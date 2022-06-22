package store

import (
	"context"
	"time"
)

// Set sets key value in store
func (s *KV) Set(ctx context.Context, k string, v interface{}) error {
	return s.Redis.Set(ctx, k, v, 0*time.Second).Err()
}

// Get gets k and return v, or error
func (s *KV) Get(ctx context.Context, k string) (string, error) {
	return s.Redis.Get(ctx, k).Result()
}
