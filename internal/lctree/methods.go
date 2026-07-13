package lctree

import "context"

// Close is a no-op for the in-memory tree service.
func (s *Service) Close(ctx context.Context) error {
	return nil
}
