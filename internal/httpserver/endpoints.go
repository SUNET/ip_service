package httpserver

import (
	"context"
	"errors"

	"ip_service/internal/apiv1"
	"ip_service/pkg/contexthandler"

	"github.com/gin-gonic/gin"
)

func (s *Service) endpointIndex(ctx context.Context, c *gin.Context) (any, error) {
	ctx, span := s.TP.Start(ctx, "httpserver:endpointIndex")
	defer span.End()

	contextRequest, err := contexthandler.Get(ctx, "request")
	if err != nil {
		return nil, err
	}

	switch contextRequest.Accept {
	case gin.MIMEPlain, "*/*":
		s.logger.Debug("index, plain")
		reply, err := s.apiv1.IPText(ctx)
		if err != nil {
			return nil, err
		}
		return reply, nil

	case gin.MIMEJSON:
		s.logger.Debug("index, json")
		reply, err := s.apiv1.IPJSON(ctx)
		if err != nil {
			return nil, err
		}
		return reply, nil

	case gin.MIMEHTML:
		s.logger.Debug("index, html")
		reply, err := s.apiv1.Index(ctx)
		if err != nil {
			return nil, err
		}
		return reply, nil
	}

	return nil, errors.New("unsupported content type")
}

func (s *Service) endpointCity(ctx context.Context, c *gin.Context) (any, error) {
	ctx, span := s.TP.Start(ctx, "httpserver:endpointCity")
	defer span.End()

	contextRequest, err := contexthandler.Get(ctx, "request")
	if err != nil {
		return nil, err
	}

	switch contextRequest.Accept {
	case gin.MIMEJSON, gin.MIMEHTML:
		reply, err := s.apiv1.CityJSON(ctx)
		if err != nil {
			return nil, err
		}
		return reply, nil
	default:
		reply, err := s.apiv1.CityText(ctx)
		if err != nil {
			return nil, err
		}
		return reply, nil
	}
}

func (s *Service) endpointCountry(ctx context.Context, c *gin.Context) (any, error) {
	ctx, span := s.TP.Start(ctx, "httpserver:endpointCountry")
	defer span.End()

	contextRequest, err := contexthandler.Get(ctx, "request")
	if err != nil {
		return nil, err
	}

	switch contextRequest.Accept {
	case gin.MIMEJSON, gin.MIMEHTML:
		reply, err := s.apiv1.CountryJSON(ctx)
		if err != nil {
			return nil, err
		}
		return reply, nil
	default:
		reply, err := s.apiv1.CountryText(ctx)
		if err != nil {
			return nil, err
		}
		return reply, nil
	}
}

func (s *Service) endpointCountryISO(ctx context.Context, c *gin.Context) (any, error) {
	ctx, span := s.TP.Start(ctx, "httpserver:endpointCountryISO")
	defer span.End()

	contextRequest, err := contexthandler.Get(ctx, "request")
	if err != nil {
		return nil, err
	}

	switch contextRequest.Accept {
	case gin.MIMEJSON, gin.MIMEHTML:
		reply, err := s.apiv1.CountryISOJSON(ctx)
		if err != nil {
			return nil, err
		}
		return reply, nil
	default:
		reply, err := s.apiv1.CountryISOText(ctx)
		if err != nil {
			return nil, err
		}
		return reply, nil
	}
}

func (s *Service) endpointASN(ctx context.Context, c *gin.Context) (any, error) {
	ctx, span := s.TP.Start(ctx, "httpserver:endpointASN")
	defer span.End()

	contextRequest, err := contexthandler.Get(ctx, "request")
	if err != nil {
		return nil, err
	}

	switch contextRequest.Accept {
	case gin.MIMEJSON, gin.MIMEHTML:
		reply, err := s.apiv1.ASNJSON(ctx)
		if err != nil {
			return nil, err
		}
		return reply, nil
	default:
		reply, err := s.apiv1.ASNText(ctx)
		if err != nil {
			return nil, err
		}
		return reply, nil
	}
}

func (s *Service) endpointCoordinates(ctx context.Context, c *gin.Context) (any, error) {
	ctx, span := s.TP.Start(ctx, "httpserver:endpointCoordinates")
	defer span.End()

	contextRequest, err := contexthandler.Get(ctx, "request")
	if err != nil {
		return nil, err
	}

	switch contextRequest.Accept {
	case gin.MIMEJSON, gin.MIMEHTML:
		reply, err := s.apiv1.CoordinatesJSON(ctx)
		if err != nil {
			return nil, err
		}
		return reply, nil

	default:
		reply, err := s.apiv1.CoordinatesText(ctx)
		if err != nil {
			return nil, err
		}
		return reply, nil
	}
}

func (s *Service) endpointAll(ctx context.Context, c *gin.Context) (any, error) {
	ctx, span := s.TP.Start(ctx, "httpserver:endpointAll")
	defer span.End()

	reply, err := s.apiv1.AllJSON(ctx)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func (s *Service) endpointLookUpIP(ctx context.Context, c *gin.Context) (any, error) {
	ctx, span := s.TP.Start(ctx, "httpserver:endpointLookUpIP")
	defer span.End()

	request := &apiv1.LookUpIPRequest{}
	if err := s.bindRequest(ctx, c, request); err != nil {
		return nil, err
	}

	reply, err := s.apiv1.LookUpIP(ctx, request)
	if err != nil {
		return nil, err
	}
	return reply, nil
}

func (s *Service) endpointStatus(ctx context.Context, c *gin.Context) (any, error) {
	ctx, span := s.TP.Start(ctx, "httpserver:endpointStatus")
	defer span.End()

	reply, err := s.apiv1.Status(ctx)
	if err != nil {
		return nil, err
	}
	return reply, nil
}
