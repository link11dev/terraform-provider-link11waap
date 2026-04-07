package resources

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/link11/terraform-provider-link11waap/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// configureResourceWithMock configures a resource with a mock client backed by a TLS test server.
func configureResourceWithMock(t *testing.T, r resource.Resource, handler http.Handler) {
	t.Helper()
	ctx := context.Background()
	c, _ := newMockClient(t, handler)

	req := resource.ConfigureRequest{ProviderData: c}
	resp := &resource.ConfigureResponse{}
	if cr, ok := r.(interface {
		Configure(context.Context, resource.ConfigureRequest, *resource.ConfigureResponse)
	}); ok {
		cr.Configure(ctx, req, resp)
	}
	require.False(t, resp.Diagnostics.HasError())
}

// readWithMock performs a full Read operation using a mock API server.
func readWithMock(t *testing.T, r resource.Resource, stateValues map[string]tftypes.Value) *resource.ReadResponse {
	t.Helper()
	ctx := context.Background()

	state := buildTerraformState(ctx, t, r, stateValues)

	sReq := schemaReq()
	sResp := schemaResp()
	r.Schema(ctx, sReq, sResp)

	readResp := &resource.ReadResponse{
		State: tfsdk.State{
			Schema: sResp.Schema,
			Raw:    state.Raw.Copy(),
		},
	}
	r.Read(ctx, resource.ReadRequest{State: state}, readResp)
	return readResp
}

// --- ACL Profile Read ---

func TestACLProfileResource_Read_WithMock(t *testing.T) {
	r := &ACLProfileResource{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ACLProfile{
			ID:          "acl1",
			Name:        "test-acl",
			Description: "desc",
			Action:      "deny",
			Tags:        []string{"tag1"},
			Allow:       []string{"allow1"},
			AllowBot:    []string{"bot1"},
			Deny:        []string{"deny1"},
			DenyBot:     []string{"dbot1"},
			ForceDeny:   []string{"fd1"},
			Passthrough: []string{"pt1"},
		})
	})
	configureResourceWithMock(t, r, handler)

	resp := readWithMock(t, r, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"id":        tftypes.NewValue(tftypes.String, "acl1"),
	})

	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}

func TestACLProfileResource_Read_NotFound(t *testing.T) {
	r := &ACLProfileResource{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"not found"}`))
	})
	configureResourceWithMock(t, r, handler)

	resp := readWithMock(t, r, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"id":        tftypes.NewValue(tftypes.String, "missing"),
	})

	assert.False(t, resp.Diagnostics.HasError(), "not found should not produce error, resource should be removed")
}

// --- Edge Function Read ---

func TestEdgeFunctionResource_Read_WithMock(t *testing.T) {
	r := &EdgeFunctionResource{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.EdgeFunction{
			ID:          "ef1",
			Name:        "test-ef",
			Description: "edge func",
			Code:        "print('hello')",
			Phase:       "request_post",
		})
	})
	configureResourceWithMock(t, r, handler)

	resp := readWithMock(t, r, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"id":        tftypes.NewValue(tftypes.String, "ef1"),
	})

	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}

func TestEdgeFunctionResource_Read_NotFound(t *testing.T) {
	r := &EdgeFunctionResource{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"not found"}`))
	})
	configureResourceWithMock(t, r, handler)

	resp := readWithMock(t, r, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"id":        tftypes.NewValue(tftypes.String, "missing"),
	})

	assert.False(t, resp.Diagnostics.HasError())
}

// --- User Read ---

func TestUserResource_Read_WithMock(t *testing.T) {
	r := &UserResource{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.User{
			ID:          "user1",
			ACL:         10,
			ContactName: "John",
			Email:       "john@test.com",
			Mobile:      "+1234567890",
			OrgID:       "org1",
			OrgName:     "Org Inc",
		})
	})
	configureResourceWithMock(t, r, handler)

	resp := readWithMock(t, r, map[string]tftypes.Value{
		"id":           tftypes.NewValue(tftypes.String, "user1"),
		"acl":          tftypes.NewValue(tftypes.Number, 10),
		"contact_name": tftypes.NewValue(tftypes.String, "John"),
		"email":        tftypes.NewValue(tftypes.String, "john@test.com"),
		"mobile":       tftypes.NewValue(tftypes.String, "+1234567890"),
		"org_id":       tftypes.NewValue(tftypes.String, "org1"),
		"org_name":     tftypes.NewValue(tftypes.String, "Org Inc"),
	})

	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}

func TestUserResource_Read_NotFound(t *testing.T) {
	r := &UserResource{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"not found"}`))
	})
	configureResourceWithMock(t, r, handler)

	resp := readWithMock(t, r, map[string]tftypes.Value{
		"id": tftypes.NewValue(tftypes.String, "missing"),
	})

	assert.False(t, resp.Diagnostics.HasError())
}

// --- Server Group Read ---

func TestServerGroupResource_Read_WithMock(t *testing.T) {
	r := &ServerGroupResource{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ServerGroup{
			ID:                     "sg1",
			Name:                   "site1",
			Description:            "desc",
			ServerNames:            []string{"example.com"},
			SecurityPolicy:         "sp1",
			RoutingProfile:         "rp1",
			ProxyTemplate:          "pt1",
			ChallengeCookieDomain:  "example.com",
			SSLCertificate:         "cert1",
			ClientCertificate:      "ccert1",
			ClientCertificateMode:  "verify",
			MobileApplicationGroup: "mag1",
		})
	})
	configureResourceWithMock(t, r, handler)

	resp := readWithMock(t, r, map[string]tftypes.Value{
		"config_id":               tftypes.NewValue(tftypes.String, "cfg1"),
		"id":                      tftypes.NewValue(tftypes.String, "sg1"),
		"name":                    tftypes.NewValue(tftypes.String, "site1"),
		"server_names":            tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{tftypes.NewValue(tftypes.String, "example.com")}),
		"security_policy":         tftypes.NewValue(tftypes.String, "sp1"),
		"routing_profile":         tftypes.NewValue(tftypes.String, "rp1"),
		"proxy_template":          tftypes.NewValue(tftypes.String, "pt1"),
		"challenge_cookie_domain": tftypes.NewValue(tftypes.String, "example.com"),
	})

	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}

func TestServerGroupResource_Read_NotFound(t *testing.T) {
	r := &ServerGroupResource{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"not found"}`))
	})
	configureResourceWithMock(t, r, handler)

	resp := readWithMock(t, r, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"id":        tftypes.NewValue(tftypes.String, "missing"),
	})

	assert.False(t, resp.Diagnostics.HasError())
}

// --- Server Group Read with empty optional fields ---

func TestServerGroupResource_Read_EmptyOptionalFields(t *testing.T) {
	r := &ServerGroupResource{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ServerGroup{
			ID:                    "sg1",
			Name:                  "site1",
			ServerNames:           []string{"example.com"},
			SecurityPolicy:        "sp1",
			RoutingProfile:        "rp1",
			ProxyTemplate:         "pt1",
			ChallengeCookieDomain: "example.com",
			// Empty optional fields: SSLCertificate, ClientCertificate, MobileApplicationGroup
		})
	})
	configureResourceWithMock(t, r, handler)

	resp := readWithMock(t, r, map[string]tftypes.Value{
		"config_id":               tftypes.NewValue(tftypes.String, "cfg1"),
		"id":                      tftypes.NewValue(tftypes.String, "sg1"),
		"server_names":            tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{tftypes.NewValue(tftypes.String, "example.com")}),
		"security_policy":         tftypes.NewValue(tftypes.String, "sp1"),
		"routing_profile":         tftypes.NewValue(tftypes.String, "rp1"),
		"proxy_template":          tftypes.NewValue(tftypes.String, "pt1"),
		"challenge_cookie_domain": tftypes.NewValue(tftypes.String, "example.com"),
	})

	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}

// --- Backend Service Read ---

func TestBackendServiceResource_Read_WithMock(t *testing.T) {
	r := &BackendServiceResource{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.BackendService{
			ID:                     "bs1",
			Name:                   "test-bs",
			Description:            "desc",
			HTTP11:                 true,
			TransportMode:          "default",
			Sticky:                 "none",
			StickyCookieName:       "",
			LeastConn:              false,
			MtlsCertificate:        "mtls-cert",
			MtlsTrustedCertificate: "mtls-trusted",
			BackHosts: []client.BackendHost{
				{
					Host:         "backend.example.com",
					HTTPPorts:    []int{80},
					HTTPSPorts:   []int{443},
					Weight:       1,
					MaxFails:     3,
					FailTimeout:  10,
					Down:         false,
					MonitorState: "up",
					Backup:       false,
				},
			},
		})
	})
	configureResourceWithMock(t, r, handler)

	resp := readWithMock(t, r, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"id":        tftypes.NewValue(tftypes.String, "bs1"),
	})

	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}

func TestBackendServiceResource_Read_NullBackHosts(t *testing.T) {
	r := &BackendServiceResource{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.BackendService{
			ID:            "bs1",
			Name:          "test-bs",
			TransportMode: "default",
			Sticky:        "none",
		})
	})
	configureResourceWithMock(t, r, handler)

	resp := readWithMock(t, r, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"id":        tftypes.NewValue(tftypes.String, "bs1"),
	})

	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}

func TestBackendServiceResource_Read_NotFound(t *testing.T) {
	r := &BackendServiceResource{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"not found"}`))
	})
	configureResourceWithMock(t, r, handler)

	resp := readWithMock(t, r, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"id":        tftypes.NewValue(tftypes.String, "missing"),
	})

	assert.False(t, resp.Diagnostics.HasError())
}

// --- Security Policy Read ---

func TestSecurityPolicyResource_Read_WithMock(t *testing.T) {
	r := &SecurityPolicyResource{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.SecurityPolicy{
			ID:          "sp1",
			Name:        "test-sp",
			Description: "desc",
			Tags:        []string{"tag1", "tag2"},
			Session: []interface{}{
				map[string]interface{}{"attrs": "ip"},
			},
			SessionIDs: []interface{}{
				map[string]interface{}{"cookies": "sid"},
			},
			Map: []client.SecProfileMap{
				{
					ID:                         "entry1",
					Name:                       "Default",
					Match:                      "/",
					ACLProfile:                 "acl1",
					ACLProfileActive:           true,
					ContentFilterProfile:       "cf1",
					ContentFilterProfileActive: true,
					BackendService:             "be1",
					Description:                "default entry",
					RateLimitRules:             []string{"rl1"},
					EdgeFunctions:              []string{"ef1"},
				},
			},
		})
	})
	configureResourceWithMock(t, r, handler)

	resp := readWithMock(t, r, map[string]tftypes.Value{
		"config_id":   tftypes.NewValue(tftypes.String, "cfg1"),
		"id":          tftypes.NewValue(tftypes.String, "sp1"),
		"name":        tftypes.NewValue(tftypes.String, "test-sp"),
		"description": tftypes.NewValue(tftypes.String, "desc"),
	})

	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}

func TestSecurityPolicyResource_Read_EmptyTags(t *testing.T) {
	r := &SecurityPolicyResource{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.SecurityPolicy{
			ID:   "sp1",
			Name: "test-sp",
			Map:  []client.SecProfileMap{},
		})
	})
	configureResourceWithMock(t, r, handler)

	resp := readWithMock(t, r, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"id":        tftypes.NewValue(tftypes.String, "sp1"),
	})

	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}

func TestSecurityPolicyResource_Read_NotFound(t *testing.T) {
	r := &SecurityPolicyResource{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"not found"}`))
	})
	configureResourceWithMock(t, r, handler)

	resp := readWithMock(t, r, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"id":        tftypes.NewValue(tftypes.String, "missing"),
	})

	assert.False(t, resp.Diagnostics.HasError())
}

// --- Rate Limit Rule Read ---

func TestRateLimitRuleResource_Read_WithMock(t *testing.T) {
	r := &RateLimitRuleResource{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.RateLimitRule{
			ID:          "rl1",
			Name:        "test-rule",
			Description: "rate limit",
			Global:      false,
			Active:      true,
			Timeframe:   60,
			Threshold:   100,
			TTL:         300,
			Action:      "action-monitor",
			IsActionBan: false,
			Tags:        []string{"tag1"},
			Key: []interface{}{
				map[string]interface{}{"attrs": "ip"},
			},
			Pairwith: map[string]interface{}{"self": "self"},
			Include: client.RateLimitTagFilter{
				Relation: "OR",
				Tags:     []string{"include-tag"},
			},
			Exclude: client.RateLimitTagFilter{
				Relation: "AND",
				Tags:     []string{"exclude-tag"},
			},
		})
	})
	configureResourceWithMock(t, r, handler)

	resp := readWithMock(t, r, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"id":        tftypes.NewValue(tftypes.String, "rl1"),
	})

	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}

func TestRateLimitRuleResource_Read_NotFound(t *testing.T) {
	r := &RateLimitRuleResource{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"not found"}`))
	})
	configureResourceWithMock(t, r, handler)

	resp := readWithMock(t, r, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"id":        tftypes.NewValue(tftypes.String, "missing"),
	})

	assert.False(t, resp.Diagnostics.HasError())
}

func TestRateLimitRuleResource_Read_NullPairwith(t *testing.T) {
	r := &RateLimitRuleResource{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.RateLimitRule{
			ID:        "rl1",
			Name:      "test-rule",
			Active:    true,
			Timeframe: 60,
			Threshold: 100,
			Action:    "action-monitor",
			Key: []interface{}{
				map[string]interface{}{"attrs": "ip"},
			},
			// Pairwith is nil
		})
	})
	configureResourceWithMock(t, r, handler)

	resp := readWithMock(t, r, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"id":        tftypes.NewValue(tftypes.String, "rl1"),
	})

	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}

// --- Mobile Application Group Read ---

func TestMobileApplicationGroupResource_Read_WithMock(t *testing.T) {
	r := &MobileApplicationGroupResource{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.MobileApplicationGroup{
			ID:          "mag1",
			Name:        "test-mag",
			Description: "desc",
			UIDHeader:   "X-UID",
			Grace:       "5m",
			ActiveConfig: []client.ActiveConfig{
				{Active: true, JSON: "{}", Name: "config1"},
			},
			Signatures: []client.Signature{
				{Active: true, Hash: "abc123", Name: "sig1"},
			},
		})
	})
	configureResourceWithMock(t, r, handler)

	resp := readWithMock(t, r, map[string]tftypes.Value{
		"config_id":   tftypes.NewValue(tftypes.String, "cfg1"),
		"id":          tftypes.NewValue(tftypes.String, "mag1"),
		"name":        tftypes.NewValue(tftypes.String, "test-mag"),
		"description": tftypes.NewValue(tftypes.String, "desc"),
		"uid_header":  tftypes.NewValue(tftypes.String, "X-UID"),
		"grace":       tftypes.NewValue(tftypes.String, "5m"),
	})

	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}

func TestMobileApplicationGroupResource_Read_NullSigsAndConfig(t *testing.T) {
	r := &MobileApplicationGroupResource{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.MobileApplicationGroup{
			ID:   "mag1",
			Name: "test-mag",
			// nil ActiveConfig and Signatures
		})
	})
	configureResourceWithMock(t, r, handler)

	resp := readWithMock(t, r, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"id":        tftypes.NewValue(tftypes.String, "mag1"),
	})

	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}

func TestMobileApplicationGroupResource_Read_NotFound(t *testing.T) {
	r := &MobileApplicationGroupResource{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"not found"}`))
	})
	configureResourceWithMock(t, r, handler)

	resp := readWithMock(t, r, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"id":        tftypes.NewValue(tftypes.String, "missing"),
	})

	assert.False(t, resp.Diagnostics.HasError())
}

// --- Load Balancer Regions Read ---

func TestLoadBalancerRegionsResource_Read_WithMock(t *testing.T) {
	r := &LoadBalancerRegionsResource{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.LoadBalancerRegions{
			CityCodes: map[string]string{"NYC": "New York"},
			LBs: []client.LoadBalancerRegion{
				{
					ID:              "lb1",
					Name:            "load-balancer-1",
					Regions:         map[string]string{"ams": "automatic", "nyc": "manual"},
					UpstreamRegions: []string{"ams", "nyc"},
				},
			},
		})
	})
	configureResourceWithMock(t, r, handler)

	resp := readWithMock(t, r, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"lb_id":     tftypes.NewValue(tftypes.String, "lb1"),
		"regions": tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, map[string]tftypes.Value{
			"ams": tftypes.NewValue(tftypes.String, "automatic"),
		}),
	})

	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}

func TestLoadBalancerRegionsResource_Read_LBNotFound(t *testing.T) {
	r := &LoadBalancerRegionsResource{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.LoadBalancerRegions{
			LBs: []client.LoadBalancerRegion{
				{ID: "other-lb", Name: "other"},
			},
		})
	})
	configureResourceWithMock(t, r, handler)

	resp := readWithMock(t, r, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"lb_id":     tftypes.NewValue(tftypes.String, "lb1"),
		"regions": tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, map[string]tftypes.Value{
			"ams": tftypes.NewValue(tftypes.String, "automatic"),
		}),
	})

	// LB not found in response - resource should be removed
	assert.False(t, resp.Diagnostics.HasError())
}

// --- Certificate Read ---

func TestCertificateResource_Read_WithMock(t *testing.T) {
	r := &CertificateResource{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.Certificate{
			ID:            "cert1",
			Name:          "test-cert",
			Subject:       "CN=test",
			Issuer:        "CN=issuer",
			SAN:           []string{"test.com", "*.test.com"},
			Expires:       "2025-12-31",
			Uploaded:      "2024-01-01",
			LEAutoRenew:   false,
			LEAutoReplace: false,
			Revoked:       false,
			Links:         []client.ProviderLink{{Provider: "aws", Link: "https://aws.example.com/cert1", Region: "us-east-1"}},
			ProviderLinks: []client.ProviderLink{{Provider: "gcp", Link: "https://gcp.example.com/cert1", Region: "us-central1"}},
		})
	})
	configureResourceWithMock(t, r, handler)

	resp := readWithMock(t, r, map[string]tftypes.Value{
		"config_id":       tftypes.NewValue(tftypes.String, "cfg1"),
		"id":              tftypes.NewValue(tftypes.String, "cert1"),
		"le_auto_renew":   tftypes.NewValue(tftypes.Bool, false),
		"le_auto_replace": tftypes.NewValue(tftypes.Bool, false),
	})

	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}

func TestCertificateResource_Read_NotFound(t *testing.T) {
	r := &CertificateResource{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"not found"}`))
	})
	configureResourceWithMock(t, r, handler)

	resp := readWithMock(t, r, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"id":        tftypes.NewValue(tftypes.String, "missing"),
	})

	assert.False(t, resp.Diagnostics.HasError())
}

// --- Load Balancer Certificate Read ---

func TestLoadBalancerCertificateResource_Read_WithMock(t *testing.T) {
	r := &LoadBalancerCertificateResource{}
	callCount := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		if callCount == 1 {
			// First call: GetCertificate
			json.NewEncoder(w).Encode(client.Certificate{
				ID:            "cert1",
				Links:         []client.ProviderLink{{Provider: "aws", Link: "https://aws.example.com/cert1", Region: "us-east-1"}},
				ProviderLinks: []client.ProviderLink{{Provider: "gcp", Link: "https://gcp.example.com/cert1", Region: "us-central1"}},
			})
		} else {
			// Second call: ListLoadBalancers
			json.NewEncoder(w).Encode(client.ListResponse[client.LoadBalancer]{
				Total: 1,
				Items: []client.LoadBalancer{
					{
						Name:               "lb1",
						Certificates:       []string{"https://aws.example.com/cert1"},
						DefaultCertificate: "",
					},
				},
			})
		}
	})
	configureResourceWithMock(t, r, handler)

	resp := readWithMock(t, r, map[string]tftypes.Value{
		"config_id":          tftypes.NewValue(tftypes.String, "cfg1"),
		"id":                 tftypes.NewValue(tftypes.String, "cfg1/lb1/cert1"),
		"load_balancer_name": tftypes.NewValue(tftypes.String, "lb1"),
		"certificate_id":     tftypes.NewValue(tftypes.String, "cert1"),
		"provider_type":      tftypes.NewValue(tftypes.String, "aws"),
		"region":             tftypes.NewValue(tftypes.String, "us-east-1"),
		"listener":           tftypes.NewValue(tftypes.String, "arn:test"),
		"listener_port":      tftypes.NewValue(tftypes.Number, 443),
		"is_default":         tftypes.NewValue(tftypes.Bool, false),
		"elbv2":              tftypes.NewValue(tftypes.Bool, true),
	})

	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}

func TestLoadBalancerCertificateResource_Read_CertNotFound(t *testing.T) {
	r := &LoadBalancerCertificateResource{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"not found"}`))
	})
	configureResourceWithMock(t, r, handler)

	resp := readWithMock(t, r, map[string]tftypes.Value{
		"config_id":          tftypes.NewValue(tftypes.String, "cfg1"),
		"id":                 tftypes.NewValue(tftypes.String, "cfg1/lb1/cert1"),
		"load_balancer_name": tftypes.NewValue(tftypes.String, "lb1"),
		"certificate_id":     tftypes.NewValue(tftypes.String, "cert1"),
		"provider_type":      tftypes.NewValue(tftypes.String, "aws"),
		"region":             tftypes.NewValue(tftypes.String, "us-east-1"),
		"listener":           tftypes.NewValue(tftypes.String, "arn:test"),
		"listener_port":      tftypes.NewValue(tftypes.Number, 443),
		"is_default":         tftypes.NewValue(tftypes.Bool, false),
		"elbv2":              tftypes.NewValue(tftypes.Bool, true),
	})

	// Should remove resource, no error
	assert.False(t, resp.Diagnostics.HasError())
}

func TestLoadBalancerCertificateResource_Read_CertNotAttached(t *testing.T) {
	r := &LoadBalancerCertificateResource{}
	callCount := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		if callCount == 1 {
			// GetCertificate
			json.NewEncoder(w).Encode(client.Certificate{
				ID:    "cert1",
				Links: []client.ProviderLink{{Provider: "aws", Link: "https://aws.example.com/cert1", Region: "us-east-1"}},
			})
		} else {
			// ListLoadBalancers - cert not attached
			json.NewEncoder(w).Encode(client.ListResponse[client.LoadBalancer]{
				Total: 1,
				Items: []client.LoadBalancer{
					{
						Name:         "lb1",
						Certificates: []string{"https://aws.example.com/other-cert"},
					},
				},
			})
		}
	})
	configureResourceWithMock(t, r, handler)

	resp := readWithMock(t, r, map[string]tftypes.Value{
		"config_id":          tftypes.NewValue(tftypes.String, "cfg1"),
		"id":                 tftypes.NewValue(tftypes.String, "cfg1/lb1/cert1"),
		"load_balancer_name": tftypes.NewValue(tftypes.String, "lb1"),
		"certificate_id":     tftypes.NewValue(tftypes.String, "cert1"),
		"provider_type":      tftypes.NewValue(tftypes.String, "aws"),
		"region":             tftypes.NewValue(tftypes.String, "us-east-1"),
		"listener":           tftypes.NewValue(tftypes.String, "arn:test"),
		"listener_port":      tftypes.NewValue(tftypes.Number, 443),
		"is_default":         tftypes.NewValue(tftypes.Bool, false),
		"elbv2":              tftypes.NewValue(tftypes.Bool, true),
	})

	// Resource removed because cert not found on LB
	assert.False(t, resp.Diagnostics.HasError())
}

// --- Load Balancer Certificate Read - cert matched by ID ---

func TestLoadBalancerCertificateResource_Read_CertMatchedByID(t *testing.T) {
	r := &LoadBalancerCertificateResource{}
	callCount := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		if callCount == 1 {
			json.NewEncoder(w).Encode(client.Certificate{
				ID: "cert1",
			})
		} else {
			json.NewEncoder(w).Encode(client.ListResponse[client.LoadBalancer]{
				Total: 1,
				Items: []client.LoadBalancer{
					{Name: "lb1", Certificates: []string{"cert1"}},
				},
			})
		}
	})
	configureResourceWithMock(t, r, handler)

	resp := readWithMock(t, r, map[string]tftypes.Value{
		"config_id":          tftypes.NewValue(tftypes.String, "cfg1"),
		"id":                 tftypes.NewValue(tftypes.String, "cfg1/lb1/cert1"),
		"load_balancer_name": tftypes.NewValue(tftypes.String, "lb1"),
		"certificate_id":     tftypes.NewValue(tftypes.String, "cert1"),
		"provider_type":      tftypes.NewValue(tftypes.String, "aws"),
		"region":             tftypes.NewValue(tftypes.String, "us-east-1"),
		"listener":           tftypes.NewValue(tftypes.String, "arn:test"),
		"listener_port":      tftypes.NewValue(tftypes.Number, 443),
		"is_default":         tftypes.NewValue(tftypes.Bool, false),
		"elbv2":              tftypes.NewValue(tftypes.Bool, true),
	})

	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}

// --- Load Balancer Certificate Read - cert matched by URL suffix ---

func TestLoadBalancerCertificateResource_Read_CertMatchedByURLSuffix(t *testing.T) {
	r := &LoadBalancerCertificateResource{}
	callCount := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		if callCount == 1 {
			json.NewEncoder(w).Encode(client.Certificate{
				ID: "cert1",
			})
		} else {
			json.NewEncoder(w).Encode(client.ListResponse[client.LoadBalancer]{
				Total: 1,
				Items: []client.LoadBalancer{
					{
						Name:         "lb1",
						Certificates: []string{"https://example.com/sometestplanet-uploaded-cert1"},
					},
				},
			})
		}
	})
	configureResourceWithMock(t, r, handler)

	resp := readWithMock(t, r, map[string]tftypes.Value{
		"config_id":          tftypes.NewValue(tftypes.String, "cfg1"),
		"id":                 tftypes.NewValue(tftypes.String, "cfg1/lb1/cert1"),
		"load_balancer_name": tftypes.NewValue(tftypes.String, "lb1"),
		"certificate_id":     tftypes.NewValue(tftypes.String, "cert1"),
		"provider_type":      tftypes.NewValue(tftypes.String, "aws"),
		"region":             tftypes.NewValue(tftypes.String, "us-east-1"),
		"listener":           tftypes.NewValue(tftypes.String, "arn:test"),
		"listener_port":      tftypes.NewValue(tftypes.Number, 443),
		"is_default":         tftypes.NewValue(tftypes.Bool, false),
		"elbv2":              tftypes.NewValue(tftypes.Bool, true),
	})

	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}

// --- Load Balancer Certificate Read - cert matched in DefaultCertificate ---

func TestLoadBalancerCertificateResource_Read_DefaultCertMatch(t *testing.T) {
	r := &LoadBalancerCertificateResource{}
	callCount := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		if callCount == 1 {
			json.NewEncoder(w).Encode(client.Certificate{
				ID: "cert1",
			})
		} else {
			json.NewEncoder(w).Encode(client.ListResponse[client.LoadBalancer]{
				Total: 1,
				Items: []client.LoadBalancer{
					{
						Name:               "lb1",
						Certificates:       []string{},
						DefaultCertificate: "cert1",
					},
				},
			})
		}
	})
	configureResourceWithMock(t, r, handler)

	resp := readWithMock(t, r, map[string]tftypes.Value{
		"config_id":          tftypes.NewValue(tftypes.String, "cfg1"),
		"id":                 tftypes.NewValue(tftypes.String, "cfg1/lb1/cert1"),
		"load_balancer_name": tftypes.NewValue(tftypes.String, "lb1"),
		"certificate_id":     tftypes.NewValue(tftypes.String, "cert1"),
		"provider_type":      tftypes.NewValue(tftypes.String, "aws"),
		"region":             tftypes.NewValue(tftypes.String, "us-east-1"),
		"listener":           tftypes.NewValue(tftypes.String, "arn:test"),
		"listener_port":      tftypes.NewValue(tftypes.Number, 443),
		"is_default":         tftypes.NewValue(tftypes.Bool, false),
		"elbv2":              tftypes.NewValue(tftypes.Bool, true),
	})

	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}
