# ip_service

## API & web

This service accept these Accept headers;

* application/json
* text/html
* text/plain

default or non Accept header will render in text/plain since it's convenient to `curl host/` and get the public ip with \n.

### Endpoints

#### /

text/html: website with all attribute formatted in a html website.

application/json / text/plain: return just public ip.

#### /city

#### /country

#### /country-iso

#### /asn

#### /coordinates

#### /lookup/\<ip\>

#### /all

application/json / text/plain: return all attributes.

#### /health

#### /metrics
