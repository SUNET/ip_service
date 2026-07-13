package apiv1

import (
	"context"
	"encoding/json"
	"ip_service/pkg/contexthandler"
	"ip_service/pkg/model"
	"net/netip"
)

// Index handler return all information in JSON format
//
//	@Summary		get IP for the request
//	@ID				ip
//	@Description	ip for the request
//	@Tags			ip_service
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	model.ReplyIPInformation	"Success"
//	@Failure		400	{object}	helpers.ErrorResponse		"Bad Request"
//	@Router			/ [get]
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
//
//	@Summary		get city for the given IP
//	@ID				city
//	@Description	city for the given IP
//	@Tags			ip_service
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	map[string]any			"Success"
//	@Failure		400	{object}	helpers.ErrorResponse	"Bad Request"
//	@Router			/city [get]
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
//
//	@Summary		get country for the given IP
//	@ID				country
//	@Description	country for the given IP
//	@Tags			ip_service
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	map[string]any			"Success"
//	@Failure		400	{object}	helpers.ErrorResponse	"Bad Request"
//	@Router			/country [get]
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
//
//	@Summary		get ASN for the given IP
//	@ID				asn
//	@Description	ASN for the given IP
//	@Tags			ip_service
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	map[string]any			"Success"
//	@Failure		400	{object}	helpers.ErrorResponse	"Bad Request"
//	@Router			/asn [get]
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
//
//	@Summary		get Country ISO for the given IP
//	@ID				countryISO
//	@Description	country ISO for the given IP
//	@Tags			ip_service
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	map[string]any			"Success"
//	@Failure		400	{object}	helpers.ErrorResponse	"Bad Request"
//	@Router			/country-iso [get]
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
func (c *Client) CoordinatesText(ctx context.Context) (*model.Coordinates, error) {
	ctx, span := c.tp.Start(ctx, "apiv1:CoordinatesText")
	defer span.End()

	return c.coordinates(ctx)
}

// CoordinatesJSON handler return Coordinates in JSON format
//
//	@Summary		get Coordinates for the given IP
//	@ID				coordinates
//	@Description	coordinates for the given IP
//	@Tags			ip_service
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	map[string]any			"Success"
//	@Failure		400	{object}	helpers.ErrorResponse	"Bad Request"
//	@Router			/coordinates [get]
func (c *Client) CoordinatesJSON(ctx context.Context) (map[string]any, error) {
	ctx, span := c.tp.Start(ctx, "apiv1:CoordinatesJSON")
	defer span.End()

	coordinates, err := c.coordinates(ctx)
	if err != nil {
		return nil, err
	}

	b, err := json.Marshal(coordinates)
	if err != nil {
		return nil, err
	}

	reply := map[string]any{}
	if err := json.Unmarshal(b, &reply); err != nil {
		return nil, err
	}

	return reply, nil
}

// AllJSON handler return all information in JSON format
func (c *Client) AllJSON(ctx context.Context) (*model.ReplyIPInformation, error) {
	ctx, span := c.tp.Start(ctx, "apiv1:AllJSON")
	defer span.End()

	c.log.Debug("allJSON handler")

	reply, err := c.formatAllJSON(ctx)
	if err != nil {
		c.log.Error(err, "failed to format all JSON")
		return nil, err
	}

	c.log.Debug("allJSON done")

	return reply, nil
}

// LookUpIPRequest is the request for the LookUpIP handler
type LookUpIPRequest struct {
	IP string `uri:"ip" validate:"required,ip"`
}

// LookUpIP handler return all information in JSON format for the given IP
//
//	@Summary		Look up all information for the given IP
//	@ID				lookUpIP
//	@Description	takes query parameter ip and returns all information
//	@Tags			ip_service
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	model.ReplyLookUp		"Success"
//	@Failure		400	{object}	helpers.ErrorResponse	"Bad Request"
//	@Param			ip	path		string					true	"ip"
//	@Router			/lookup/{ip} [get]
func (c *Client) LookUpIP(ctx context.Context, indata *LookUpIPRequest) (*model.ReplyLookUp, error) {
	c.log.Info("LookUpIP")
	ctx, span := c.tp.Start(ctx, "apiv1:LookUpIP")
	defer span.End()

	// Update the IP in the context to the one provided in the request
	contextRequest, err := contexthandler.Get(ctx, "request")
	if err != nil {
		c.log.Error(err, "failed to get context request")
		return nil, err
	}

	contextRequest.ClientIP = indata.IP

	formatJSON, err := c.formatLookUpJSON(ctx)
	if err != nil {
		c.log.Error(err, "failed to format LookUp JSON")
		return nil, err
	}

	return formatJSON, nil
}

type CollisionRequest struct {
	IP1 string `json:"ip_1" validate:"required,cidr"`
	IP2 string `json:"ip_2" validate:"required,cidr"`
}

type CollisionReply struct {
	Collision bool `json:"collision"`
}

// Collision handler checks if two CIDR blocks overlap
//
//	@Summary		Check if two CIDR blocks overlap
//	@ID				collision
//	@Description	takes two CIDR blocks and checks if they overlap
//	@Tags			ip_service
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	CollisionReply			"Success"
//	@Failure		400	{object}	helpers.ErrorResponse	"Bad Request"
//	@Param			req	body		CollisionRequest		true	" "
//	@Router			/collision [post]
func (c *Client) Collision(ctx context.Context, indata *CollisionRequest) (*CollisionReply, error) {
	_, span := c.tp.Start(ctx, "apiv1:Collision")
	defer span.End()

	p1, err := netip.ParsePrefix(indata.IP1)
	if err != nil {
		c.log.Error(err, "failed to parse IP1", "ip1", indata.IP1)
		return nil, err
	}

	p2, err := netip.ParsePrefix(indata.IP2)
	if err != nil {
		c.log.Error(err, "failed to parse IP2", "ip2", indata.IP2)
		return nil, err
	}

	reply := &CollisionReply{}
	reply.Collision = p1.Overlaps(p2)

	return reply, nil
}

// Status return status
//
//	@Summary		Status of the service
//	@ID				status
//	@Description	returns the status of the service
//	@Tags			ip_service
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	model.StatusReply		"Success"
//	@Failure		400	{object}	helpers.ErrorResponse	"Bad Request"
//	@Router			/health [get]
func (c *Client) Status(ctx context.Context) (*model.StatusReply, error) {
	ctx, span := c.tp.Start(ctx, "apiv1:Status")
	defer span.End()

	probes := model.StatusProbes{}
	probes = append(probes, c.store.Status(ctx))
	//probes = append(probes, c.max.Status(ctx))

	status := probes.Check("ip_service")

	return status, nil
}
