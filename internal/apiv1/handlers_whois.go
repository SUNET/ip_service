package apiv1

import (
	"context"
	"ip_service/pkg/rpsl"
	"net/netip"
)

type WhoisRequest struct {
	IP string `uri:"ip" validate:"required,ip"`
}

// Whois handler return whois information in JSON format
//
//	@Summary		Whois information in JSON format
//	@ID				whois
//	@Description	takes query parameter ip and returns whois information in JSON format
//	@Tags			ip_service
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	rpsl.ASN				"Success"
//	@Failure		400	{object}	helpers.ErrorResponse	"Bad Request"
//	@Param			ip	path		string					true	"ip"
//	@Router			/whois/{ip} [get]
func (c *Client) Whois(ctx context.Context, indata *WhoisRequest) ([]rpsl.ASN, error) {
	ctx, span := c.tp.Start(ctx, "apiv1:Whois")
	defer span.End()

	ip, err := netip.ParseAddr(indata.IP)
	if err != nil {
		c.log.Error(err, "failed to parse ip", "ip", indata.IP)
		return nil, err
	}

	reply, err := c.whois.QueryIPAll(ctx, ip)
	if err != nil {
		c.log.Error(err, "failed to get route info from whois", "ip", indata.IP)
		return nil, err
	}

	return reply, nil
}
