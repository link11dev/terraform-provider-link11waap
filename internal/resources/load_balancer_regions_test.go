package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLoadBalancerRegionsResource(t *testing.T) {
	r := NewLoadBalancerRegionsResource()
	require.NotNil(t, r)
	_, ok := r.(*LoadBalancerRegionsResource)
	assert.True(t, ok)
}

func TestLoadBalancerRegionsResource_Metadata(t *testing.T) {
	r := &LoadBalancerRegionsResource{}
	ctx := context.Background()

	req := metadataReq("link11waap")
	resp := metadataResp()
	r.Metadata(ctx, req, resp)

	assert.Equal(t, "link11waap_load_balancer_regions", resp.TypeName)
}

func TestLoadBalancerRegionsResource_Schema(t *testing.T) {
	r := &LoadBalancerRegionsResource{}
	ctx := context.Background()

	req := schemaReq()
	resp := schemaResp()
	r.Schema(ctx, req, resp)

	schema := resp.Schema
	assert.NotEmpty(t, schema.Attributes)

	expectedAttrs := []string{
		"config_id", "lb_id", "regions", "name", "upstream_regions",
	}
	for _, attr := range expectedAttrs {
		_, ok := schema.Attributes[attr]
		assert.True(t, ok, "expected attribute %q in schema", attr)
	}
}

func TestLoadBalancerRegionsResource_Configure_NilProvider(t *testing.T) {
	r := &LoadBalancerRegionsResource{}
	ctx := context.Background()

	req := configureReq(nil)
	resp := configureResp()
	r.Configure(ctx, req, resp)

	assert.Nil(t, r.client)
	assert.False(t, resp.Diagnostics.HasError())
}

func TestLoadBalancerRegionsResource_ImportState_Valid(t *testing.T) {
	r := &LoadBalancerRegionsResource{}
	resp := testImportState(t, r, "config123/lb456")

	assert.False(t, resp.Diagnostics.HasError())
}

func TestLoadBalancerRegionsResource_ImportState_Invalid(t *testing.T) {
	r := &LoadBalancerRegionsResource{}
	resp := testImportState(t, r, "invalid")

	assert.True(t, resp.Diagnostics.HasError())
}

func TestLoadBalancerRegionsResource_ImportState_TooManyParts(t *testing.T) {
	r := &LoadBalancerRegionsResource{}
	resp := testImportState(t, r, "a/b/c")

	assert.True(t, resp.Diagnostics.HasError())
}

func TestKnownRegions_Contains_ExpectedRegions(t *testing.T) {
	expected := []string{"ams", "ash", "ffm", "hkg", "lax", "lon", "nyc", "sgp", "stl"}
	assert.Equal(t, expected, knownRegions)
}

func TestRegionsSchemaAttribute_HasOneEntryPerKnownRegionWithDefaults(t *testing.T) {
	r := &LoadBalancerRegionsResource{}
	ctx := context.Background()

	req := schemaReq()
	resp := schemaResp()
	r.Schema(ctx, req, resp)

	regionsAttr, ok := resp.Schema.Attributes["regions"].(schema.SingleNestedAttribute)
	require.True(t, ok, "expected regions to be a SingleNestedAttribute")
	require.Len(t, regionsAttr.Attributes, len(knownRegions))

	for _, region := range knownRegions {
		sub, ok := regionsAttr.Attributes[region].(schema.StringAttribute)
		require.True(t, ok, "expected region %q to be a StringAttribute", region)
		assert.True(t, sub.Optional, "region %q should be optional", region)
		assert.True(t, sub.Computed, "region %q should be computed", region)
		require.NotNil(t, sub.Default, "region %q should have a default", region)
	}

	// Unknown city codes must not be accepted as valid attributes.
	_, ok = regionsAttr.Attributes["sfo"]
	assert.False(t, ok, "unexpected region attribute for unknown city code")
}

func TestRegionsObjectToMap(t *testing.T) {
	values := make(map[string]attr.Value, len(knownRegions))
	for _, region := range knownRegions {
		values[region] = types.StringValue("custom-" + region)
	}
	obj, diags := types.ObjectValue(regionAttributeTypes(), values)
	require.False(t, diags.HasError())

	result := regionsObjectToMap(obj)
	require.Len(t, result, len(knownRegions))
	for _, region := range knownRegions {
		assert.Equal(t, "custom-"+region, result[region])
	}
}

func TestRegionsMapToObject_FillsMissingRegionsWithAutomatic(t *testing.T) {
	// Only provide two regions from the "API".
	apiRegions := map[string]string{
		"ams": "ffm",
		"ffm": "custom",
	}

	obj, diags := regionsMapToObject(apiRegions)
	require.False(t, diags.HasError())

	result := regionsObjectToMap(obj)
	require.Len(t, result, len(knownRegions))
	assert.Equal(t, "ffm", result["ams"])
	assert.Equal(t, "custom", result["ffm"])

	for _, region := range knownRegions {
		if region != "ams" && region != "ffm" {
			assert.Equal(t, automaticRegionValue, result[region], "region %q should default to automatic", region)
		}
	}
}

func TestDefaultRegionsObject_AllRegionsAutomatic(t *testing.T) {
	obj := defaultRegionsObject()

	result := regionsObjectToMap(obj)
	require.Len(t, result, len(knownRegions))
	for _, region := range knownRegions {
		assert.Equal(t, automaticRegionValue, result[region])
	}
}
