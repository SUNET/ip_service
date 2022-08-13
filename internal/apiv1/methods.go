package apiv1

import (
	"context"
	"ip_service/pkg/model"
	"math/big"
	"net"

	ua "github.com/mileusna/useragent"
	"inet.af/netaddr"
)

func (c *Client) getIP(ctx context.Context) string {
	return ctx.Value("ip").(string)
}

func (c *Client) getUserAgent(ctx context.Context) string {
	return ctx.Value("ua").(string)
}

func (c *Client) ua(ctx context.Context) ua.UserAgent {
	return ua.Parse(c.getUserAgent(ctx))
}

func (c *Client) IPDecimal(ctx context.Context) (string, error) {
	IP := netaddr.MustParseIP(c.getIP(ctx))
	ipBinary, err := IP.MarshalBinary()
	if err != nil {
		return "", err
	}

	ipInt := big.NewInt(0)
	ipInt.SetBytes(ipBinary)
	return ipInt.String(), nil
}

func (c *Client) asn(ctx context.Context) (uint, error) {
	m, err := c.max.ASN(net.ParseIP(c.getIP(ctx)))
	if err != nil {
		return 0, err
	}
	return m.AutonomousSystemNumber, nil
}

func (c *Client) asnOrganization(ctx context.Context) (string, error) {
	m, err := c.max.ASN(net.ParseIP(c.getIP(ctx)))
	if err != nil {
		return "", nil
	}
	return m.AutonomousSystemOrganization, nil
}

func (c *Client) postal(ctx context.Context) (string, error) {
	m, err := c.max.City(net.ParseIP(c.getIP(ctx)))
	if err != nil {
		return "", nil
	}
	return m.Postal.Code, nil
}

func (c *Client) city(ctx context.Context) (string, error) {
	m, err := c.max.City(net.ParseIP(c.getIP(ctx)))
	if err != nil {
		return "", err
	}
	return m.City.Names["en"], nil
}

func (c *Client) coordinates(ctx context.Context) ([]float64, error) {
	m, err := c.max.City(net.ParseIP(c.getIP(ctx)))
	if err != nil {
		return nil, err
	}
	return []float64{m.Location.Latitude, m.Location.Longitude}, nil
}

func (c *Client) country(ctx context.Context) (string, error) {
	m, err := c.max.City(net.ParseIP(c.getIP(ctx)))
	if err != nil {
		return "", err
	}
	return m.Country.Names["en"], nil
}

func (c *Client) countryISO(ctx context.Context) (string, error) {
	m, err := c.max.City(net.ParseIP(c.getIP(ctx)))
	if err != nil {
		return "", nil
	}
	return m.Country.IsoCode, nil
}

func (c *Client) isEU(ctx context.Context) (bool, error) {
	m, err := c.max.City(net.ParseIP(c.getIP(ctx)))
	if err != nil {
		return false, nil
	}
	return m.Country.IsInEuropeanUnion, nil
}

func (c *Client) timezone(ctx context.Context) (string, error) {
	m, err := c.max.City(net.ParseIP(c.getIP(ctx)))
	if err != nil {
		return "", nil
	}
	return m.Location.TimeZone, nil
}

func (c *Client) continent(ctx context.Context) (string, error) {
	m, err := c.max.City(net.ParseIP(c.getIP(ctx)))
	if err != nil {
		return "", nil
	}
	return m.Continent.Names["en"], nil
}

func (c *Client) formatAllJSON(ctx context.Context) (*model.ReplyIPInformation, error) {
	ipDecimal, err := c.IPDecimal(ctx)
	if err != nil {
		return nil, err
	}
	asn, err := c.asn(ctx)
	if err != nil {
		return nil, err
	}
	asnOrg, err := c.asnOrganization(ctx)
	if err != nil {
		return nil, err
	}
	city, err := c.city(ctx)
	if err != nil {
		return nil, err
	}
	country, err := c.country(ctx)
	if err != nil {
		return nil, err
	}
	countryISO, err := c.countryISO(ctx)
	if err != nil {
		return nil, err
	}
	eu, err := c.isEU(ctx)
	if err != nil {
		return nil, err
	}
	postal, err := c.postal(ctx)
	if err != nil {
		return nil, err
	}
	coordinates, err := c.coordinates(ctx)
	if err != nil {
		return nil, err
	}
	tz, err := c.timezone(ctx)
	if err != nil {
		return nil, err
	}
	continent, err := c.continent(ctx)
	if err != nil {
		return nil, err
	}

	return &model.ReplyIPInformation{
		IP:              c.getIP(ctx),
		IPDecimal:       ipDecimal,
		ASN:             asn,
		ASNOrganization: asnOrg,
		City:            city,
		Country:         country,
		CountryISO:      countryISO,
		IsEU:            eu,
		Region:          "",
		RegionCode:      "",
		PostalCode:      postal,
		Latitude:        coordinates[0],
		Longitude:       coordinates[1],
		Timezone:        tz,
		Hostname:        "",
		UserAgent:       c.ua(ctx),
		Continent:       continent,
	}, nil
}

func (c *Client) region(ctx context.Context) (string, error) {
	//	m, err := c.max.City(net.IP(c.getRemoteIP(ctx)))
	//	if err != nil {
	//		return "", nil
	//	}
	return "", nil

}
