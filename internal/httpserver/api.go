package httpserver

import (
	"context"
	"ip_service/internal/apiv1"
	"ip_service/pkg/model"
)

// Apiv1 interface
type Apiv1 interface {
	IP(ctx context.Context) (string, error)
	City(ctx context.Context) (string, error)
	ASN(ctx context.Context) (uint, error)
	JSON(ctx context.Context) (*model.RequestInformation, error)
	Country(ctx context.Context) (string, error)
	CountryISO(ctx context.Context) (string, error)

	LookUpIP(ctx context.Context, indata *apiv1.RequestLookUpIP) (*model.RequestInformation, error)
	Status(ctx context.Context) (*model.Status, error)
}
