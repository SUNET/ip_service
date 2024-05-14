package model

import (
	ua "github.com/mileusna/useragent"
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
	Is1918Network   bool         `json:"is_1918_network"`
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

type ReplyLookUp struct {
	IP              string  `json:"ip"`
	IPDecimal       string  `json:"ip_decimal"`
	ASN             uint    `json:"asn"`
	ASNOrganization string  `json:"asn_organization"`
	City            string  `json:"city"`
	Country         string  `json:"country"`
	CountryISO      string  `json:"country_iso"`
	IsEU            bool    `json:"is_eu"`
	Is1918Network   bool    `json:"is_1918_network"`
	Region          string  `json:"region"`
	RegionCode      string  `json:"region_code"`
	PostalCode      string  `json:"postal_code"`
	Latitude        float64 `json:"latitude"`
	Longitude       float64 `json:"longitude"`
	Timezone        string  `json:"timezone"`
	Hostname        string  `json:"hostname"`
	Continent       string  `json:"continent"`
}
