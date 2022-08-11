package model

import (
	ua "github.com/mileusna/useragent"
)

type (
	ContextIP        string
	ContextUserAgent string
)

type ReplyIPInformation struct {
	IP              string       `json:"ip"`
	IPDecimal       string       `json:"ip_decimal"`
	ASN             uint         `json:"asn"`
	ASNOrganization string       `json:"asn_organization"`
	City            string       `json:"city"`
	Country         string       `json:"country"`
	CountryISO      string       `json:"country_iso"`
	IsEU            bool         `json:"is_eu"`
	Region          string       `json:"region"`
	RegionCode      string       `json:"region_code"`
	PostalCode      string       `json:"postal_code"`
	Latitude        float64      `json:"latitude"`
	Longitude       float64      `json:"longitude"`
	Timezone        string       `json:"timezone"`
	Hostname        string       `json:"hostname"`
	UserAgent       ua.UserAgent `json:"user_agent"`
	Continent       string       `json:"continent"`
}

type Maxmind struct {
	ASN  MaxmindInformation `json:"asn"`
	City MaxmindInformation `json:"city"`
}

type MaxmindInformation struct {
	Version string `json:"version"`
}

type ReplyInfo struct {
	MaxMind Maxmind `json:"maxmind"`
	Started string  `json:"started"`
}
