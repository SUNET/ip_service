package rpsl

import (
	"context"
)

type Client struct {
	currentRouteObject *Object
	currentKey         *string

	RouterClass RouterClass
}

func New(ctx context.Context) (*Client, error) {
	service := &Client{
		currentRouteObject: &Object{},
		RouterClass:        make(RouterClass),
	}

	return service, nil
}
