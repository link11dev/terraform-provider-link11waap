package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
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

func TestAllRegionsDefaultModifier_Description(t *testing.T) {
	m := allRegionsDefaultModifier{}
	ctx := context.Background()

	desc := m.Description(ctx)
	assert.Contains(t, desc, "automatic")

	mdDesc := m.MarkdownDescription(ctx)
	assert.Contains(t, mdDesc, "automatic")
}

func TestAllRegionsDefaultModifier_PlanModifyMap_NullValue(t *testing.T) {
	m := allRegionsDefaultModifier{}
	ctx := context.Background()

	req := planmodifier.MapRequest{
		PlanValue: types.MapNull(types.StringType),
	}
	resp := &planmodifier.MapResponse{
		PlanValue: req.PlanValue,
	}

	m.PlanModifyMap(ctx, req, resp)

	// Should not modify null values
	assert.True(t, resp.PlanValue.IsNull())
	assert.False(t, resp.Diagnostics.HasError())
}

func TestAllRegionsDefaultModifier_PlanModifyMap_UnknownValue(t *testing.T) {
	m := allRegionsDefaultModifier{}
	ctx := context.Background()

	req := planmodifier.MapRequest{
		PlanValue: types.MapUnknown(types.StringType),
	}
	resp := &planmodifier.MapResponse{
		PlanValue: req.PlanValue,
	}

	m.PlanModifyMap(ctx, req, resp)

	// Should not modify unknown values
	assert.True(t, resp.PlanValue.IsUnknown())
	assert.False(t, resp.Diagnostics.HasError())
}

func TestAllRegionsDefaultModifier_PlanModifyMap_FillsMissingRegions(t *testing.T) {
	m := allRegionsDefaultModifier{}
	ctx := context.Background()

	// Only provide two regions
	planMap, diags := types.MapValue(types.StringType, map[string]attr.Value{
		"ams": types.StringValue("us-east-1"),
		"ffm": types.StringValue("eu-central-1"),
	})
	require.False(t, diags.HasError())

	req := planmodifier.MapRequest{
		PlanValue: planMap,
	}
	resp := &planmodifier.MapResponse{
		PlanValue: req.PlanValue,
	}

	m.PlanModifyMap(ctx, req, resp)

	assert.False(t, resp.Diagnostics.HasError())
	assert.False(t, resp.PlanValue.IsNull())

	var elements map[string]string
	resp.PlanValue.ElementsAs(ctx, &elements, false)

	// Should have all known regions
	assert.Len(t, elements, len(knownRegions))

	// Provided regions should be preserved
	assert.Equal(t, "us-east-1", elements["ams"])
	assert.Equal(t, "eu-central-1", elements["ffm"])

	// Missing regions should be filled with "automatic"
	for _, region := range knownRegions {
		if region != "ams" && region != "ffm" {
			assert.Equal(t, "automatic", elements[region], "region %q should be 'automatic'", region)
		}
	}
}

func TestAllRegionsDefaultModifier_PlanModifyMap_AllRegionsPresent(t *testing.T) {
	m := allRegionsDefaultModifier{}
	ctx := context.Background()

	// Provide all regions
	values := make(map[string]attr.Value)
	for _, region := range knownRegions {
		values[region] = types.StringValue("custom-" + region)
	}
	planMap, diags := types.MapValue(types.StringType, values)
	require.False(t, diags.HasError())

	req := planmodifier.MapRequest{
		PlanValue: planMap,
	}
	resp := &planmodifier.MapResponse{
		PlanValue: req.PlanValue,
	}

	m.PlanModifyMap(ctx, req, resp)

	assert.False(t, resp.Diagnostics.HasError())

	var elements map[string]string
	resp.PlanValue.ElementsAs(ctx, &elements, false)

	// All custom values should be preserved
	for _, region := range knownRegions {
		assert.Equal(t, "custom-"+region, elements[region])
	}
}

func TestKnownRegions_Contains_ExpectedRegions(t *testing.T) {
	expected := []string{"ams", "ash", "ffm", "hkg", "lax", "lon", "nyc", "sgp", "stl"}
	assert.Equal(t, expected, knownRegions)
}
