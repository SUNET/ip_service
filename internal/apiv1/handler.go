package apiv1

import (
	"context"
	"ip_service/pkg/model"
)

// IP handler
func (c *Client) IP(ctx context.Context) (string, error) {
	return c.getIP(ctx), nil
}

func (c *Client) City(ctx context.Context) (string, error) {
	return c.city(ctx)
}

func (c *Client) Country(ctx context.Context) (string, error) {
	return c.country(ctx)
}

func (c *Client) ASN(ctx context.Context) (uint, error) {
	return c.asn(ctx)
}

func (c *Client) CountryISO(ctx context.Context) (string, error) {
	return c.countryISO(ctx)
}

// SchoolInfo return a list of schoolNames
func (c *Client) JSON(ctx context.Context) (*model.RequestInformation, error) {
	return c.formatJSON(ctx)
}

type RequestLookUpIP struct {
	IP string `uri:"ip" validate:"required"`
}

func (c *Client) LookUpIP(ctx context.Context, indata *RequestLookUpIP) (*model.RequestInformation, error) {
	ctx = context.WithValue(ctx, "ip", indata.IP)
	return c.formatJSON(ctx)
}

// Status return status for each ladok instance
func (c *Client) Status(ctx context.Context) (*model.Status, error) {
	//	manyStatus := model.ManyStatus{}
	//	manyStatus = append(manyStatus, redis)
	//	status := manyStatus.Check()
	//
	//	return status, nil
	return nil, nil
}
