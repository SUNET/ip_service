package apiv1

import (
	"context"
	"ip_service/pkg/contexthandler"
	"ip_service/pkg/model"
)

// Index handler return all information in JSON format
func (c *Client) Index(ctx context.Context) (*model.ReplyIPInformation, error) {
	ctx, span := c.tp.Start(ctx, "apiv1:Index")
	defer span.End()

	return c.formatAllJSON(ctx)
}

// IPText handler return IP in text format
func (c *Client) IPText(ctx context.Context) (string, error) {
	ctx, span := c.tp.Start(ctx, "apiv1:IPText")
	defer span.End()

	return c.getIP(ctx)
}

// IPJSON handler return IP in JSON format
func (c *Client) IPJSON(ctx context.Context) (map[string]any, error) {
	ctx, span := c.tp.Start(ctx, "apiv1:IPJSON")
	defer span.End()

	ip, err := c.getIP(ctx)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"ip": ip,
	}, nil
}

// CityText handler return City in text format
func (c *Client) CityText(ctx context.Context) (string, error) {
	ctx, span := c.tp.Start(ctx, "apiv1:CityText")
	defer span.End()

	return c.city(ctx)
}

// CityJSON handler return City in JSON format
func (c *Client) CityJSON(ctx context.Context) (map[string]any, error) {
	ctx, span := c.tp.Start(ctx, "apiv1:CityJSON")
	defer span.End()

	city, err := c.city(ctx)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"city": city,
	}, nil
}

// CountryText handler return Country in text format
func (c *Client) CountryText(ctx context.Context) (string, error) {
	ctx, span := c.tp.Start(ctx, "apiv1:CountryText")
	defer span.End()

	return c.country(ctx)
}

// CountryJSON handler return Country in JSON format
func (c *Client) CountryJSON(ctx context.Context) (map[string]any, error) {
	ctx, span := c.tp.Start(ctx, "apiv1:CountryJSON")
	defer span.End()

	country, err := c.country(ctx)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"country": country,
	}, nil
}

// ASNText handler return ASN in text format
func (c *Client) ASNText(ctx context.Context) (uint, error) {
	ctx, span := c.tp.Start(ctx, "apiv1:ASNText")
	defer span.End()

	return c.asn(ctx)
}

// ASNJSON handler return ASN in JSON format
func (c *Client) ASNJSON(ctx context.Context) (map[string]any, error) {
	ctx, span := c.tp.Start(ctx, "apiv1:ASNJSON")
	defer span.End()

	asn, err := c.asn(ctx)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"asn": asn,
	}, nil
}

// CountryISOText handler return Country ISO in text format
func (c *Client) CountryISOText(ctx context.Context) (string, error) {
	ctx, span := c.tp.Start(ctx, "apiv1:CountryISOText")
	defer span.End()

	return c.countryISO(ctx)
}

// CountryISOJSON handler return Country ISO in JSON format
func (c *Client) CountryISOJSON(ctx context.Context) (map[string]any, error) {
	ctx, span := c.tp.Start(ctx, "apiv1:CountryISOJSON")
	defer span.End()

	countryISO, err := c.countryISO(ctx)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		"country_iso": countryISO,
	}, nil
}

// CoordinatesText handler return Coordinates in text format
func (c *Client) CoordinatesText(ctx context.Context) ([]float64, error) {
	ctx, span := c.tp.Start(ctx, "apiv1:CoordinatesText")
	defer span.End()

	return c.coordinates(ctx)
}

// CoordinatesJSON handler return Coordinates in JSON format
func (c *Client) CoordinatesJSON(ctx context.Context) (map[string]any, error) {
	ctx, span := c.tp.Start(ctx, "apiv1:CoordinatesJSON")
	defer span.End()

	coordinates, err := c.coordinates(ctx)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"coordinates": map[string]any{
			"lat":  coordinates[0],
			"long": coordinates[1],
		},
	}, nil

}

// AllJSON handler return all information in JSON format
func (c *Client) AllJSON(ctx context.Context) (*model.ReplyIPInformation, error) {
	ctx, span := c.tp.Start(ctx, "apiv1:AllJSON")
	defer span.End()

	return c.formatAllJSON(ctx)
}

// LookUpIPRequest is the request for the LookUpIP handler
type LookUpIPRequest struct {
	IP string `uri:"ip" validate:"required"`
}

// LookUpIP handler return all information in JSON format for the given IP
func (c *Client) LookUpIP(ctx context.Context, indata *LookUpIPRequest) (*model.ReplyLookUp, error) {
	ctx, span := c.tp.Start(ctx, "apiv1:LookUpIP")
	defer span.End()

	// Update the IP in the context to the one provided in the request
	contextRequest, err := contexthandler.Get(ctx, "request")
	if err != nil {
		return nil, err
	}

	contextRequest.ClientIP = indata.IP
	ctx = contexthandler.Add(ctx, "request", contextRequest)

	return c.formatLookUpJSON(ctx)
}

// Status return status
func (c *Client) Status(ctx context.Context) (*model.StatusReply, error) {
	ctx, span := c.tp.Start(ctx, "apiv1:Status")
	defer span.End()

	probes := model.StatusProbes{}
	probes = append(probes, c.store.Status(ctx))
	probes = append(probes, c.max.Status(ctx))

	status := probes.Check("ip_service")

	return status, nil
}
