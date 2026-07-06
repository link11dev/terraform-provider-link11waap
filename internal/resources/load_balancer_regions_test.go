package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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

func TestRegionsSchemaAttribute_IsMapWithKeyAndValueValidators(t *testing.T) {
	r := &LoadBalancerRegionsResource{}
	ctx := context.Background()

	req := schemaReq()
	resp := schemaResp()
	r.Schema(ctx, req, resp)

	regionsAttr, ok := resp.Schema.Attributes["regions"].(schema.MapAttribute)
	require.True(t, ok, "expected regions to be a MapAttribute")
	assert.True(t, regionsAttr.Optional, "regions should be optional")
	assert.False(t, regionsAttr.Computed, "regions should not be computed")
	assert.Equal(t, types.StringType, regionsAttr.ElementType)
	assert.NotEmpty(t, regionsAttr.Validators, "regions should have validators to reject unknown city codes")
}

// TestRegionsSchemaAttribute_RejectsUnknownCityCode is a regression test
// for a bug where a typo'd city code (e.g. "lon111" instead of "lon") was
// silently dropped instead of producing a validation error. This exercises
// the schema's Map validators the same way Terraform Core's
// ValidateResourceConfig RPC would.
func TestRegionsSchemaAttribute_RejectsUnknownCityCode(t *testing.T) {
	r := &LoadBalancerRegionsResource{}
	ctx := context.Background()

	req := schemaReq()
	resp := schemaResp()
	r.Schema(ctx, req, resp)

	regionsAttr, ok := resp.Schema.Attributes["regions"].(schema.MapAttribute)
	require.True(t, ok, "expected regions to be a MapAttribute")

	configValue, diags := types.MapValueFrom(ctx, types.StringType, map[string]string{
		"ams":    "ffm",
		"lon111": "automatic",
	})
	require.False(t, diags.HasError())

	var respDiags diag.Diagnostics
	for _, v := range regionsAttr.Validators {
		validateResp := &validator.MapResponse{}
		v.ValidateMap(ctx, validator.MapRequest{
			Path:        path.Root("regions"),
			ConfigValue: configValue,
		}, validateResp)
		respDiags.Append(validateResp.Diagnostics...)
	}

	assert.True(t, respDiags.HasError(), "expected an error for the unknown city code %q", "lon111")
}

func TestRegionsMapToStringMap(t *testing.T) {
	ctx := context.Background()

	m, diags := types.MapValueFrom(ctx, types.StringType, map[string]string{
		"ams": "ffm",
		"lon": "automatic",
	})
	require.False(t, diags.HasError())

	result, diags := regionsMapToStringMap(ctx, m)
	require.False(t, diags.HasError())
	assert.Equal(t, map[string]string{"ams": "ffm", "lon": "automatic"}, result)
}

func TestRegionsMapToStringMap_NullReturnsEmptyMap(t *testing.T) {
	ctx := context.Background()

	result, diags := regionsMapToStringMap(ctx, types.MapNull(types.StringType))
	require.False(t, diags.HasError())
	assert.Empty(t, result)
}

func TestFullRegionsMap_FillsMissingRegionsWithAutomatic(t *testing.T) {
	configured := map[string]string{
		"ams": "ffm",
		"ffm": "custom",
	}

	result := fullRegionsMap(configured)
	require.Len(t, result, len(knownRegions))
	assert.Equal(t, "ffm", result["ams"])
	assert.Equal(t, "custom", result["ffm"])

	for _, region := range knownRegions {
		if region != "ams" && region != "ffm" {
			assert.Equal(t, automaticRegionValue, result[region], "region %q should default to automatic", region)
		}
	}
}

func TestRefreshTrackedRegions_OnlyKeepsPreviouslyTrackedKeys(t *testing.T) {
	ctx := context.Background()

	prior, diags := types.MapValueFrom(ctx, types.StringType, map[string]string{
		"ams": "automatic",
	})
	require.False(t, diags.HasError())

	apiRegions := map[string]string{
		"ams": "ffm",
		"lon": "automatic",
	}

	refreshed, diags := refreshTrackedRegions(ctx, prior, apiRegions)
	require.False(t, diags.HasError())

	var result map[string]string
	diags = refreshed.ElementsAs(ctx, &result, false)
	require.False(t, diags.HasError())

	assert.Equal(t, map[string]string{"ams": "ffm"}, result)
}

func TestRefreshTrackedRegions_NullPassesThrough(t *testing.T) {
	ctx := context.Background()

	refreshed, diags := refreshTrackedRegions(ctx, types.MapNull(types.StringType), map[string]string{"ams": "ffm"})
	require.False(t, diags.HasError())
	assert.True(t, refreshed.IsNull())
}
