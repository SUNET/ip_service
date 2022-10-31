package apiv1

import (
	"context"
	"ip_service/pkg/model"
)

func (c *Client) Index(ctx context.Context) (*model.ReplyIPInformation, error) {
	return c.formatAllJSON(ctx)
}

// IPText handler
func (c *Client) IPText(ctx context.Context) (string, error) {
	return c.getIP(ctx), nil
}
func (c *Client) IPJSON(ctx context.Context) (map[string]interface{}, error) {
	return map[string]interface{}{
		"ip": c.getIP(ctx),
	}, nil
}

func (c *Client) CityText(ctx context.Context) (string, error) {
	return c.city(ctx)
}
func (c *Client) CityJSON(ctx context.Context) (map[string]interface{}, error) {
	city, err := c.city(ctx)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"city": city,
	}, nil
}

func (c *Client) CountryText(ctx context.Context) (string, error) {
	return c.country(ctx)
}
func (c *Client) CountryJSON(ctx context.Context) (map[string]interface{}, error) {
	country, err := c.country(ctx)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"country": country,
	}, nil
}

func (c *Client) ASNText(ctx context.Context) (uint, error) {
	return c.asn(ctx)
}
func (c *Client) ASNJSON(ctx context.Context) (map[string]interface{}, error) {
	asn, err := c.asn(ctx)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"asn": asn,
	}, nil
}

func (c *Client) CountryISOText(ctx context.Context) (string, error) {
	return c.countryISO(ctx)
}
func (c *Client) CountryISOJSON(ctx context.Context) (map[string]interface{}, error) {
	countryISO, err := c.countryISO(ctx)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"country_iso": countryISO,
	}, nil
}

func (c *Client) CoordinatesText(ctx context.Context) ([]float64, error) {
	return c.coordinates(ctx)
}
func (c *Client) CoordinatesJSON(ctx context.Context) (map[string]interface{}, error) {
	coordinates, err := c.coordinates(ctx)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"coordinates": map[string]interface{}{
			"lat":  coordinates[0],
			"long": coordinates[1],
		},
	}, nil

}

func (c *Client) AllJSON(ctx context.Context) (*model.ReplyIPInformation, error) {
	return c.formatAllJSON(ctx)
}

type RequestLookUpIP struct {
	IP string `uri:"ip" validate:"required"`
}

func (c *Client) LookUpIP(ctx context.Context, indata *RequestLookUpIP) (*model.ReplyLookUp, error) {
	ctx = context.WithValue(ctx, "ip", indata.IP)
	return c.formatLookUpJSON(ctx)
}

func (c *Client) Info(ctx context.Context) (*model.ReplyInfo, error) {
	replyInfo := &model.ReplyInfo{
		MaxMind: model.Maxmind{
			ASN:  model.MaxmindInformation{Version: c.store.KV.Get(ctx, "ASN")},
			City: model.MaxmindInformation{Version: c.store.KV.Get(ctx, "City")},
		},
		Started: c.store.KV.Get(ctx, "started"),
	}

	return replyInfo, nil
}

// Status return status
func (c *Client) Status(ctx context.Context) (*model.StatusService, error) {
	status := model.AllStatus{
		c.max.Status(ctx),
		c.store.Status(ctx),
	}

	status.Check()

	return status.Check(), nil
}
