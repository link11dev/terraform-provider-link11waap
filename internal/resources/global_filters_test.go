package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGlobalFilterResource(t *testing.T) {
	r := NewGlobalFilterResource()
	require.NotNil(t, r)
	_, ok := r.(*GlobalFilterResource)
	assert.True(t, ok)
}

func TestGlobalFilterResource_Metadata(t *testing.T) {
	r := &GlobalFilterResource{}
	ctx := context.Background()

	req := metadataReq("link11waap")
	resp := metadataResp()
	r.Metadata(ctx, req, resp)

	assert.Equal(t, "link11waap_global_filter", resp.TypeName)
}

func TestGlobalFilterResource_Schema(t *testing.T) {
	r := &GlobalFilterResource{}
	ctx := context.Background()

	req := schemaReq()
	resp := schemaResp()
	r.Schema(ctx, req, resp)

	schema := resp.Schema
	assert.NotEmpty(t, schema.Attributes)

	expectedAttrs := []string{
		"config_id", "id", "name", "description", "source",
		"mdate", "active", "tags", "action", "rule",
	}
	for _, attr := range expectedAttrs {
		_, ok := schema.Attributes[attr]
		assert.True(t, ok, "expected attribute %q in schema", attr)
	}
}

func TestGlobalFilterResource_Configure_NilProvider(t *testing.T) {
	r := &GlobalFilterResource{}
	ctx := context.Background()

	req := configureReq(nil)
	resp := configureResp()
	r.Configure(ctx, req, resp)

	assert.Nil(t, r.client)
	assert.False(t, resp.Diagnostics.HasError())
}

func TestGlobalFilterResource_ImportState_Valid(t *testing.T) {
	r := &GlobalFilterResource{}
	resp := testImportState(t, r, "config123/filter456")

	assert.False(t, resp.Diagnostics.HasError())
}

func TestGlobalFilterResource_ImportState_Invalid(t *testing.T) {
	r := &GlobalFilterResource{}
	resp := testImportState(t, r, "invalidformat")

	assert.True(t, resp.Diagnostics.HasError())
}

func TestGlobalFilterResource_ImportState_TooManyParts(t *testing.T) {
	r := &GlobalFilterResource{}
	resp := testImportState(t, r, "a/b/c")

	assert.True(t, resp.Diagnostics.HasError())
}

func TestBuildGlobalFilterAPIModel(t *testing.T) {
	ctx := context.Background()

	tagsList, diags := types.ListValueFrom(ctx, types.StringType, []string{"tag1", "tag2"})
	require.False(t, diags.HasError())

	plan := &GlobalFilterResourceModel{
		ConfigID:    types.StringValue("cfg1"),
		ID:          types.StringValue("gf1"),
		Name:        types.StringValue("test-filter"),
		Description: types.StringValue("A test filter"),
		Source:      types.StringValue("https://example.com/filter.json"),
		Active:      types.BoolValue(true),
		Tags:        tagsList,
		Action:      types.StringValue("action-monitor"),
		Rule:        types.StringValue(`{"relation":"OR","entries":[]}`),
	}

	filter, buildDiags := buildGlobalFilterAPIModel(ctx, plan)
	require.False(t, buildDiags.HasError())
	require.NotNil(t, filter)

	assert.Equal(t, "gf1", filter.ID)
	assert.Equal(t, "test-filter", filter.Name)
	assert.Equal(t, "A test filter", filter.Description)
	assert.Equal(t, "https://example.com/filter.json", filter.Source)
	assert.True(t, filter.Active)
	assert.Equal(t, []string{"tag1", "tag2"}, filter.Tags)
	assert.Equal(t, "action-monitor", filter.Action)

	// Rule should be parsed from JSON
	ruleMap, ok := filter.Rule.(map[string]interface{})
	require.True(t, ok, "rule should be a map")
	assert.Equal(t, "OR", ruleMap["relation"])
}
