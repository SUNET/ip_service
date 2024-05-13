package contexthandler

import (
	"context"
	"fmt"
)

// RequestContext holds the request context
type RequestContext struct {
	ClientIP  string
	LookupIP  string
	UserAgent string
	Accept    string
}

// Add adds a key value pair to the context
func Add(ctx context.Context, key any, value *RequestContext) context.Context {
	ctx = context.WithValue(ctx, key, value)
	return ctx
}

// Get gets a value from the context
func Get(ctx context.Context, key any) (*RequestContext, error) {
	value := ctx.Value(key)
	if value == nil {
		return nil, fmt.Errorf("no value found in context for key %s", key)
	}

	return value.(*RequestContext), nil
}
