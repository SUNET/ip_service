package httpserver

import (
	"context"
	"ip_service/internal/apiv1"

	"github.com/gin-gonic/gin"
)

func (s *Service) endpointJSON(ctx context.Context, c *gin.Context) (interface{}, error) {
	ctx = context.WithValue(ctx, "ua", c.Request.UserAgent())
	ctx = context.WithValue(ctx, "ip", c.RemoteIP())

	reply, err := s.apiv1.JSON(ctx)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func (s *Service) endpointCity(ctx context.Context, c *gin.Context) (interface{}, error) {
	ctx = context.WithValue(ctx, "ip", c.RemoteIP())

	reply, err := s.apiv1.City(ctx)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func (s *Service) endpointASN(ctx context.Context, c *gin.Context) (interface{}, error) {
	ctx = context.WithValue(ctx, "ip", c.RemoteIP())

	reply, err := s.apiv1.ASN(ctx)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func (s *Service) endpointIP(ctx context.Context, c *gin.Context) (interface{}, error) {
	ctx = context.WithValue(ctx, "ip", c.RemoteIP())

	reply, err := s.apiv1.IP(ctx)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func (s *Service) endpointCountry(ctx context.Context, c *gin.Context) (interface{}, error) {
	ctx = context.WithValue(ctx, "ip", c.RemoteIP())

	reply, err := s.apiv1.Country(ctx)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func (s *Service) endpointCountryISO(ctx context.Context, c *gin.Context) (interface{}, error) {
	ctx = context.WithValue(ctx, "ip", c.RemoteIP())

	reply, err := s.apiv1.CountryISO(ctx)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func (s *Service) endpointLookUpIP(ctx context.Context, c *gin.Context) (interface{}, error) {
	request := &apiv1.RequestLookUpIP{}
	if err := s.bindRequest(c, request); err != nil {
		return nil, err
	}

	ctx = context.WithValue(ctx, "ua", c.Request.UserAgent())

	reply, err := s.apiv1.LookUpIP(ctx, request)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func (s *Service) endpointStatus(ctx context.Context, c *gin.Context) (interface{}, error) {
	reply, err := s.apiv1.Status(ctx)
	if err != nil {
		return nil, err
	}
	return reply, nil
}
