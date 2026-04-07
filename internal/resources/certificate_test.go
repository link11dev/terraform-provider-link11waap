package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/link11/terraform-provider-link11waap/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCertificateResource(t *testing.T) {
	r := NewCertificateResource()
	require.NotNil(t, r)
	_, ok := r.(*CertificateResource)
	assert.True(t, ok)
}

func TestCertificateResource_Metadata(t *testing.T) {
	r := &CertificateResource{}
	ctx := context.Background()

	req := metadataReq("link11waap")
	resp := metadataResp()
	r.Metadata(ctx, req, resp)

	assert.Equal(t, "link11waap_certificate", resp.TypeName)
}

func TestCertificateResource_Schema(t *testing.T) {
	r := &CertificateResource{}
	ctx := context.Background()

	req := schemaReq()
	resp := schemaResp()
	r.Schema(ctx, req, resp)

	schema := resp.Schema
	assert.NotEmpty(t, schema.Attributes)

	expectedAttrs := []string{
		"config_id", "id", "cert_body", "private_key", "domains",
		"le_auto_renew", "le_auto_replace", "side",
		"name", "subject", "issuer", "san", "expires", "uploaded",
		"revoked", "links",
	}
	for _, attr := range expectedAttrs {
		_, ok := schema.Attributes[attr]
		assert.True(t, ok, "expected attribute %q in schema", attr)
	}
}

func TestCertificateResource_Configure_NilProvider(t *testing.T) {
	r := &CertificateResource{}
	ctx := context.Background()

	req := configureReq(nil)
	resp := configureResp()
	r.Configure(ctx, req, resp)

	assert.Nil(t, r.client)
	assert.False(t, resp.Diagnostics.HasError())
}

func TestCertificateResource_ImportState_Valid(t *testing.T) {
	r := &CertificateResource{}
	resp := testImportState(t, r, "config123/cert456")

	assert.False(t, resp.Diagnostics.HasError())
}

func TestCertificateResource_ImportState_Invalid(t *testing.T) {
	r := &CertificateResource{}
	resp := testImportState(t, r, "invalid")

	assert.True(t, resp.Diagnostics.HasError())
}

func TestCertificateResource_ImportState_TooManyParts(t *testing.T) {
	r := &CertificateResource{}
	resp := testImportState(t, r, "a/b/c")

	assert.True(t, resp.Diagnostics.HasError())
}

func TestCertificateResource_UpdateModelFromAPI_BasicFields(t *testing.T) {
	r := &CertificateResource{}
	ctx := context.Background()
	model := &CertificateResourceModel{}
	var diags diag.Diagnostics

	cert := &client.Certificate{
		Name:          "test-cert",
		Subject:       "CN=test.com",
		Issuer:        "CN=Issuer",
		Expires:       "2025-12-31",
		Uploaded:      "2024-01-01",
		Revoked:       false,
		LEAutoRenew:   true,
		LEAutoReplace: false,
		Side:          "server",
		SAN:           []string{"test.com", "www.test.com"},
	}

	r.updateModelFromAPI(ctx, model, cert, &diags)

	assert.False(t, diags.HasError())
	assert.Equal(t, "test-cert", model.Name.ValueString())
	assert.Equal(t, "CN=test.com", model.Subject.ValueString())
	assert.Equal(t, "CN=Issuer", model.Issuer.ValueString())
	assert.Equal(t, "2025-12-31", model.Expires.ValueString())
	assert.Equal(t, "2024-01-01", model.Uploaded.ValueString())
	assert.False(t, model.Revoked.ValueBool())
	assert.True(t, model.LEAutoRenew.ValueBool())
	assert.False(t, model.LEAutoReplace.ValueBool())
	assert.Equal(t, "server", model.Side.ValueString())
	assert.False(t, model.SAN.IsNull())
}

func TestCertificateResource_UpdateModelFromAPI_NilSAN(t *testing.T) {
	r := &CertificateResource{}
	ctx := context.Background()
	model := &CertificateResourceModel{}
	var diags diag.Diagnostics

	cert := &client.Certificate{
		Name: "no-san",
		SAN:  nil,
	}

	r.updateModelFromAPI(ctx, model, cert, &diags)

	assert.False(t, diags.HasError())
	assert.True(t, model.SAN.IsNull())
}

func TestCertificateResource_UpdateModelFromAPI_EmptySide(t *testing.T) {
	r := &CertificateResource{}
	ctx := context.Background()
	model := &CertificateResourceModel{
		Side: types.StringValue("existing"),
	}
	var diags diag.Diagnostics

	cert := &client.Certificate{
		Name: "no-side",
		Side: "",
	}

	r.updateModelFromAPI(ctx, model, cert, &diags)

	assert.False(t, diags.HasError())
	// Side should remain as the existing value since cert.Side is empty
	assert.Equal(t, "existing", model.Side.ValueString())
}

func TestCertificateResource_UpdateModelFromAPI_WithLinks(t *testing.T) {
	r := &CertificateResource{}
	ctx := context.Background()
	model := &CertificateResourceModel{}
	var diags diag.Diagnostics

	cert := &client.Certificate{
		Name: "with-links",
		Links: []client.ProviderLink{
			{Provider: "aws", Link: "arn:aws:acm:us-east-1:123:certificate/abc", Region: "us-east-1"},
		},
	}

	r.updateModelFromAPI(ctx, model, cert, &diags)

	assert.False(t, diags.HasError())
	assert.False(t, model.Links.IsNull())
}

func TestCertificateResource_UpdateModelFromAPI_WithProviderLinks(t *testing.T) {
	r := &CertificateResource{}
	ctx := context.Background()
	model := &CertificateResourceModel{}
	var diags diag.Diagnostics

	cert := &client.Certificate{
		Name: "with-provider-links",
		ProviderLinks: []client.ProviderLink{
			{Provider: "gcp", Link: "projects/123/certs/abc", Region: "us-central1"},
		},
	}

	r.updateModelFromAPI(ctx, model, cert, &diags)

	assert.False(t, diags.HasError())
	assert.False(t, model.Links.IsNull())
}

func TestCertificateResource_UpdateModelFromAPI_NoLinks(t *testing.T) {
	r := &CertificateResource{}
	ctx := context.Background()
	model := &CertificateResourceModel{}
	var diags diag.Diagnostics

	cert := &client.Certificate{
		Name: "no-links",
	}

	r.updateModelFromAPI(ctx, model, cert, &diags)

	assert.False(t, diags.HasError())
	assert.True(t, model.Links.IsNull())
}

func TestCertificateResourceModel_Fields(t *testing.T) {
	model := CertificateResourceModel{
		ConfigID:      types.StringValue("cfg1"),
		ID:            types.StringValue("cert1"),
		CertBody:      types.StringValue("-----BEGIN CERTIFICATE-----"),
		PrivateKey:    types.StringValue("-----BEGIN PRIVATE KEY-----"),
		LEAutoRenew:   types.BoolValue(true),
		LEAutoReplace: types.BoolValue(false),
		Side:          types.StringValue("server"),
	}

	assert.Equal(t, "cfg1", model.ConfigID.ValueString())
	assert.Equal(t, "cert1", model.ID.ValueString())
	assert.True(t, model.LEAutoRenew.ValueBool())
	assert.False(t, model.LEAutoReplace.ValueBool())
}
