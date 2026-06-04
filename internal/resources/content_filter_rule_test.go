package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewContentFilterRuleResource(t *testing.T) {
	r := NewContentFilterRuleResource()
	require.NotNil(t, r)
	_, ok := r.(*ContentFilterRuleResource)
	assert.True(t, ok)
}

func TestContentFilterRuleResource_Metadata(t *testing.T) {
	r := &ContentFilterRuleResource{}
	ctx := context.Background()

	req := metadataReq("link11waap")
	resp := metadataResp()
	r.Metadata(ctx, req, resp)

	assert.Equal(t, "link11waap_content_filter_rule", resp.TypeName)
}

func TestContentFilterRuleResource_Schema(t *testing.T) {
	r := &ContentFilterRuleResource{}
	ctx := context.Background()

	req := schemaReq()
	resp := schemaResp()
	r.Schema(ctx, req, resp)

	s := resp.Schema
	assert.NotEmpty(t, s.Attributes)

	expectedAttrs := []string{
		"config_id", "id", "name", "msg", "operand",
		"category", "subcategory", "risk", "description", "tags",
	}
	for _, attr := range expectedAttrs {
		_, ok := s.Attributes[attr]
		assert.True(t, ok, "expected attribute %q in schema", attr)
	}
}

func TestContentFilterRuleResource_Schema_RiskValidators(t *testing.T) {
	r := &ContentFilterRuleResource{}
	resp := schemaResp()
	r.Schema(context.Background(), schemaReq(), resp)

	risk, ok := resp.Schema.Attributes["risk"].(schema.Int64Attribute)
	require.True(t, ok, "risk attribute should be Int64Attribute")
	assert.True(t, risk.Required)
	assert.NotEmpty(t, risk.Validators, "risk should have a 1..5 range validator")
}

func TestContentFilterRuleResource_Schema_NameOperandValidators(t *testing.T) {
	r := &ContentFilterRuleResource{}
	resp := schemaResp()
	r.Schema(context.Background(), schemaReq(), resp)

	name, ok := resp.Schema.Attributes["name"].(schema.StringAttribute)
	require.True(t, ok, "name attribute should be StringAttribute")
	assert.True(t, name.Required)
	assert.NotEmpty(t, name.Validators, "name should reject empty strings via LengthAtLeast(1)")

	operand, ok := resp.Schema.Attributes["operand"].(schema.StringAttribute)
	require.True(t, ok, "operand attribute should be StringAttribute")
	assert.True(t, operand.Required)
	assert.NotEmpty(t, operand.Validators, "operand should reject empty strings via LengthAtLeast(1)")
}

func TestContentFilterRuleResource_Configure_NilProvider(t *testing.T) {
	r := &ContentFilterRuleResource{}
	ctx := context.Background()

	req := configureReq(nil)
	resp := configureResp()
	r.Configure(ctx, req, resp)

	assert.Nil(t, r.client)
	assert.False(t, resp.Diagnostics.HasError())
}

func TestContentFilterRuleResource_ImportState_Valid(t *testing.T) {
	r := &ContentFilterRuleResource{}
	resp := testImportState(t, r, "config123/cfr456")

	assert.False(t, resp.Diagnostics.HasError())
}

func TestContentFilterRuleResource_ImportState_Invalid(t *testing.T) {
	r := &ContentFilterRuleResource{}
	resp := testImportState(t, r, "invalid")

	assert.True(t, resp.Diagnostics.HasError())
}

func TestContentFilterRuleResource_ImportState_TooManyParts(t *testing.T) {
	r := &ContentFilterRuleResource{}
	resp := testImportState(t, r, "a/b/c")

	assert.True(t, resp.Diagnostics.HasError())
}
