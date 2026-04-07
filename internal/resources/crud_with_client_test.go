package resources

import (
	"context"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/link11/terraform-provider-link11waap/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newFailingClient creates a client that will always fail API calls
// (points to an unreachable address with minimal timeout).
func newFailingClient(t *testing.T) *client.Client {
	t.Helper()
	c, err := client.New(client.Config{
		Domain:    "localhost:1", // port 1 is typically unreachable
		APIKey:    "dGVzdC1rZXk=",
		Timeout:   100 * time.Millisecond,
		RetryMax:  0,
		RetryWait: time.Millisecond,
	})
	require.NoError(t, err)
	return c
}

// configureResource sets up a resource with a failing client via Configure.
func configureResource(t *testing.T, r resource.Resource) {
	t.Helper()
	ctx := context.Background()
	c := newFailingClient(t)

	req := resource.ConfigureRequest{ProviderData: c}
	resp := &resource.ConfigureResponse{}
	// Call Configure using type assertion since not all resources implement the same interface
	if cr, ok := r.(interface {
		Configure(context.Context, resource.ConfigureRequest, *resource.ConfigureResponse)
	}); ok {
		cr.Configure(ctx, req, resp)
	}
	assert.False(t, resp.Diagnostics.HasError())
}

// crudCreateWithClient tests Create with a real (but failing) client.
func crudCreateWithClient(t *testing.T, r resource.Resource, planValues map[string]tftypes.Value) {
	t.Helper()
	ctx := context.Background()
	configureResource(t, r)

	plan := buildTerraformPlan(ctx, t, r, planValues)

	sReq := schemaReq()
	sResp := schemaResp()
	r.Schema(ctx, sReq, sResp)

	emptyState := tfsdk.State{
		Schema: sResp.Schema,
		Raw:    tftypes.NewValue(sResp.Schema.Type().TerraformType(ctx), nil),
	}

	createReq := resource.CreateRequest{Plan: plan}
	createResp := &resource.CreateResponse{State: emptyState}

	r.Create(ctx, createReq, createResp)

	// Should have error because the client cannot connect
	assert.True(t, createResp.Diagnostics.HasError(), "expected error from failing client")
}

// crudReadWithClient tests Read with a real (but failing) client.
func crudReadWithClient(t *testing.T, r resource.Resource, stateValues map[string]tftypes.Value) {
	t.Helper()
	ctx := context.Background()
	configureResource(t, r)

	state := buildTerraformState(ctx, t, r, stateValues)

	sReq := schemaReq()
	sResp := schemaResp()
	r.Schema(ctx, sReq, sResp)

	readReq := resource.ReadRequest{State: state}
	readResp := &resource.ReadResponse{
		State: tfsdk.State{
			Schema: sResp.Schema,
			Raw:    state.Raw.Copy(),
		},
	}

	r.Read(ctx, readReq, readResp)

	// Should have error because the client cannot connect
	assert.True(t, readResp.Diagnostics.HasError(), "expected error from failing client")
}

// crudUpdateWithClient tests Update with a real (but failing) client.
func crudUpdateWithClient(t *testing.T, r resource.Resource, planValues map[string]tftypes.Value) {
	t.Helper()
	ctx := context.Background()
	configureResource(t, r)

	plan := buildTerraformPlan(ctx, t, r, planValues)

	sReq := schemaReq()
	sResp := schemaResp()
	r.Schema(ctx, sReq, sResp)

	emptyState := tfsdk.State{
		Schema: sResp.Schema,
		Raw:    tftypes.NewValue(sResp.Schema.Type().TerraformType(ctx), nil),
	}

	updateReq := resource.UpdateRequest{Plan: plan}
	updateResp := &resource.UpdateResponse{State: emptyState}

	r.Update(ctx, updateReq, updateResp)

	assert.True(t, updateResp.Diagnostics.HasError(), "expected error from failing client")
}

// crudDeleteWithClient tests Delete with a real (but failing) client.
func crudDeleteWithClient(t *testing.T, r resource.Resource, stateValues map[string]tftypes.Value) {
	t.Helper()
	ctx := context.Background()
	configureResource(t, r)

	state := buildTerraformState(ctx, t, r, stateValues)

	deleteReq := resource.DeleteRequest{State: state}
	deleteResp := &resource.DeleteResponse{}

	r.Delete(ctx, deleteReq, deleteResp)

	assert.True(t, deleteResp.Diagnostics.HasError(), "expected error from failing client")
}

// --- ACL Profile with client ---

func TestACLProfileResource_CRUD_WithFailingClient(t *testing.T) {
	r := &ACLProfileResource{}
	planVals := map[string]tftypes.Value{
		"config_id":   tftypes.NewValue(tftypes.String, "cfg1"),
		"id":          tftypes.NewValue(tftypes.String, nil),
		"name":        tftypes.NewValue(tftypes.String, "test"),
		"description": tftypes.NewValue(tftypes.String, ""),
		"action":      tftypes.NewValue(tftypes.String, nil),
	}
	stateVals := map[string]tftypes.Value{
		"config_id":   tftypes.NewValue(tftypes.String, "cfg1"),
		"id":          tftypes.NewValue(tftypes.String, "acl1"),
		"name":        tftypes.NewValue(tftypes.String, "test"),
		"description": tftypes.NewValue(tftypes.String, ""),
	}

	t.Run("Create", func(t *testing.T) { crudCreateWithClient(t, r, planVals) })
	t.Run("Read", func(t *testing.T) { crudReadWithClient(t, r, stateVals) })
	t.Run("Update", func(t *testing.T) { crudUpdateWithClient(t, r, planVals) })
	t.Run("Delete", func(t *testing.T) { crudDeleteWithClient(t, r, stateVals) })
}

// --- Edge Function with client ---

func TestEdgeFunctionResource_CRUD_WithFailingClient(t *testing.T) {
	r := &EdgeFunctionResource{}
	vals := map[string]tftypes.Value{
		"config_id":   tftypes.NewValue(tftypes.String, "cfg1"),
		"id":          tftypes.NewValue(tftypes.String, "ef1"),
		"name":        tftypes.NewValue(tftypes.String, "test-ef"),
		"description": tftypes.NewValue(tftypes.String, ""),
		"code":        tftypes.NewValue(tftypes.String, "print('hello')"),
		"phase":       tftypes.NewValue(tftypes.String, "request_post"),
	}
	createVals := map[string]tftypes.Value{
		"config_id":   tftypes.NewValue(tftypes.String, "cfg1"),
		"id":          tftypes.NewValue(tftypes.String, nil),
		"name":        tftypes.NewValue(tftypes.String, "test-ef"),
		"description": tftypes.NewValue(tftypes.String, ""),
		"code":        tftypes.NewValue(tftypes.String, "print('hello')"),
		"phase":       tftypes.NewValue(tftypes.String, "request_post"),
	}

	t.Run("Create", func(t *testing.T) { crudCreateWithClient(t, r, createVals) })
	t.Run("Read", func(t *testing.T) { crudReadWithClient(t, r, vals) })
	t.Run("Update", func(t *testing.T) { crudUpdateWithClient(t, r, vals) })
	t.Run("Delete", func(t *testing.T) { crudDeleteWithClient(t, r, vals) })
}

// --- Certificate with client ---

func TestCertificateResource_CRUD_WithFailingClient(t *testing.T) {
	r := &CertificateResource{}
	createVals := map[string]tftypes.Value{
		"config_id":       tftypes.NewValue(tftypes.String, "cfg1"),
		"id":              tftypes.NewValue(tftypes.String, nil),
		"cert_body":       tftypes.NewValue(tftypes.String, "cert-body"),
		"private_key":     tftypes.NewValue(tftypes.String, "key"),
		"le_auto_renew":   tftypes.NewValue(tftypes.Bool, false),
		"le_auto_replace": tftypes.NewValue(tftypes.Bool, false),
		"side":            tftypes.NewValue(tftypes.String, "server"),
	}
	stateVals := map[string]tftypes.Value{
		"config_id":       tftypes.NewValue(tftypes.String, "cfg1"),
		"id":              tftypes.NewValue(tftypes.String, "cert1"),
		"le_auto_renew":   tftypes.NewValue(tftypes.Bool, false),
		"le_auto_replace": tftypes.NewValue(tftypes.Bool, false),
	}

	t.Run("Create", func(t *testing.T) { crudCreateWithClient(t, r, createVals) })
	t.Run("Read", func(t *testing.T) { crudReadWithClient(t, r, stateVals) })
	t.Run("Update", func(t *testing.T) { crudUpdateWithClient(t, r, stateVals) })
	t.Run("Delete", func(t *testing.T) { crudDeleteWithClient(t, r, stateVals) })
}

// --- Server Group with client ---

func TestServerGroupResource_CRUD_WithFailingClient(t *testing.T) {
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
	createVals := make(map[string]tftypes.Value)
	for k, v := range vals {
		createVals[k] = v
	}
	createVals["id"] = tftypes.NewValue(tftypes.String, nil)

	t.Run("Create", func(t *testing.T) { crudCreateWithClient(t, r, createVals) })
	t.Run("Read", func(t *testing.T) { crudReadWithClient(t, r, vals) })
	t.Run("Update", func(t *testing.T) { crudUpdateWithClient(t, r, vals) })
	t.Run("Delete", func(t *testing.T) { crudDeleteWithClient(t, r, vals) })
}

// --- User with client ---

func TestUserResource_CRUD_WithFailingClient(t *testing.T) {
	r := &UserResource{}
	vals := map[string]tftypes.Value{
		"id":           tftypes.NewValue(tftypes.String, "user1"),
		"acl":          tftypes.NewValue(tftypes.Number, 10),
		"contact_name": tftypes.NewValue(tftypes.String, "John"),
		"email":        tftypes.NewValue(tftypes.String, "john@test.com"),
		"mobile":       tftypes.NewValue(tftypes.String, "+1234567890"),
		"org_id":       tftypes.NewValue(tftypes.String, "org1"),
		"org_name":     tftypes.NewValue(tftypes.String, "Org"),
	}
	createVals := make(map[string]tftypes.Value)
	for k, v := range vals {
		createVals[k] = v
	}
	createVals["id"] = tftypes.NewValue(tftypes.String, nil)
	createVals["org_name"] = tftypes.NewValue(tftypes.String, nil)

	t.Run("Create", func(t *testing.T) { crudCreateWithClient(t, r, createVals) })
	t.Run("Read", func(t *testing.T) { crudReadWithClient(t, r, vals) })
	t.Run("Update", func(t *testing.T) { crudUpdateWithClient(t, r, vals) })
	t.Run("Delete", func(t *testing.T) { crudDeleteWithClient(t, r, vals) })
}

// --- Publish with client ---

func TestPublishResource_CRUD_WithFailingClient(t *testing.T) {
	r := &PublishResource{}
	vals := map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"id":        tftypes.NewValue(tftypes.String, "cfg1"),
	}

	t.Run("Create", func(t *testing.T) { crudCreateWithClient(t, r, vals) })
	t.Run("Update", func(t *testing.T) { crudUpdateWithClient(t, r, vals) })
}

// --- Mobile Application Group with client ---

func TestMobileApplicationGroupResource_CRUD_WithFailingClient(t *testing.T) {
	r := &MobileApplicationGroupResource{}
	vals := map[string]tftypes.Value{
		"config_id":   tftypes.NewValue(tftypes.String, "cfg1"),
		"id":          tftypes.NewValue(tftypes.String, "mag1"),
		"name":        tftypes.NewValue(tftypes.String, "test"),
		"description": tftypes.NewValue(tftypes.String, ""),
		"uid_header":  tftypes.NewValue(tftypes.String, ""),
		"grace":       tftypes.NewValue(tftypes.String, ""),
	}
	createVals := make(map[string]tftypes.Value)
	for k, v := range vals {
		createVals[k] = v
	}
	createVals["id"] = tftypes.NewValue(tftypes.String, nil)

	t.Run("Create", func(t *testing.T) { crudCreateWithClient(t, r, createVals) })
	t.Run("Read", func(t *testing.T) { crudReadWithClient(t, r, vals) })
	t.Run("Update", func(t *testing.T) { crudUpdateWithClient(t, r, vals) })
	t.Run("Delete", func(t *testing.T) { crudDeleteWithClient(t, r, vals) })
}

// --- Load Balancer Certificate with client ---

func TestLoadBalancerCertificateResource_CRUD_WithFailingClient(t *testing.T) {
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
	createVals := make(map[string]tftypes.Value)
	for k, v := range vals {
		createVals[k] = v
	}
	createVals["id"] = tftypes.NewValue(tftypes.String, nil)

	t.Run("Create", func(t *testing.T) { crudCreateWithClient(t, r, createVals) })
	t.Run("Read", func(t *testing.T) { crudReadWithClient(t, r, vals) })
	t.Run("Delete", func(t *testing.T) { crudDeleteWithClient(t, r, vals) })
}

// --- Load Balancer Regions with client ---

func TestLoadBalancerRegionsResource_CRUD_WithFailingClient(t *testing.T) {
	r := &LoadBalancerRegionsResource{}
	vals := map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"lb_id":     tftypes.NewValue(tftypes.String, "lb1"),
		"regions":   tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, map[string]tftypes.Value{"ams": tftypes.NewValue(tftypes.String, "automatic")}),
	}

	t.Run("Create", func(t *testing.T) { crudCreateWithClient(t, r, vals) })
	t.Run("Read", func(t *testing.T) { crudReadWithClient(t, r, vals) })
	t.Run("Update", func(t *testing.T) { crudUpdateWithClient(t, r, vals) })
}

// --- Backend Service with client ---

func TestBackendServiceResource_CRUD_WithFailingClient(t *testing.T) {
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
	createVals := make(map[string]tftypes.Value)
	for k, v := range vals {
		createVals[k] = v
	}
	createVals["id"] = tftypes.NewValue(tftypes.String, nil)

	t.Run("Create", func(t *testing.T) { crudCreateWithClient(t, r, createVals) })
	t.Run("Read", func(t *testing.T) { crudReadWithClient(t, r, vals) })
	t.Run("Update", func(t *testing.T) { crudUpdateWithClient(t, r, vals) })
	t.Run("Delete", func(t *testing.T) { crudDeleteWithClient(t, r, vals) })
}

// --- Security Policy with client ---

func TestSecurityPolicyResource_CRUD_WithFailingClient(t *testing.T) {
	r := &SecurityPolicyResource{}
	vals := map[string]tftypes.Value{
		"config_id":   tftypes.NewValue(tftypes.String, "cfg1"),
		"id":          tftypes.NewValue(tftypes.String, "sp1"),
		"name":        tftypes.NewValue(tftypes.String, "test-sp"),
		"description": tftypes.NewValue(tftypes.String, ""),
	}
	createVals := make(map[string]tftypes.Value)
	for k, v := range vals {
		createVals[k] = v
	}
	createVals["id"] = tftypes.NewValue(tftypes.String, nil)

	t.Run("Create", func(t *testing.T) { crudCreateWithClient(t, r, createVals) })
	t.Run("Read", func(t *testing.T) { crudReadWithClient(t, r, vals) })
	t.Run("Update", func(t *testing.T) { crudUpdateWithClient(t, r, vals) })
	t.Run("Delete", func(t *testing.T) { crudDeleteWithClient(t, r, vals) })
}

// --- Rate Limit Rule with client ---

func TestRateLimitRuleResource_CRUD_WithFailingClient(t *testing.T) {
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
	createVals := make(map[string]tftypes.Value)
	for k, v := range vals {
		createVals[k] = v
	}
	createVals["id"] = tftypes.NewValue(tftypes.String, nil)

	t.Run("Create", func(t *testing.T) { crudCreateWithClient(t, r, createVals) })
	t.Run("Read", func(t *testing.T) { crudReadWithClient(t, r, vals) })
	t.Run("Update", func(t *testing.T) { crudUpdateWithClient(t, r, vals) })
	t.Run("Delete", func(t *testing.T) { crudDeleteWithClient(t, r, vals) })
}
