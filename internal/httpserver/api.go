package httpserver

import (
	"context"
	"ip_service/internal/apiv1"
	"ip_service/pkg/model"
	"ip_service/pkg/rpsl"
)

// Apiv1 interface
type Apiv1 interface {
	Index(ctx context.Context) (*model.ReplyIPInformation, error)

	IPText(ctx context.Context) (string, error)
	IPJSON(ctx context.Context) (map[string]any, error)

	CityText(ctx context.Context) (string, error)
	CityJSON(ctx context.Context) (map[string]any, error)

	CountryText(ctx context.Context) (string, error)
	CountryJSON(ctx context.Context) (map[string]any, error)

	CountryISOText(ctx context.Context) (string, error)
	CountryISOJSON(ctx context.Context) (map[string]any, error)

	ASNText(ctx context.Context) (uint, error)
	ASNJSON(ctx context.Context) (map[string]any, error)

	CoordinatesText(ctx context.Context) (*model.Coordinates, error)
	CoordinatesJSON(ctx context.Context) (map[string]any, error)

	AllJSON(ctx context.Context) (*model.ReplyIPInformation, error)

	LookUpIP(ctx context.Context, indata *apiv1.LookUpIPRequest) (*model.ReplyLookUp, error)

	Collision(ctx context.Context, indata *apiv1.CollisionRequest) (*apiv1.CollisionReply, error)

	Whois(ctx context.Context, indata *apiv1.WhoisRequest) ([]rpsl.ASN, error)

	Status(ctx context.Context) (*model.StatusReply, error)
}
