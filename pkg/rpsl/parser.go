package rpsl

import (
	"context"
	"fmt"
	"time"

	"github.com/3th1nk/cidr"
)

const (
	Abuse        = "abuse"
	AbuseC       = "abuse-c"
	AbuseMailbox = "abuse-mailbox:"
	Address      = "address"
	AdminC       = "admin-c"
	AggrBNDRY    = "aggr-bndry"
	AggrMTD      = "aggr-mtd"
	Alias        = "alias"
	ASName       = "as-name"
	ASSet        = "as-set"
	Auth         = "auth"
	AuthNum      = "aut-num"
	Certif       = "certif"
	Changed      = "changed"
	Components   = "components"
	Correo       = "correo"
	Country      = "country"
	Created      = "created"
	Dba          = "dba"
	Default      = "default"
	Descr        = "descr"
	EMail        = "e-mail"
	Email        = "email"
	Export       = "export"
	ExportComps  = "export-comps"
	ExportVia    = "export-via"
	FaxNo        = "fax-no"
	Filter       = "filter"
	FilterSet    = "filter-set"
	FingerPR     = "fingerpr"
	GeoIDX       = "geoidx"
	Holes        = "holes"
	IFAddr       = "ifaddr"
	Import       = "import"
	ImportVia    = "import-via"
	Inet6num     = "inet6num"
	InetNum      = "inetnum"
	InetRTR      = "inet-rtr"
	Interface    = "interface"
	Inject       = "inject"
	KeyCert      = "key-cert"
	LastModified = "last-modified"
	LocalAS      = "local-as"
	MailTo       = "mailto"
	MBRSByRef    = "mbrs-by-ref"
	MemberOf     = "member-of"
	Members      = "members"
	Method       = "method"
	MNTBy        = "mnt-by"
	MNTLower     = "mnt-lower"
	Mntner       = "mntner"
	MNTNFY       = "mnt-nfy"
	MNTRoutes    = "mnt-routes"
	MPDefault    = "mp-default"
	MPExport     = "mp-export"
	MPFilter     = "mp-filter"
	MPImport     = "mp-import"
	MPMember     = "mp-members"
	MPPeer       = "mp-peer"
	MPPeering    = "mp-peering"
	Netname      = "netname"
	NICHDL       = "nic-hdl"
	Notify       = "notify"
	ORG          = "org"
	ORGName      = "org-name"
	ORGType      = "org-type"
	Origin       = "origin"
	Owner        = "owner"
	OwnerC       = "owner-c"
	OwnerID      = "ownerid"
	Peer         = "peer"
	Peering      = "peering"
	PeeringSet   = "peering-set"
	Person       = "person"
	Phone        = "phone"
	Pingable     = "pingable"
	pingHDL      = "ping-hdl"
	Remarks      = "remarks"
	Responsible  = "responsible"
	ROAURI       = "roa-uri"
	Role         = "role"
	Route        = "route"
	Route6       = "route6"
	RouteSet     = "route-set"
	RSIn         = "rs-in"
	RSOut        = "rs-out"
	RTRSet       = "rtr-set"
	Source       = "source"
	Status       = "status"
	Support      = "support"
	TechC        = "tech-c"
	Tel          = "tel."
	Trouble      = "trouble"
	UPDTo        = "upd-to"
	Website      = "website"
)

type ASN map[string]*Object

type RouterClass map[string]ASN

func (r RouterClass) removeBlankRecords(ctx context.Context) {
	_, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	for network := range r {
		if network == "" {
			delete(r, network)
		}
	}
}

// Object represents a parsed RPSL route object with a minimal set of fields
// needed for the API response. Keeping this small is critical — ~1.5M objects in memory.
type Object struct {
	Network      string   `json:"network,omitempty"`
	Origin       string   `json:"origin,omitempty"`
	Country      []string `json:"country,omitempty"`
	Remarks      []string `json:"remarks,omitempty"`
	Created      []string `json:"created,omitempty"`
	LastModified string   `json:"last-modified,omitempty"`
	Owner        string   `json:"owner,omitempty"`
	ORGName      string   `json:"org-name,omitempty"`
	ORG          string   `json:"org,omitempty"`
	OwnerID      string   `json:"ownerid,omitempty"`
}

func (no *Object) FindNetwork(ctx context.Context, ip string) (bool, error) {
	n, err := cidr.Parse(no.Network)
	if err != nil {
		return false, err
	}

	if n.Contains(ip) {
		return true, nil
	}

	return false, nil
}

func (r *Object) Add(key, value string) error {
	switch key {
	case Route, Route6:
		if value == "" {
			return fmt.Errorf("route/route6 value is empty")
		}
		r.Network = value
	case Origin:
		r.Origin = value
	case Country:
		r.Country = append(r.Country, value)
	case Remarks:
		r.Remarks = append(r.Remarks, value)
	case Created:
		r.Created = append(r.Created, value)
	case LastModified:
		r.LastModified = value
	case Owner:
		r.Owner = value
	case ORGName:
		r.ORGName = value
	case ORG:
		r.ORG = value
	case OwnerID:
		r.OwnerID = value
	}
	return nil
}
