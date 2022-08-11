package apiv1

import (
	"context"
	"ip_service/pkg/model"
)

// IP handler
func (c *Client) IP(ctx context.Context) (string, error) {
	return c.getIP(ctx), nil
}
func (c *Client) IPJSON(ctx context.Context) (map[string]interface{}, error) {
	return map[string]interface{}{
		"ip": c.getIP(ctx),
	}, nil
}

func (c *Client) City(ctx context.Context) (string, error) {
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

func (c *Client) Country(ctx context.Context) (string, error) {
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

func (c *Client) ASN(ctx context.Context) (uint, error) {
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

func (c *Client) CountryISO(ctx context.Context) (string, error) {
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

// SchoolInfo return a list of schoolNames
func (c *Client) JSON(ctx context.Context) (*model.ReplyIPInformation, error) {
	return c.formatJSON(ctx)
}

type RequestLookUpIP struct {
	IP string `uri:"ip" validate:"required"`
}

func (c *Client) LookUpIP(ctx context.Context, indata *RequestLookUpIP) (*model.ReplyIPInformation, error) {
	ctx = context.WithValue(ctx, "ip", indata.IP)
	return c.formatJSON(ctx)
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
