package resources

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
)

// --- ACL Profile CRUD ---

func TestACLProfileResource_Create_NilClient(t *testing.T) {
	r := &ACLProfileResource{}
	vals := map[string]tftypes.Value{
		"config_id":   tftypes.NewValue(tftypes.String, "cfg1"),
		"id":          tftypes.NewValue(tftypes.String, nil),
		"name":        tftypes.NewValue(tftypes.String, "test"),
		"description": tftypes.NewValue(tftypes.String, ""),
		"action":      tftypes.NewValue(tftypes.String, nil),
	}
	testCreateWithNilClient(t, r, vals)
}

func TestACLProfileResource_Read_NilClient(t *testing.T) {
	r := &ACLProfileResource{}
	vals := map[string]tftypes.Value{
		"config_id":   tftypes.NewValue(tftypes.String, "cfg1"),
		"id":          tftypes.NewValue(tftypes.String, "acl1"),
		"name":        tftypes.NewValue(tftypes.String, "test"),
		"description": tftypes.NewValue(tftypes.String, ""),
	}
	testReadWithNilClient(t, r, vals)
}

func TestACLProfileResource_Update_NilClient(t *testing.T) {
	r := &ACLProfileResource{}
	vals := map[string]tftypes.Value{
		"config_id":   tftypes.NewValue(tftypes.String, "cfg1"),
		"id":          tftypes.NewValue(tftypes.String, "acl1"),
		"name":        tftypes.NewValue(tftypes.String, "test"),
		"description": tftypes.NewValue(tftypes.String, ""),
	}
	testUpdateWithNilClient(t, r, vals)
}

func TestACLProfileResource_Delete_NilClient(t *testing.T) {
	r := &ACLProfileResource{}
	vals := map[string]tftypes.Value{
		"config_id":   tftypes.NewValue(tftypes.String, "cfg1"),
		"id":          tftypes.NewValue(tftypes.String, "acl1"),
		"name":        tftypes.NewValue(tftypes.String, "test"),
		"description": tftypes.NewValue(tftypes.String, ""),
	}
	testDeleteWithNilClient(t, r, vals)
}

func TestACLProfileResource_Configure_InvalidType(t *testing.T) {
	r := &ACLProfileResource{}
	testConfigureWithInvalidType(t, r)
}

// --- Edge Function CRUD ---

func TestEdgeFunctionResource_Create_NilClient(t *testing.T) {
	r := &EdgeFunctionResource{}
	vals := map[string]tftypes.Value{
		"config_id":   tftypes.NewValue(tftypes.String, "cfg1"),
		"id":          tftypes.NewValue(tftypes.String, nil),
		"name":        tftypes.NewValue(tftypes.String, "test-ef"),
		"description": tftypes.NewValue(tftypes.String, ""),
		"code":        tftypes.NewValue(tftypes.String, "print('hello')"),
		"phase":       tftypes.NewValue(tftypes.String, "request_post"),
	}
	testCreateWithNilClient(t, r, vals)
}

func TestEdgeFunctionResource_Read_NilClient(t *testing.T) {
	r := &EdgeFunctionResource{}
	vals := map[string]tftypes.Value{
		"config_id":   tftypes.NewValue(tftypes.String, "cfg1"),
		"id":          tftypes.NewValue(tftypes.String, "ef1"),
		"name":        tftypes.NewValue(tftypes.String, "test"),
		"description": tftypes.NewValue(tftypes.String, ""),
		"code":        tftypes.NewValue(tftypes.String, "print('hello')"),
		"phase":       tftypes.NewValue(tftypes.String, "request_post"),
	}
	testReadWithNilClient(t, r, vals)
}

func TestEdgeFunctionResource_Update_NilClient(t *testing.T) {
	r := &EdgeFunctionResource{}
	vals := map[string]tftypes.Value{
		"config_id":   tftypes.NewValue(tftypes.String, "cfg1"),
		"id":          tftypes.NewValue(tftypes.String, "ef1"),
		"name":        tftypes.NewValue(tftypes.String, "test"),
		"description": tftypes.NewValue(tftypes.String, ""),
		"code":        tftypes.NewValue(tftypes.String, "print('hello')"),
		"phase":       tftypes.NewValue(tftypes.String, "request_post"),
	}
	testUpdateWithNilClient(t, r, vals)
}

func TestEdgeFunctionResource_Delete_NilClient(t *testing.T) {
	r := &EdgeFunctionResource{}
	vals := map[string]tftypes.Value{
		"config_id":   tftypes.NewValue(tftypes.String, "cfg1"),
		"id":          tftypes.NewValue(tftypes.String, "ef1"),
		"name":        tftypes.NewValue(tftypes.String, "test"),
		"description": tftypes.NewValue(tftypes.String, ""),
		"code":        tftypes.NewValue(tftypes.String, "print('hello')"),
		"phase":       tftypes.NewValue(tftypes.String, "request_post"),
	}
	testDeleteWithNilClient(t, r, vals)
}

func TestEdgeFunctionResource_Configure_InvalidType(t *testing.T) {
	r := &EdgeFunctionResource{}
	testConfigureWithInvalidType(t, r)
}

// --- Certificate CRUD ---

func TestCertificateResource_Create_NilClient(t *testing.T) {
	r := &CertificateResource{}
	vals := map[string]tftypes.Value{
		"config_id":       tftypes.NewValue(tftypes.String, "cfg1"),
		"id":              tftypes.NewValue(tftypes.String, nil),
		"cert_body":       tftypes.NewValue(tftypes.String, "cert-body"),
		"private_key":     tftypes.NewValue(tftypes.String, "private-key"),
		"le_auto_renew":   tftypes.NewValue(tftypes.Bool, false),
		"le_auto_replace": tftypes.NewValue(tftypes.Bool, false),
		"side":            tftypes.NewValue(tftypes.String, "server"),
	}
	testCreateWithNilClient(t, r, vals)
}

func TestCertificateResource_Read_NilClient(t *testing.T) {
	r := &CertificateResource{}
	vals := map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"id":        tftypes.NewValue(tftypes.String, "cert1"),
	}
	testReadWithNilClient(t, r, vals)
}

func TestCertificateResource_Update_NilClient(t *testing.T) {
	r := &CertificateResource{}
	vals := map[string]tftypes.Value{
		"config_id":       tftypes.NewValue(tftypes.String, "cfg1"),
		"id":              tftypes.NewValue(tftypes.String, "cert1"),
		"le_auto_renew":   tftypes.NewValue(tftypes.Bool, true),
		"le_auto_replace": tftypes.NewValue(tftypes.Bool, false),
	}
	testUpdateWithNilClient(t, r, vals)
}

func TestCertificateResource_Delete_NilClient(t *testing.T) {
	r := &CertificateResource{}
	vals := map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"id":        tftypes.NewValue(tftypes.String, "cert1"),
	}
	testDeleteWithNilClient(t, r, vals)
}

func TestCertificateResource_Configure_InvalidType(t *testing.T) {
	r := &CertificateResource{}
	testConfigureWithInvalidType(t, r)
}

// --- Server Group CRUD ---

func TestServerGroupResource_Create_NilClient(t *testing.T) {
	r := &ServerGroupResource{}
	vals := map[string]tftypes.Value{
		"config_id":               tftypes.NewValue(tftypes.String, "cfg1"),
		"id":                      tftypes.NewValue(tftypes.String, nil),
		"name":                    tftypes.NewValue(tftypes.String, "site1"),
		"description":             tftypes.NewValue(tftypes.String, ""),
		"server_names":            tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{tftypes.NewValue(tftypes.String, "example.com")}),
		"security_policy":         tftypes.NewValue(tftypes.String, "sp1"),
		"routing_profile":         tftypes.NewValue(tftypes.String, "rp1"),
		"proxy_template":          tftypes.NewValue(tftypes.String, "pt1"),
		"challenge_cookie_domain": tftypes.NewValue(tftypes.String, "example.com"),
		"client_certificate_mode": tftypes.NewValue(tftypes.String, "off"),
	}
	testCreateWithNilClient(t, r, vals)
}

func TestServerGroupResource_Read_NilClient(t *testing.T) {
	r := &ServerGroupResource{}
	vals := map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"id":        tftypes.NewValue(tftypes.String, "sg1"),
	}
	testReadWithNilClient(t, r, vals)
}

func TestServerGroupResource_Update_NilClient(t *testing.T) {
	r := &ServerGroupResource{}
	vals := map[string]tftypes.Value{
		"config_id":               tftypes.NewValue(tftypes.String, "cfg1"),
		"id":                      tftypes.NewValue(tftypes.String, "sg1"),
		"name":                    tftypes.NewValue(tftypes.String, "site1"),
		"description":             tftypes.NewValue(tftypes.String, ""),
		"server_names":            tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{tftypes.NewValue(tftypes.String, "example.com")}),
		"security_policy":         tftypes.NewValue(tftypes.String, "sp1"),
		"routing_profile":         tftypes.NewValue(tftypes.String, "rp1"),
		"proxy_template":          tftypes.NewValue(tftypes.String, "pt1"),
		"challenge_cookie_domain": tftypes.NewValue(tftypes.String, "example.com"),
		"client_certificate_mode": tftypes.NewValue(tftypes.String, "off"),
	}
	testUpdateWithNilClient(t, r, vals)
}

func TestServerGroupResource_Delete_NilClient(t *testing.T) {
	r := &ServerGroupResource{}
	vals := map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"id":        tftypes.NewValue(tftypes.String, "sg1"),
	}
	testDeleteWithNilClient(t, r, vals)
}

func TestServerGroupResource_Configure_InvalidType(t *testing.T) {
	r := &ServerGroupResource{}
	testConfigureWithInvalidType(t, r)
}

// --- User CRUD ---

func TestUserResource_Create_NilClient(t *testing.T) {
	r := &UserResource{}
	vals := map[string]tftypes.Value{
		"id":           tftypes.NewValue(tftypes.String, nil),
		"acl":          tftypes.NewValue(tftypes.Number, 10),
		"contact_name": tftypes.NewValue(tftypes.String, "John"),
		"email":        tftypes.NewValue(tftypes.String, "john@test.com"),
		"mobile":       tftypes.NewValue(tftypes.String, "+1234567890"),
		"org_id":       tftypes.NewValue(tftypes.String, "org1"),
		"org_name":     tftypes.NewValue(tftypes.String, nil),
	}
	testCreateWithNilClient(t, r, vals)
}

func TestUserResource_Read_NilClient(t *testing.T) {
	r := &UserResource{}
	vals := map[string]tftypes.Value{
		"id":           tftypes.NewValue(tftypes.String, "user1"),
		"acl":          tftypes.NewValue(tftypes.Number, 10),
		"contact_name": tftypes.NewValue(tftypes.String, "John"),
		"email":        tftypes.NewValue(tftypes.String, "john@test.com"),
		"mobile":       tftypes.NewValue(tftypes.String, "+1234567890"),
		"org_id":       tftypes.NewValue(tftypes.String, "org1"),
		"org_name":     tftypes.NewValue(tftypes.String, "Test Org"),
	}
	testReadWithNilClient(t, r, vals)
}

func TestUserResource_Update_NilClient(t *testing.T) {
	r := &UserResource{}
	vals := map[string]tftypes.Value{
		"id":           tftypes.NewValue(tftypes.String, "user1"),
		"acl":          tftypes.NewValue(tftypes.Number, 15),
		"contact_name": tftypes.NewValue(tftypes.String, "Jane"),
		"email":        tftypes.NewValue(tftypes.String, "jane@test.com"),
		"mobile":       tftypes.NewValue(tftypes.String, "+0987654321"),
		"org_id":       tftypes.NewValue(tftypes.String, "org1"),
		"org_name":     tftypes.NewValue(tftypes.String, "Test Org"),
	}
	testUpdateWithNilClient(t, r, vals)
}

func TestUserResource_Delete_NilClient(t *testing.T) {
	r := &UserResource{}
	vals := map[string]tftypes.Value{
		"id":           tftypes.NewValue(tftypes.String, "user1"),
		"acl":          tftypes.NewValue(tftypes.Number, 10),
		"contact_name": tftypes.NewValue(tftypes.String, "John"),
		"email":        tftypes.NewValue(tftypes.String, "john@test.com"),
		"mobile":       tftypes.NewValue(tftypes.String, "+1234567890"),
		"org_id":       tftypes.NewValue(tftypes.String, "org1"),
		"org_name":     tftypes.NewValue(tftypes.String, "Test Org"),
	}
	testDeleteWithNilClient(t, r, vals)
}

func TestUserResource_Configure_InvalidType(t *testing.T) {
	r := &UserResource{}
	testConfigureWithInvalidType(t, r)
}

func TestUserResource_ImportState_PassthroughID(t *testing.T) {
	r := &UserResource{}
	resp := testImportState(t, r, "user123")

	assert.False(t, resp.Diagnostics.HasError())
}

// --- Publish CRUD ---

func TestPublishResource_Create_NilClient(t *testing.T) {
	r := &PublishResource{}
	vals := map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"id":        tftypes.NewValue(tftypes.String, nil),
	}
	testCreateWithNilClient(t, r, vals)
}

func TestPublishResource_Read_NilClient(t *testing.T) {
	r := &PublishResource{}
	vals := map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"id":        tftypes.NewValue(tftypes.String, "cfg1"),
	}
	testReadWithNilClient(t, r, vals)
}

func TestPublishResource_Update_NilClient(t *testing.T) {
	r := &PublishResource{}
	vals := map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"id":        tftypes.NewValue(tftypes.String, "cfg1"),
	}
	testUpdateWithNilClient(t, r, vals)
}

// --- Load Balancer Certificate CRUD ---

func TestLoadBalancerCertificateResource_Create_NilClient(t *testing.T) {
	r := &LoadBalancerCertificateResource{}
	vals := map[string]tftypes.Value{
		"config_id":          tftypes.NewValue(tftypes.String, "cfg1"),
		"id":                 tftypes.NewValue(tftypes.String, nil),
		"load_balancer_name": tftypes.NewValue(tftypes.String, "lb1"),
		"certificate_id":     tftypes.NewValue(tftypes.String, "cert1"),
		"provider_type":      tftypes.NewValue(tftypes.String, "aws"),
		"region":             tftypes.NewValue(tftypes.String, "us-east-1"),
		"listener":           tftypes.NewValue(tftypes.String, "arn:aws:elasticloadbalancing:us-east-1:123456789:listener/abc"),
		"listener_port":      tftypes.NewValue(tftypes.Number, 443),
		"is_default":         tftypes.NewValue(tftypes.Bool, false),
		"elbv2":              tftypes.NewValue(tftypes.Bool, true),
	}
	testCreateWithNilClient(t, r, vals)
}

func TestLoadBalancerCertificateResource_Read_NilClient(t *testing.T) {
	r := &LoadBalancerCertificateResource{}
	vals := map[string]tftypes.Value{
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
	}
	testReadWithNilClient(t, r, vals)
}

func TestLoadBalancerCertificateResource_Update_NilClient(t *testing.T) {
	r := &LoadBalancerCertificateResource{}
	vals := map[string]tftypes.Value{
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
	}
	testUpdateWithNilClient(t, r, vals)
}

func TestLoadBalancerCertificateResource_Delete_NilClient(t *testing.T) {
	r := &LoadBalancerCertificateResource{}
	vals := map[string]tftypes.Value{
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
	}
	testDeleteWithNilClient(t, r, vals)
}

func TestLoadBalancerCertificateResource_Configure_InvalidType(t *testing.T) {
	r := &LoadBalancerCertificateResource{}
	testConfigureWithInvalidType(t, r)
}

// --- Load Balancer Regions CRUD ---

func TestLoadBalancerRegionsResource_Create_NilClient(t *testing.T) {
	r := &LoadBalancerRegionsResource{}
	vals := map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"lb_id":     tftypes.NewValue(tftypes.String, "lb1"),
		"regions":   tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, map[string]tftypes.Value{"ams": tftypes.NewValue(tftypes.String, "automatic")}),
	}
	testCreateWithNilClient(t, r, vals)
}

func TestLoadBalancerRegionsResource_Read_NilClient(t *testing.T) {
	r := &LoadBalancerRegionsResource{}
	vals := map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"lb_id":     tftypes.NewValue(tftypes.String, "lb1"),
		"regions":   tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, map[string]tftypes.Value{"ams": tftypes.NewValue(tftypes.String, "automatic")}),
	}
	testReadWithNilClient(t, r, vals)
}

func TestLoadBalancerRegionsResource_Update_NilClient(t *testing.T) {
	r := &LoadBalancerRegionsResource{}
	vals := map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"lb_id":     tftypes.NewValue(tftypes.String, "lb1"),
		"regions":   tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, map[string]tftypes.Value{"ams": tftypes.NewValue(tftypes.String, "eu-west-1")}),
	}
	testUpdateWithNilClient(t, r, vals)
}

func TestLoadBalancerRegionsResource_Configure_InvalidType(t *testing.T) {
	r := &LoadBalancerRegionsResource{}
	testConfigureWithInvalidType(t, r)
}

// --- Mobile Application Group CRUD ---

func TestMobileApplicationGroupResource_Create_NilClient(t *testing.T) {
	r := &MobileApplicationGroupResource{}
	vals := map[string]tftypes.Value{
		"config_id":   tftypes.NewValue(tftypes.String, "cfg1"),
		"id":          tftypes.NewValue(tftypes.String, nil),
		"name":        tftypes.NewValue(tftypes.String, "test-mag"),
		"description": tftypes.NewValue(tftypes.String, ""),
		"uid_header":  tftypes.NewValue(tftypes.String, ""),
		"grace":       tftypes.NewValue(tftypes.String, ""),
	}
	testCreateWithNilClient(t, r, vals)
}

func TestMobileApplicationGroupResource_Read_NilClient(t *testing.T) {
	r := &MobileApplicationGroupResource{}
	vals := map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"id":        tftypes.NewValue(tftypes.String, "mag1"),
	}
	testReadWithNilClient(t, r, vals)
}

func TestMobileApplicationGroupResource_Update_NilClient(t *testing.T) {
	r := &MobileApplicationGroupResource{}
	vals := map[string]tftypes.Value{
		"config_id":   tftypes.NewValue(tftypes.String, "cfg1"),
		"id":          tftypes.NewValue(tftypes.String, "mag1"),
		"name":        tftypes.NewValue(tftypes.String, "test"),
		"description": tftypes.NewValue(tftypes.String, ""),
		"uid_header":  tftypes.NewValue(tftypes.String, ""),
		"grace":       tftypes.NewValue(tftypes.String, ""),
	}
	testUpdateWithNilClient(t, r, vals)
}

func TestMobileApplicationGroupResource_Delete_NilClient(t *testing.T) {
	r := &MobileApplicationGroupResource{}
	vals := map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"id":        tftypes.NewValue(tftypes.String, "mag1"),
	}
	testDeleteWithNilClient(t, r, vals)
}

func TestMobileApplicationGroupResource_Configure_InvalidType(t *testing.T) {
	r := &MobileApplicationGroupResource{}
	testConfigureWithInvalidType(t, r)
}

// --- Rate Limit Rule CRUD ---

func TestRateLimitRuleResource_Create_NilClient(t *testing.T) {
	r := &RateLimitRuleResource{}
	vals := map[string]tftypes.Value{
		"config_id":     tftypes.NewValue(tftypes.String, "cfg1"),
		"id":            tftypes.NewValue(tftypes.String, nil),
		"name":          tftypes.NewValue(tftypes.String, "test-rule"),
		"description":   tftypes.NewValue(tftypes.String, ""),
		"global":        tftypes.NewValue(tftypes.Bool, false),
		"active":        tftypes.NewValue(tftypes.Bool, true),
		"timeframe":     tftypes.NewValue(tftypes.Number, 60),
		"threshold":     tftypes.NewValue(tftypes.Number, 100),
		"ttl":           tftypes.NewValue(tftypes.Number, nil),
		"action":        tftypes.NewValue(tftypes.String, "action-monitor"),
		"is_action_ban": tftypes.NewValue(tftypes.Bool, false),
		"pairwith":      tftypes.NewValue(tftypes.String, `{"self":"self"}`),
	}
	testCreateWithNilClient(t, r, vals)
}

func TestRateLimitRuleResource_Read_NilClient(t *testing.T) {
	r := &RateLimitRuleResource{}
	vals := map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"id":        tftypes.NewValue(tftypes.String, "rl1"),
	}
	testReadWithNilClient(t, r, vals)
}

func TestRateLimitRuleResource_Update_NilClient(t *testing.T) {
	r := &RateLimitRuleResource{}
	vals := map[string]tftypes.Value{
		"config_id":     tftypes.NewValue(tftypes.String, "cfg1"),
		"id":            tftypes.NewValue(tftypes.String, "rl1"),
		"name":          tftypes.NewValue(tftypes.String, "test-rule"),
		"description":   tftypes.NewValue(tftypes.String, ""),
		"global":        tftypes.NewValue(tftypes.Bool, false),
		"active":        tftypes.NewValue(tftypes.Bool, true),
		"timeframe":     tftypes.NewValue(tftypes.Number, 60),
		"threshold":     tftypes.NewValue(tftypes.Number, 100),
		"ttl":           tftypes.NewValue(tftypes.Number, nil),
		"action":        tftypes.NewValue(tftypes.String, "action-monitor"),
		"is_action_ban": tftypes.NewValue(tftypes.Bool, false),
		"pairwith":      tftypes.NewValue(tftypes.String, `{"self":"self"}`),
	}
	testUpdateWithNilClient(t, r, vals)
}

func TestRateLimitRuleResource_Delete_NilClient(t *testing.T) {
	r := &RateLimitRuleResource{}
	vals := map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"id":        tftypes.NewValue(tftypes.String, "rl1"),
	}
	testDeleteWithNilClient(t, r, vals)
}

func TestRateLimitRuleResource_Configure_InvalidType(t *testing.T) {
	r := &RateLimitRuleResource{}
	testConfigureWithInvalidType(t, r)
}

// --- Security Policy CRUD ---

func TestSecurityPolicyResource_Create_NilClient(t *testing.T) {
	r := &SecurityPolicyResource{}
	vals := map[string]tftypes.Value{
		"config_id":   tftypes.NewValue(tftypes.String, "cfg1"),
		"id":          tftypes.NewValue(tftypes.String, nil),
		"name":        tftypes.NewValue(tftypes.String, "test-sp"),
		"description": tftypes.NewValue(tftypes.String, ""),
	}
	testCreateWithNilClient(t, r, vals)
}

func TestSecurityPolicyResource_Read_NilClient(t *testing.T) {
	r := &SecurityPolicyResource{}
	vals := map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"id":        tftypes.NewValue(tftypes.String, "sp1"),
	}
	testReadWithNilClient(t, r, vals)
}

func TestSecurityPolicyResource_Update_NilClient(t *testing.T) {
	r := &SecurityPolicyResource{}
	vals := map[string]tftypes.Value{
		"config_id":   tftypes.NewValue(tftypes.String, "cfg1"),
		"id":          tftypes.NewValue(tftypes.String, "sp1"),
		"name":        tftypes.NewValue(tftypes.String, "test-sp"),
		"description": tftypes.NewValue(tftypes.String, ""),
	}
	testUpdateWithNilClient(t, r, vals)
}

func TestSecurityPolicyResource_Delete_NilClient(t *testing.T) {
	r := &SecurityPolicyResource{}
	vals := map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"id":        tftypes.NewValue(tftypes.String, "sp1"),
	}
	testDeleteWithNilClient(t, r, vals)
}

func TestSecurityPolicyResource_Configure_InvalidType(t *testing.T) {
	r := &SecurityPolicyResource{}
	testConfigureWithInvalidType(t, r)
}

// --- Backend Service CRUD ---

func TestBackendServiceResource_Create_NilClient(t *testing.T) {
	r := &BackendServiceResource{}
	vals := map[string]tftypes.Value{
		"config_id":          tftypes.NewValue(tftypes.String, "cfg1"),
		"id":                 tftypes.NewValue(tftypes.String, nil),
		"name":               tftypes.NewValue(tftypes.String, "test-bs"),
		"description":        tftypes.NewValue(tftypes.String, ""),
		"http11":             tftypes.NewValue(tftypes.Bool, true),
		"transport_mode":     tftypes.NewValue(tftypes.String, "default"),
		"sticky":             tftypes.NewValue(tftypes.String, "none"),
		"sticky_cookie_name": tftypes.NewValue(tftypes.String, ""),
		"least_conn":         tftypes.NewValue(tftypes.Bool, false),
	}
	testCreateWithNilClient(t, r, vals)
}

func TestBackendServiceResource_Read_NilClient(t *testing.T) {
	r := &BackendServiceResource{}
	vals := map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"id":        tftypes.NewValue(tftypes.String, "bs1"),
	}
	testReadWithNilClient(t, r, vals)
}

func TestBackendServiceResource_Update_NilClient(t *testing.T) {
	r := &BackendServiceResource{}
	vals := map[string]tftypes.Value{
		"config_id":          tftypes.NewValue(tftypes.String, "cfg1"),
		"id":                 tftypes.NewValue(tftypes.String, "bs1"),
		"name":               tftypes.NewValue(tftypes.String, "test-bs"),
		"description":        tftypes.NewValue(tftypes.String, ""),
		"http11":             tftypes.NewValue(tftypes.Bool, true),
		"transport_mode":     tftypes.NewValue(tftypes.String, "default"),
		"sticky":             tftypes.NewValue(tftypes.String, "none"),
		"sticky_cookie_name": tftypes.NewValue(tftypes.String, ""),
		"least_conn":         tftypes.NewValue(tftypes.Bool, false),
	}
	testUpdateWithNilClient(t, r, vals)
}

func TestBackendServiceResource_Delete_NilClient(t *testing.T) {
	r := &BackendServiceResource{}
	vals := map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"id":        tftypes.NewValue(tftypes.String, "bs1"),
	}
	testDeleteWithNilClient(t, r, vals)
}

func TestBackendServiceResource_Configure_InvalidType(t *testing.T) {
	r := &BackendServiceResource{}
	testConfigureWithInvalidType(t, r)
}

// --- Publish Configure ---

func TestPublishResource_Configure_InvalidType(t *testing.T) {
	r := &PublishResource{}
	testConfigureWithInvalidType(t, r)
}
