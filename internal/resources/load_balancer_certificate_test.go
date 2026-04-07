package resources

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLoadBalancerCertificateResource(t *testing.T) {
	r := NewLoadBalancerCertificateResource()
	require.NotNil(t, r)
	_, ok := r.(*LoadBalancerCertificateResource)
	assert.True(t, ok)
}

func TestLoadBalancerCertificateResource_Metadata(t *testing.T) {
	r := &LoadBalancerCertificateResource{}
	ctx := context.Background()

	req := metadataReq("link11waap")
	resp := metadataResp()
	r.Metadata(ctx, req, resp)

	assert.Equal(t, "link11waap_load_balancer_certificate", resp.TypeName)
}

func TestLoadBalancerCertificateResource_Schema(t *testing.T) {
	r := &LoadBalancerCertificateResource{}
	ctx := context.Background()

	req := schemaReq()
	resp := schemaResp()
	r.Schema(ctx, req, resp)

	schema := resp.Schema
	assert.NotEmpty(t, schema.Attributes)

	expectedAttrs := []string{
		"id", "config_id", "load_balancer_name", "certificate_id",
		"provider_type", "region", "listener", "listener_port",
		"is_default", "elbv2",
	}
	for _, attr := range expectedAttrs {
		_, ok := schema.Attributes[attr]
		assert.True(t, ok, "expected attribute %q in schema", attr)
	}
}

func TestLoadBalancerCertificateResource_Configure_NilProvider(t *testing.T) {
	r := &LoadBalancerCertificateResource{}
	ctx := context.Background()

	req := configureReq(nil)
	resp := configureResp()
	r.Configure(ctx, req, resp)

	assert.Nil(t, r.client)
	assert.False(t, resp.Diagnostics.HasError())
}

func TestLoadBalancerCertificateResource_ImportState_Valid(t *testing.T) {
	r := &LoadBalancerCertificateResource{}
	resp := testImportState(t, r, "config123/lb-name/cert456")

	assert.False(t, resp.Diagnostics.HasError())
}

func TestLoadBalancerCertificateResource_ImportState_Invalid_TwoParts(t *testing.T) {
	r := &LoadBalancerCertificateResource{}
	resp := testImportState(t, r, "config123/lb-name")

	assert.True(t, resp.Diagnostics.HasError())
}

func TestLoadBalancerCertificateResource_ImportState_Invalid_OnePart(t *testing.T) {
	r := &LoadBalancerCertificateResource{}
	resp := testImportState(t, r, "invalid")

	assert.True(t, resp.Diagnostics.HasError())
}

func TestLoadBalancerCertificateResource_ImportState_Invalid_FourParts(t *testing.T) {
	r := &LoadBalancerCertificateResource{}
	resp := testImportState(t, r, "a/b/c/d")

	assert.True(t, resp.Diagnostics.HasError())
}
