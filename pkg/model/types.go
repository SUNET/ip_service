package model

import (
	ua "github.com/mileusna/useragent"
)

type (
	ContextIP        string
	ContextUserAgent string
)

type RequestInformation struct {
	IP              string       `json:"ip"`
	IPDecimal       string       `json:"ip_decimal"`
	ASN             uint         `json:"asn"`
	ASNOrganization string       `json:"asn_organization"`
	City            string       `json:"city"`
	Country         string       `json:"country"`
	CountryISO      string       `json:"country_iso"`
	IsEU            bool         `json:"is_eu"`
	Region          string       `json:"region"`
	RegionCode      string       `json:"regionCode"`
	PostalCode      string       `json:"postal_code"`
	Latitude        float64      `json:"Latitude"`
	Longitude       float64      `json:"Longitude"`
	Timezone        string       `json:"timezone"`
	Hostname        string       `json:"hostname"`
	UserAgent       ua.UserAgent `json:"user_agent"`
	Continent       string       `json:"continent"`
}
