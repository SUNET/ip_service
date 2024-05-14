package httpserver

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// metrics is the metrics object for httpserver
type metrics struct {
	EndpointIndexHTMLCounter   prometheus.Counter
	EndpointIndexJSONCounter   prometheus.Counter
	EndpointIndexPLAINCounter  prometheus.Counter
	EndpointCityCounter        prometheus.Counter
	EndpointCountryCounter     prometheus.Counter
	EndpointCountryISOCounter  prometheus.Counter
	EndpointASNCounter         prometheus.Counter
	EndpointCoordinatesCounter prometheus.Counter
	EndpointAllCounter         prometheus.Counter
	EndpointLookUpIPCounter    prometheus.Counter
	HealthCounter              prometheus.Counter
}

func (m *metrics) init() {
	m.EndpointIndexHTMLCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "ip_service_http_endpoint_index_html_total",
		Help: "The total number of request to endpoint / with Accept: text/html",
	})
	m.EndpointIndexJSONCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "ip_service_http_endpoint_index_json_total",
		Help: "The total number of request to endpoint / with Accept: application/json",
	})
	m.EndpointIndexPLAINCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "ip_service_http_endpoint_index_plain_total",
		Help: "The total number of request to endpoint / with Accept: text/plain",
	})
	m.EndpointCityCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "ip_service_http_endpoint_city_total",
		Help: "The total number of request to endpoint /city",
	})
	m.EndpointCountryCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "ip_service_http_endpoint_country_total",
		Help: "The total number of request to endpoint /country",
	})
	m.EndpointCountryISOCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "ip_service_http_endpoint_country_iso_total",
		Help: "The total number of request to endpoint /country-iso",
	})
	m.EndpointASNCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "ip_service_http_endpoint_asn_total",
		Help: "The total number of request to endpoint /asn",
	})
	m.EndpointCoordinatesCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "ip_service_http_endpoint_coordinates_total",
		Help: "The total number of request to endpoint /coordinates",
	})
	m.EndpointAllCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "ip_service_http_endpoint_all_total",
		Help: "The total number of request to endpoint /all",
	})
	m.EndpointLookUpIPCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "ip_service_http_endpoint_lookup_ip_total",
		Help: "The total number of request to endpoint /lookup",
	})
	m.HealthCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "ip_service_http_health_total",
		Help: "The total number of request to endpoint /health",
	})
}
