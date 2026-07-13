package apiv1

import (
	"context"
	"ip_service/pkg/contexthandler"
	"ip_service/pkg/model"
	"math/big"
	"net"

	ua "github.com/mileusna/useragent"
	"inet.af/netaddr"
)

func (c *Client) getIP(ctx context.Context) (string, error) {
	requestContext, err := contexthandler.Get(ctx, "request")
	if err != nil {
		c.log.Error(err, "failed to get request context")
		return "", err
	}

	return requestContext.ClientIP, nil
}

func (c *Client) getUserAgent(ctx context.Context) (string, error) {
	requestContext, err := contexthandler.Get(ctx, "request")
	if err != nil {
		return "", err
	}

	return requestContext.UserAgent, nil
}

func (c *Client) ua(ctx context.Context) (ua.UserAgent, error) {
	userAgent, err := c.getUserAgent(ctx)
	if err != nil {
		return ua.UserAgent{}, err
	}
	parsedUserAgent := ua.Parse(userAgent)

	return parsedUserAgent, nil
}

// IPDecimal returns the IP address in decimal format
func (c *Client) IPDecimal(ctx context.Context) (string, error) {
	ip, err := c.getIP(ctx)
	if err != nil {
		c.log.Error(err, "failed to get IP")
		return "", err
	}
	ipBinary, err := netaddr.MustParseIP(ip).MarshalBinary()
	if err != nil {
		c.log.Error(err, "failed to parse IP to binary")
		return "", err
	}

	ipInt := big.NewInt(0)
	ipInt.SetBytes(ipBinary)

	return ipInt.String(), nil
}

func (c *Client) asn(ctx context.Context) (uint, error) {
	ip, err := c.getIP(ctx)
	if err != nil {
		c.log.Error(err, "failed to get ip")
		return 0, err
	}

	m, err := c.max.ASN(ctx, net.ParseIP(ip))
	if err != nil {
		c.log.Error(err, "failed to get ASN")
		return 0, err
	}
	return m.AutonomousSystemNumber, nil
}

func (c *Client) asnOrganization(ctx context.Context) (string, error) {
	ip, err := c.getIP(ctx)
	if err != nil {
		return "", err
	}
	m, err := c.max.ASN(ctx, net.ParseIP(ip))
	if err != nil {
		c.log.Error(err, "failed to get ASN organization")
		return "", err
	}
	return m.AutonomousSystemOrganization, nil
}

func (c *Client) postal(ctx context.Context) (string, error) {
	ip, err := c.getIP(ctx)
	if err != nil {
		c.log.Error(err, "failed to get IP")
		return "", err
	}
	m, err := c.max.City(ctx, net.ParseIP(ip))
	if err != nil {
		c.log.Error(err, "failed to get City")
		return "", err
	}
	return m.Postal.Code, nil
}

func (c *Client) city(ctx context.Context) (string, error) {
	ip, err := c.getIP(ctx)
	if err != nil {
		c.log.Error(err, "failed to get IP")
		return "", err
	}
	m, err := c.max.City(ctx, net.ParseIP(ip))
	if err != nil {
		c.log.Error(err, "failed to get City")
		return "", err
	}

	reply, ok := m.City.Names["en"]
	if !ok {
		c.log.Debug("no City name in English available")
		return "", nil
	}
	return reply, nil
}

func (c *Client) coordinates(ctx context.Context) (*model.Coordinates, error) {
	ip, err := c.getIP(ctx)
	if err != nil {
		return nil, err
	}
	m, err := c.max.City(ctx, net.ParseIP(ip))
	if err != nil {
		return nil, err
	}
	reply := &model.Coordinates{
		Latitude:  m.Location.Latitude,
		Longitude: m.Location.Longitude,
	}
	return reply, nil
}

func (c *Client) country(ctx context.Context) (string, error) {
	ip, err := c.getIP(ctx)
	if err != nil {
		return "", err
	}
	m, err := c.max.City(ctx, net.ParseIP(ip))
	if err != nil {
		return "", err
	}

	return m.Country.Names["en"], nil
}

func (c *Client) countryISO(ctx context.Context) (string, error) {
	ip, err := c.getIP(ctx)
	if err != nil {
		c.log.Error(err, "failed to get IP")
		return "", err
	}
	m, err := c.max.City(ctx, net.ParseIP(ip))
	if err != nil {
		c.log.Error(err, "failed to get City for country ISO")
		return "", err
	}
	return m.Country.IsoCode, nil
}

func (c *Client) isEU(ctx context.Context) (bool, error) {
	ip, err := c.getIP(ctx)
	if err != nil {
		return false, err
	}
	m, err := c.max.City(ctx, net.ParseIP(ip))
	if err != nil {
		c.log.Error(err, "failed to get City for EU check")
		return false, err
	}
	return m.Country.IsInEuropeanUnion, nil
}

func (c *Client) is1918Network(ctx context.Context) (bool, error) {
	ip, err := c.getIP(ctx)
	if err != nil {
		return false, err
	}
	return net.ParseIP(ip).IsPrivate(), nil
}

func (c *Client) timezone(ctx context.Context) (string, error) {
	ip, err := c.getIP(ctx)
	if err != nil {
		return "", err
	}
	m, err := c.max.City(ctx, net.ParseIP(ip))
	if err != nil {
		c.log.Error(err, "failed to get City for timezone")
		return "", err
	}
	return m.Location.TimeZone, nil
}

func (c *Client) continent(ctx context.Context) (string, error) {
	ip, err := c.getIP(ctx)
	if err != nil {
		return "", err
	}
	m, err := c.max.City(ctx, net.ParseIP(ip))
	if err != nil {
		c.log.Error(err, "failed to get City for continent")
		return "", err
	}
	return m.Continent.Names["en"], nil
}

func (c *Client) formatAllJSON(ctx context.Context) (*model.ReplyIPInformation, error) {
	c.log.Debug("formatAllJSON start")

	reply := &model.ReplyIPInformation{}

	ip, err := c.getIP(ctx)
	if err != nil {
		c.log.Error(err, "failed to get IP")
		return nil, err
	}
	reply.IP = ip

	reply.IPDecimal, err = c.IPDecimal(ctx)
	if err != nil {
		c.log.Error(err, "failed to get IPDecimal")
		return nil, err
	}

	parsedIP := net.ParseIP(ip)

	// Single ASN lookup instead of two separate calls
	asnRecord, err := c.max.ASN(ctx, parsedIP)
	if err != nil {
		c.log.Error(err, "failed to get ASN")
		return nil, err
	}
	reply.ASN = asnRecord.AutonomousSystemNumber
	reply.ASNOrganization = asnRecord.AutonomousSystemOrganization

	// Single City lookup instead of eight separate calls
	cityRecord, err := c.max.City(ctx, parsedIP)
	if err != nil {
		c.log.Error(err, "failed to get City")
		return nil, err
	}

	reply.City = cityRecord.City.Names["en"]
	reply.Country = cityRecord.Country.Names["en"]
	reply.CountryISO = cityRecord.Country.IsoCode
	reply.IsEU = cityRecord.Country.IsInEuropeanUnion
	reply.Is1918Network = parsedIP.IsPrivate()
	reply.PostalCode = cityRecord.Postal.Code
	reply.Coordinates = &model.Coordinates{
		Latitude:  cityRecord.Location.Latitude,
		Longitude: cityRecord.Location.Longitude,
	}
	reply.Timezone = cityRecord.Location.TimeZone
	reply.Continent = cityRecord.Continent.Names["en"]

	reply.UserAgent, err = c.ua(ctx)
	if err != nil {
		c.log.Error(err, "failed to get UserAgent")
		return nil, err
	}

	return reply, nil
}

func (c *Client) formatLookUpJSON(ctx context.Context) (*model.ReplyLookUp, error) {
	reply := &model.ReplyLookUp{}

	ip, err := c.getIP(ctx)
	if err != nil {
		c.log.Error(err, "failed to get IP")
		return nil, err
	}
	reply.IP = ip

	reply.IPDecimal, err = c.IPDecimal(ctx)
	if err != nil {
		c.log.Error(err, "failed to get IPDecimal")
		return nil, err
	}

	parsedIP := net.ParseIP(ip)

	// Single ASN lookup instead of two separate calls
	asnRecord, err := c.max.ASN(ctx, parsedIP)
	if err != nil {
		c.log.Error(err, "failed to get ASN")
		return nil, err
	}
	reply.ASN = asnRecord.AutonomousSystemNumber
	reply.ASNOrganization = asnRecord.AutonomousSystemOrganization

	c.log.Debug("before whois")

	reply.Whois, err = c.whois.QueryIP(ctx, reply.IP)
	if err != nil {
		c.log.Error(err, "failed to get route info from radb", "ip", reply.IP)
		return nil, err
	}

	c.log.Debug("after whois")

	// Single City lookup instead of eight separate calls
	cityRecord, err := c.max.City(ctx, parsedIP)
	if err != nil {
		c.log.Error(err, "failed to get City")
		return nil, err
	}

	reply.City = cityRecord.City.Names["en"]
	reply.Country = cityRecord.Country.Names["en"]
	reply.CountryISO = cityRecord.Country.IsoCode
	reply.IsEU = cityRecord.Country.IsInEuropeanUnion
	reply.Is1918Network = parsedIP.IsPrivate()
	reply.PostalCode = cityRecord.Postal.Code
	reply.Coordinates = &model.Coordinates{
		Latitude:  cityRecord.Location.Latitude,
		Longitude: cityRecord.Location.Longitude,
	}
	reply.Timezone = cityRecord.Location.TimeZone
	reply.Continent = cityRecord.Continent.Names["en"]

	// Reverse DNS lookup
	names, err := net.DefaultResolver.LookupAddr(ctx, ip)
	if err == nil && len(names) > 0 {
		reply.PTR = names
	}

	return reply, nil
}
