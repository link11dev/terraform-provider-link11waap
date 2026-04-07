package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewServerGroupResource(t *testing.T) {
	r := NewServerGroupResource()
	require.NotNil(t, r)
	_, ok := r.(*ServerGroupResource)
	assert.True(t, ok)
}

func TestServerGroupResource_Metadata(t *testing.T) {
	r := &ServerGroupResource{}
	ctx := context.Background()

	req := metadataReq("link11waap")
	resp := metadataResp()
	r.Metadata(ctx, req, resp)

	assert.Equal(t, "link11waap_server_group", resp.TypeName)
}

func TestServerGroupResource_Schema(t *testing.T) {
	r := &ServerGroupResource{}
	ctx := context.Background()

	req := schemaReq()
	resp := schemaResp()
	r.Schema(ctx, req, resp)

	schema := resp.Schema
	assert.NotEmpty(t, schema.Attributes)

	expectedAttrs := []string{
		"config_id", "id", "name", "description", "server_names",
		"security_policy", "routing_profile", "proxy_template",
		"challenge_cookie_domain", "ssl_certificate", "client_certificate",
		"client_certificate_mode", "mobile_application_group",
	}
	for _, attr := range expectedAttrs {
		_, ok := schema.Attributes[attr]
		assert.True(t, ok, "expected attribute %q in schema", attr)
	}
}

func TestServerGroupResource_Configure_NilProvider(t *testing.T) {
	r := &ServerGroupResource{}
	ctx := context.Background()

	req := configureReq(nil)
	resp := configureResp()
	r.Configure(ctx, req, resp)

	assert.Nil(t, r.client)
	assert.False(t, resp.Diagnostics.HasError())
}

func TestServerGroupResource_ImportState_Valid(t *testing.T) {
	r := &ServerGroupResource{}
	resp := testImportState(t, r, "config123/sg456")

	assert.False(t, resp.Diagnostics.HasError())
}

func TestServerGroupResource_ImportState_Invalid(t *testing.T) {
	r := &ServerGroupResource{}
	resp := testImportState(t, r, "invalid")

	assert.True(t, resp.Diagnostics.HasError())
}

func TestServerGroupResource_ImportState_TooManyParts(t *testing.T) {
	r := &ServerGroupResource{}
	resp := testImportState(t, r, "a/b/c")

	assert.True(t, resp.Diagnostics.HasError())
}

func TestServerGroupResourceModel_Fields(t *testing.T) {
	model := ServerGroupResourceModel{
		ConfigID:               types.StringValue("cfg1"),
		ID:                     types.StringValue("sg1"),
		Name:                   types.StringValue("test-site"),
		Description:            types.StringValue("desc"),
		SecurityPolicy:         types.StringValue("sp-1"),
		RoutingProfile:         types.StringValue("rp-1"),
		ProxyTemplate:          types.StringValue("pt-1"),
		ChallengeCookieDomain:  types.StringValue("example.com"),
		SSLCertificate:         types.StringValue("cert-1"),
		ClientCertificate:      types.StringValue("client-cert-1"),
		ClientCertificateMode:  types.StringValue("on"),
		MobileApplicationGroup: types.StringValue("mag-1"),
	}

	assert.Equal(t, "cfg1", model.ConfigID.ValueString())
	assert.Equal(t, "sg1", model.ID.ValueString())
	assert.Equal(t, "test-site", model.Name.ValueString())
	assert.Equal(t, "sp-1", model.SecurityPolicy.ValueString())
	assert.Equal(t, "rp-1", model.RoutingProfile.ValueString())
	assert.Equal(t, "pt-1", model.ProxyTemplate.ValueString())
	assert.Equal(t, "example.com", model.ChallengeCookieDomain.ValueString())
	assert.Equal(t, "cert-1", model.SSLCertificate.ValueString())
	assert.Equal(t, "client-cert-1", model.ClientCertificate.ValueString())
	assert.Equal(t, "on", model.ClientCertificateMode.ValueString())
	assert.Equal(t, "mag-1", model.MobileApplicationGroup.ValueString())
}
