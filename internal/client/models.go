package client

import "time"

// ListResponse is a generic list response wrapper
type ListResponse[T any] struct {
	Total int `json:"total"`
	Items []T `json:"items"`
}

// WAAPConfig represents a configuration metadata in the WAAP API
type WAAPConfig struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	Version     string    `json:"version,omitempty"`
	Date        time.Time `json:"date,omitempty"`
}

// ResponseMessage represents a success response
type ResponseMessage struct {
	Message string `json:"message"`
}

// ServerGroup represents a server group (site) in the API
type ServerGroup struct {
	ID                     string   `json:"id"`
	Name                   string   `json:"name"`
	Description            string   `json:"description,omitempty"`
	ServerNames            []string `json:"server_names"`
	SecurityPolicy         string   `json:"security_policy"`
	RoutingProfile         string   `json:"routing_profile"`
	ProxyTemplate          string   `json:"proxy_template"`
	ChallengeCookieDomain  string   `json:"challenge_cookie_domain"`
	SSLCertificate         string   `json:"ssl_certificate,omitempty"`
	ClientCertificate      string   `json:"client_certificate,omitempty"`
	ClientCertificateMode  string   `json:"client_certificate_mode,omitempty"`
	MobileApplicationGroup string   `json:"mobile_application_group,omitempty"`
}

// ServerGroupCreateRequest is the request body for creating a server group
type ServerGroupCreateRequest struct {
	ID                     string   `json:"id"`
	Name                   string   `json:"name"`
	Description            string   `json:"description"`
	ServerNames            []string `json:"server_names"`
	SecurityPolicy         string   `json:"security_policy"`
	RoutingProfile         string   `json:"routing_profile"`
	ProxyTemplate          string   `json:"proxy_template"`
	ChallengeCookieDomain  string   `json:"challenge_cookie_domain"`
	SSLCertificate         string   `json:"ssl_certificate,omitempty"`
	ClientCertificate      string   `json:"client_certificate,omitempty"`
	ClientCertificateMode  string   `json:"client_certificate_mode,omitempty"`
	MobileApplicationGroup string   `json:"mobile_application_group,omitempty"`
}

// ProviderLink represents a cloud provider link for a certificate
type ProviderLink struct {
	Provider string `json:"provider"`
	Link     string `json:"link"`
	Region   string `json:"region"`
}

// Certificate represents a certificate in the API
type Certificate struct {
	ID            string         `json:"id"`
	Name          string         `json:"name,omitempty"`
	CertBody      string         `json:"cert_body,omitempty"`
	PrivateKey    string         `json:"private_key,omitempty"`
	Subject       string         `json:"subject,omitempty"`
	Issuer        string         `json:"issuer,omitempty"`
	SAN           []string       `json:"san,omitempty"`
	Expires       string         `json:"expires,omitempty"`
	Uploaded      string         `json:"uploaded,omitempty"`
	LEAutoRenew   bool           `json:"le_auto_renew"`
	LEAutoReplace bool           `json:"le_auto_replace"`
	LEHash        string         `json:"le_hash,omitempty"`
	Revoked       bool           `json:"revoked"`
	CRL           []string       `json:"crl,omitempty"`
	CDP           []string       `json:"cdp,omitempty"`
	Side          string         `json:"side,omitempty"`
	Links         []ProviderLink `json:"links,omitempty"`
	ProviderLinks []ProviderLink `json:"provider_links,omitempty"`
}

// CertificateCreateRequest is the request body for creating a certificate
type CertificateCreateRequest struct {
	ID            string         `json:"id"`
	CertBody      string         `json:"cert_body,omitempty"`
	PrivateKey    string         `json:"private_key,omitempty"`
	LEAutoRenew   bool           `json:"le_auto_renew"`
	LEAutoReplace bool           `json:"le_auto_replace"`
	LEHash        string         `json:"le_hash,omitempty"`
	Side          string         `json:"side,omitempty"`
	ProviderLinks []ProviderLink `json:"provider_links,omitempty"`
}

// LoadBalancer represents a load balancer in the API (read-only)
type LoadBalancer struct {
	Name               string   `json:"name"`
	Provider           string   `json:"provider"`
	Region             string   `json:"region"`
	DNSName            string   `json:"dns_name"`
	ListenerName       string   `json:"listener_name"`
	ListenerPort       int      `json:"listener_port"`
	LoadBalancerType   string   `json:"load_balancer_type"`
	MaxCertificates    int      `json:"max_certificates"`
	DefaultCertificate string   `json:"default_certificate"`
	Certificates       []string `json:"certificates"`
}

// LoadBalancerRegion represents a load balancer's region configuration
type LoadBalancerRegion struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Regions         map[string]string `json:"regions"`
	UpstreamRegions []string          `json:"upstream_regions"`
}

// LoadBalancerRegions represents the regions response
type LoadBalancerRegions struct {
	CityCodes map[string]string    `json:"city_codes"`
	LBs       []LoadBalancerRegion `json:"lbs"`
}

// LoadBalancerRegionsUpdateRequest is the request body for updating regions
type LoadBalancerRegionsUpdateRequest struct {
	LBs []LoadBalancerRegionUpdate `json:"lbs"`
}

// LoadBalancerRegionUpdate represents a single load balancer region update
type LoadBalancerRegionUpdate struct {
	ID      string            `json:"id"`
	Regions map[string]string `json:"regions"`
}

// AttachCertificateOptions contains options for attaching a certificate
type AttachCertificateOptions struct {
	Provider     string
	Region       string
	Listener     string
	ListenerPort int
	IsDefault    bool
	ELBv2        bool
}

// DetachCertificateOptions contains options for detaching a certificate
type DetachCertificateOptions struct {
	Provider      string
	Region        string
	CertificateID string
	Listener      string
	ListenerPort  string
	ELBv2         bool
}

// PublishBucket represents a publish target bucket
type PublishBucket struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// PublishRequest is the request body for publishing a configuration
type PublishRequest []PublishBucket

// SecurityPolicy represents a security policy in the API
type SecurityPolicy struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Tags        []string        `json:"tags,omitempty"`
	Map         []SecProfileMap `json:"map,omitempty"`
	Session     interface{}     `json:"session,omitempty"`
	SessionIDs  interface{}     `json:"session_ids,omitempty"`
}

// SecProfileMap represents a security profile mapping entry
type SecProfileMap struct {
	ID                         string   `json:"id"`
	Name                       string   `json:"name"`
	Match                      string   `json:"match"`
	ACLProfile                 string   `json:"acl_profile"`
	ACLProfileActive           bool     `json:"acl_profile_active"`
	ContentFilterProfile       string   `json:"content_filter_profile"`
	ContentFilterProfileActive bool     `json:"content_filter_profile_active"`
	BackendService             string   `json:"backend_service"`
	Description                string   `json:"description,omitempty"`
	RateLimitRules             []string `json:"rate_limit_rules,omitempty"`
	EdgeFunctions              []string `json:"edge_functions,omitempty"`
}

// ACLProfile represents an ACL profile in the API
type ACLProfile struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	Action      string   `json:"action,omitempty"`
	Allow       []string `json:"allow,omitempty"`
	AllowBot    []string `json:"allow_bot,omitempty"`
	Deny        []string `json:"deny,omitempty"`
	DenyBot     []string `json:"deny_bot,omitempty"`
	ForceDeny   []string `json:"force_deny,omitempty"`
	Passthrough []string `json:"passthrough,omitempty"`
}

// ProxyTemplate represents a proxy template in the API
type ProxyTemplate struct {
	ID                            string                        `json:"id,omitempty"`
	Name                          string                        `json:"name"`
	Description                   string                        `json:"description,omitempty"`
	ACAOHeader                    bool                          `json:"acao_header"`
	XFFHeaderName                 string                        `json:"xff_header_name"`
	XRealIPHeaderName             string                        `json:"xrealip_header_name"`
	ProxyConnectTimeout           string                        `json:"proxy_connect_timeout"`
	ProxyReadTimeout              string                        `json:"proxy_read_timeout"`
	ProxySendTimeout              string                        `json:"proxy_send_timeout"`
	UpstreamHost                  string                        `json:"upstream_host"`
	ClientBodyTimeout             string                        `json:"client_body_timeout"`
	ClientBodyBufferSize          string                        `json:"client_body_buffer_size"`
	ClientHeaderTimeout           string                        `json:"client_header_timeout"`
	ClientHeaderBufferSize        string                        `json:"client_header_buffer_size"`
	ClientMaxBodySize             string                        `json:"client_max_body_size"`
	KeepaliveTimeout              string                        `json:"keepalive_timeout"`
	SendTimeout                   string                        `json:"send_timeout"`
	LimitReqRate                  string                        `json:"limit_req_rate"`
	LimitReqBurst                 string                        `json:"limit_req_burst"`
	MaskHeaders                   string                        `json:"mask_headers"`
	CustomListener                bool                          `json:"custom_listener"`
	LargeClientHeaderBuffersCount string                        `json:"large_client_header_buffers_count"`
	LargeClientHeaderBuffersSize  string                        `json:"large_client_header_buffers_size"`
	ConfSpecific                  string                        `json:"conf_specific,omitempty"`
	SSLConfSpecific               string                        `json:"ssl_conf_specific,omitempty"`
	SSLCiphers                    string                        `json:"ssl_ciphers,omitempty"`
	SSLProtocols                  []string                      `json:"ssl_protocols,omitempty"`
	AdvancedConfiguration         []ProxyTemplateAdvancedConfig `json:"advanced_configuration,omitempty"`
}

// ProxyTemplateAdvancedConfig represents an advanced nginx configuration block
type ProxyTemplateAdvancedConfig struct {
	Name          string   `json:"name"`
	Protocol      []string `json:"protocol"`
	Configuration string   `json:"configuration"`
	Description   string   `json:"description,omitempty"`
}

// RateLimitRule represents a rate limit rule in the API
type RateLimitRule struct {
	ID            string             `json:"id"`
	Name          string             `json:"name"`
	Description   string             `json:"description,omitempty"`
	Global        bool               `json:"global"`
	Active        bool               `json:"active"`
	Timeframe     int                `json:"timeframe"`
	Threshold     int                `json:"threshold"`
	TTL           int                `json:"ttl,omitempty"`
	Action        string             `json:"action"`
	IsActionBan   bool               `json:"is_action_ban,omitempty"`
	Tags          []string           `json:"tags,omitempty"`
	Include       RateLimitTagFilter `json:"include"`
	Exclude       RateLimitTagFilter `json:"exclude"`
	Key           interface{}        `json:"key,omitempty"`
	Pairwith      interface{}        `json:"pairwith,omitempty"`
	LastActivated int                `json:"last_activated,omitempty"`
}

// RateLimitTagFilter represents the include/exclude tag filter
type RateLimitTagFilter struct {
	Relation string   `json:"relation"`
	Tags     []string `json:"tags"`
}

// User represents a user account in the API
type User struct {
	ID          string `json:"id"`
	ACL         int    `json:"acl"`
	ContactName string `json:"contact_name"`
	Email       string `json:"email"`
	Mobile      string `json:"mobile"`
	OrgID       string `json:"org_id"`
	OrgName     string `json:"org_name,omitempty"`
	OTPSeed     string `json:"otpseed,omitempty"`
}

// UserCreateRequest is the request body for creating a user
type UserCreateRequest struct {
	ACL         int    `json:"acl"`
	ContactName string `json:"contact_name"`
	Email       string `json:"email"`
	Mobile      string `json:"mobile"`
	OrgID       string `json:"org_id"`
}

// UserCreateResponse is the response from creating a user
type UserCreateResponse struct {
	ID string `json:"id"`
}

// UserUpdateRequest is the request body for updating a user
type UserUpdateRequest struct {
	ACL         int    `json:"acl"`
	ContactName string `json:"contact_name"`
	Mobile      string `json:"mobile"`
}

// UserOrganization represents an organization with its users (list response)
type UserOrganization struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Users []User `json:"users"`
}

// BackendHost represents a backend host entry in a backend service
type BackendHost struct {
	Host         string `json:"host"`
	HTTPPorts    []int  `json:"http_ports"`
	HTTPSPorts   []int  `json:"https_ports"`
	Weight       int    `json:"weight"`
	MaxFails     int    `json:"max_fails"`
	FailTimeout  int    `json:"fail_timeout"`
	Down         bool   `json:"down"`
	MonitorState string `json:"monitor_state"`
	Backup       bool   `json:"backup"`
}

// BackendService represents a backend service in the API
type BackendService struct {
	ID                     string        `json:"id"`
	Name                   string        `json:"name"`
	Description            string        `json:"description,omitempty"`
	HTTP11                 bool          `json:"http11"`
	TransportMode          string        `json:"transport_mode"`
	Sticky                 string        `json:"sticky"`
	StickyCookieName       string        `json:"sticky_cookie_name,omitempty"`
	LeastConn              bool          `json:"least_conn"`
	BackHosts              []BackendHost `json:"back_hosts"`
	MtlsCertificate        string        `json:"mtls_certificate,omitempty"`
	MtlsTrustedCertificate string        `json:"mtls_trusted_certificate,omitempty"`
}

// ActiveConfig represents an active configuration entry in a mobile application group
type ActiveConfig struct {
	Active bool   `json:"active"`
	JSON   string `json:"json"`
	Name   string `json:"name"`
}

// Signature represents a signature entry in a mobile application group
type Signature struct {
	Active bool   `json:"active"`
	Hash   string `json:"hash"`
	Name   string `json:"name"`
}

// MobileApplicationGroup represents a mobile application group in the API
type MobileApplicationGroup struct {
	ID           string         `json:"id"`
	Name         string         `json:"name"`
	Description  string         `json:"description,omitempty"`
	UIDHeader    string         `json:"uid_header,omitempty"`
	Grace        string         `json:"grace,omitempty"`
	ActiveConfig []ActiveConfig `json:"active_config"`
	Signatures   []Signature    `json:"signatures"`
}

// EdgeFunction represents an edge function in the API
type EdgeFunction struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Code        string `json:"code"`
	Phase       string `json:"phase"`
}
