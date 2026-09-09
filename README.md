# ip_service

## API & web

This service accepts these `Accept` headers:

* `application/json`
* `text/html`
* `text/plain`

A missing or unknown `Accept` header renders as `text/plain` so `curl ip.sunet.se`
conveniently returns the public IP terminated with `\n`.

### Endpoints

| Method | Path                     | Description                                                                 |
| ------ | ------------------------ | --------------------------------------------------------------------------- |
| GET    | `/`                      | Public IP (`text/plain`/`application/json`) or full info page (`text/html`) |
| GET    | `/city`                  | City for the client IP                                                      |
| GET    | `/country`               | Country name for the client IP                                              |
| GET    | `/country-iso`           | Country ISO code for the client IP                                          |
| GET    | `/asn`                   | Autonomous System Number for the client IP                                  |
| GET    | `/coordinates`           | Latitude/longitude for the client IP                                        |
| GET    | `/all`                   | All available attributes for the client IP (JSON)                           |
| GET    | `/lookup/{ip}`           | All available attributes for the supplied IP                                |
| GET    | `/whois/{ip}`            | RPSL/whois information for the supplied IP                                  |
| POST   | `/collision`             | Check if two CIDR blocks overlap                                            |
| GET    | `/health`                | Service health/status                                                       |
| GET    | `/metrics`               | Prometheus metrics                                                          |
| GET    | `/swagger/*`             | Swagger UI and OpenAPI spec                                                 |
| GET    | `/assets/*`              | Embedded static assets used by the HTML view                                |
| GET    | `/debug/pprof/`          | pprof index (non-production only)                                           |
| GET    | `/debug/pprof/heap`      | pprof heap profile (non-production only)                                    |
| GET    | `/debug/pprof/goroutine` | pprof goroutine profile (non-production only)                               |
| GET    | `/debug/pprof/allocs`    | pprof allocation profile (non-production only)                              |

#### `GET /`

* `text/plain` / `application/json`: returns only the public IP.
* `text/html`: renders a web page with all attributes.

#### `GET /city`, `/country`, `/country-iso`, `/asn`, `/coordinates`

Returns a single attribute for the client IP. `text/plain` returns the raw
value; `application/json` returns it wrapped in a JSON object (e.g.
`{"city": "Stockholm"}`).

#### `GET /all`

Returns every available attribute for the client IP as JSON.

#### `GET /lookup/{ip}`

Path parameter `ip` (IPv4 or IPv6). Returns every available attribute for the
supplied IP as JSON.

#### `GET /whois/{ip}`

Path parameter `ip` (IPv4 or IPv6). Returns RPSL/whois information for the
supplied IP as JSON.

#### `POST /collision`

JSON body with two CIDR blocks; response indicates whether they overlap.

Request:

```json
{
  "ip_1": "10.0.0.0/8",
  "ip_2": "10.1.0.0/16"
}
```

Response:

```json
{
  "collision": true
}
```

#### `GET /health`

Returns the service status, including probe results for internal dependencies.

#### `GET /metrics`

Prometheus scrape endpoint.

#### `GET /swagger/*`

Interactive Swagger UI. The generated OpenAPI documents are also available in
[docs/](docs/).

#### `GET /debug/pprof/*`

Standard Go `net/http/pprof` endpoints for `heap`, `goroutine` and `allocs`
profiles. Only registered when `ip_service.production` is `false`; in
production these paths return `404`.
