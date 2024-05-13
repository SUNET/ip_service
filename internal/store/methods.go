package store

import (
	"context"
	"errors"
	"strings"
	"time"
)

// Set sets key value in store
func (s *KV) Set(ctx context.Context, k, v string) error {
	return s.File.WriteString(k, v)
}

func (s *KV) SetLastChecked(ctx context.Context, k string) error {
	return s.File.WriteString(strings.Join([]string{k, "last_checked"}, "/"), time.Now().String())
}

func (s *KV) SetPreviousVersion(ctx context.Context, k string) error {
	return s.File.WriteString(strings.Join([]string{k, "previous_version"}, "/"), time.Now().String())
}

func (s *KV) SetRemoteVersion(ctx context.Context, k, v string) error {
	return s.File.WriteString(strings.Join([]string{k, "remote_version"}, "/"), v)
}

// Get gets k and return v, or error
func (s *KV) Get(ctx context.Context, k string) string {
	return s.File.ReadString(k)
}
func (s *KV) GetRemoteVersion(ctx context.Context, k string) string {
	return s.File.ReadString(strings.Join([]string{k, "remote_version"}, "/"))
}
func (s *KV) GetLastChecked(ctx context.Context, k string) string {
	return s.File.ReadString(strings.Join([]string{k, "last_checked"}, "/"))
}

func (s *KV) Del(ctx context.Context, k string) error {
	return s.File.Erase(k)
}

func (s *KV) statusTest(ctx context.Context) error {
	if err := s.Set(ctx, "testK", "testV"); err != nil {
		return err
	}
	if get := s.Get(ctx, "testK"); get != "testV" {
		return errors.New("can't find key")
	}
	if err := s.Del(ctx, "testK"); err != nil {
		return err
	}
	return nil
}
