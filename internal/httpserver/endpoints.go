package httpserver

import (
	"context"
	"ip_service/internal/apiv1"

	"github.com/gin-gonic/gin"
	"github.com/golang/gddo/httputil"
)

func (s *Service) negotiateContentType(ctx context.Context, c *gin.Context) string {
	return httputil.NegotiateContentType(
		c.Request,
		[]string{"*/*", gin.MIMEPlain, gin.MIMEJSON, gin.MIMEHTML},
		"*/*",
	)
}

func (s *Service) endpointIndex(ctx context.Context, c *gin.Context) (interface{}, error) {
	ctx = context.WithValue(ctx, "ua", c.Request.UserAgent())
	ctx = context.WithValue(ctx, "ip", c.ClientIP())

	switch s.negotiateContentType(ctx, c) {
	case gin.MIMEPlain, "*/*":
		reply, err := s.apiv1.IPText(ctx)
		if err != nil {
			return nil, err
		}
		return reply, nil
	case gin.MIMEJSON:
		reply, err := s.apiv1.IPJSON(ctx)
		if err != nil {
			return nil, err
		}
		return reply, nil
	case gin.MIMEHTML:
		reply, err := s.apiv1.Index(ctx)
		if err != nil {
			return nil, err
		}
		return reply, nil

	}
	return nil, nil
}

func (s *Service) endpointCity(ctx context.Context, c *gin.Context) (interface{}, error) {
	ctx = context.WithValue(ctx, "ip", c.ClientIP())

	switch s.negotiateContentType(ctx, c) {
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

func (s *Service) endpointCountry(ctx context.Context, c *gin.Context) (interface{}, error) {
	ctx = context.WithValue(ctx, "ip", c.ClientIP())

	switch s.negotiateContentType(ctx, c) {
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

func (s *Service) endpointCountryISO(ctx context.Context, c *gin.Context) (interface{}, error) {
	ctx = context.WithValue(ctx, "ip", c.ClientIP())

	switch s.negotiateContentType(ctx, c) {
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

func (s *Service) endpointASN(ctx context.Context, c *gin.Context) (interface{}, error) {
	ctx = context.WithValue(ctx, "ip", c.ClientIP())

	switch s.negotiateContentType(ctx, c) {
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

func (s *Service) endpointCoordinates(ctx context.Context, c *gin.Context) (interface{}, error) {
	ctx = context.WithValue(ctx, "ip", c.ClientIP())

	switch s.negotiateContentType(ctx, c) {
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

func (s *Service) endpointAll(ctx context.Context, c *gin.Context) (interface{}, error) {
	ctx = context.WithValue(ctx, "ua", c.Request.UserAgent())
	ctx = context.WithValue(ctx, "ip", c.ClientIP())

	reply, err := s.apiv1.AllJSON(ctx)
	if err != nil {
		return nil, err
	}
	return reply, nil
	//switch c.Request.Header.Get("Accept") {
	//case gin.MIMEJSON, model.WebBrowserAccept, gin.MIMEHTML:
	//default:
	//	reply, err := s.apiv1.AllText(ctx)
	//	if err != nil {
	//		return nil, err
	//	}
	//	return reply, nil
	//}
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

func (s *Service) endpointInfo(ctx context.Context, c *gin.Context) (interface{}, error) {
	reply, err := s.apiv1.Info(ctx)
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
