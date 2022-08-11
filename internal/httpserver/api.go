package httpserver

import (
	"context"
	"ip_service/internal/apiv1"
	"ip_service/pkg/model"
)

// Apiv1 interface
type Apiv1 interface {
	IP(ctx context.Context) (string, error)
	IPJSON(ctx context.Context) (map[string]interface{}, error)

	City(ctx context.Context) (string, error)
	CityJSON(ctx context.Context) (map[string]interface{}, error)

	Country(ctx context.Context) (string, error)
	CountryJSON(ctx context.Context) (map[string]interface{}, error)

	CountryISO(ctx context.Context) (string, error)
	CountryISOJSON(ctx context.Context) (map[string]interface{}, error)

	ASN(ctx context.Context) (uint, error)
	ASNJSON(ctx context.Context) (map[string]interface{}, error)

	JSON(ctx context.Context) (*model.ReplyIPInformation, error)

	LookUpIP(ctx context.Context, indata *apiv1.RequestLookUpIP) (*model.ReplyIPInformation, error)

	Info(ctx context.Context) (*model.ReplyInfo, error)

	Status(ctx context.Context) (*model.StatusService, error)
}
