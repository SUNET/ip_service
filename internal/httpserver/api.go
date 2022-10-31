package httpserver

import (
	"context"
	"ip_service/internal/apiv1"
	"ip_service/pkg/model"
)

// Apiv1 interface
type Apiv1 interface {
	Index(ctx context.Context) (*model.ReplyIPInformation, error)

	IPText(ctx context.Context) (string, error)
	IPJSON(ctx context.Context) (map[string]interface{}, error)

	CityText(ctx context.Context) (string, error)
	CityJSON(ctx context.Context) (map[string]interface{}, error)

	CountryText(ctx context.Context) (string, error)
	CountryJSON(ctx context.Context) (map[string]interface{}, error)

	CountryISOText(ctx context.Context) (string, error)
	CountryISOJSON(ctx context.Context) (map[string]interface{}, error)

	ASNText(ctx context.Context) (uint, error)
	ASNJSON(ctx context.Context) (map[string]interface{}, error)

	CoordinatesJSON(ctx context.Context) (map[string]interface{}, error)
	CoordinatesText(ctx context.Context) ([]float64, error)

	AllJSON(ctx context.Context) (*model.ReplyIPInformation, error)

	LookUpIP(ctx context.Context, indata *apiv1.RequestLookUpIP) (*model.ReplyLookUp, error)

	Info(ctx context.Context) (*model.ReplyInfo, error)

	Status(ctx context.Context) (*model.StatusService, error)
}
