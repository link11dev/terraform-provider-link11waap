package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewActionResource(t *testing.T) {
	r := NewActionResource()
	require.NotNil(t, r)
	_, ok := r.(*ActionResource)
	assert.True(t, ok)
}

func TestActionResource_Metadata(t *testing.T) {
	r := &ActionResource{}
	ctx := context.Background()

	req := metadataReq("link11waap")
	resp := metadataResp()
	r.Metadata(ctx, req, resp)

	assert.Equal(t, "link11waap_action", resp.TypeName)
}

func TestActionResource_Schema(t *testing.T) {
	r := &ActionResource{}
	ctx := context.Background()

	req := schemaReq()
	resp := schemaResp()
	r.Schema(ctx, req, resp)

	s := resp.Schema
	assert.NotEmpty(t, s.Attributes)

	expectedAttrs := []string{
		"config_id", "id", "name", "description", "type", "tags", "params",
	}
	for _, attr := range expectedAttrs {
		_, ok := s.Attributes[attr]
		assert.True(t, ok, "expected attribute %q in schema", attr)
	}
}

func TestActionResource_Schema_NameValidators(t *testing.T) {
	r := &ActionResource{}
	resp := schemaResp()
	r.Schema(context.Background(), schemaReq(), resp)

	name, ok := resp.Schema.Attributes["name"].(schema.StringAttribute)
	require.True(t, ok, "name attribute should be StringAttribute")
	assert.True(t, name.Required)
	assert.NotEmpty(t, name.Validators, "name should reject empty strings via LengthAtLeast(1)")
}

func TestActionResource_Schema_TypeValidators(t *testing.T) {
	r := &ActionResource{}
	resp := schemaResp()
	r.Schema(context.Background(), schemaReq(), resp)

	typ, ok := resp.Schema.Attributes["type"].(schema.StringAttribute)
	require.True(t, ok, "type attribute should be StringAttribute")
	assert.True(t, typ.Required)
	assert.NotEmpty(t, typ.Validators, "type should have a OneOf enum validator")
}

func TestActionResource_Schema_ParamsNested(t *testing.T) {
	r := &ActionResource{}
	resp := schemaResp()
	r.Schema(context.Background(), schemaReq(), resp)

	params, ok := resp.Schema.Attributes["params"].(schema.SingleNestedAttribute)
	require.True(t, ok, "params attribute should be SingleNestedAttribute")
	assert.True(t, params.Optional)
	assert.True(t, params.Computed)
	for _, attr := range []string{"content", "status", "headers"} {
		_, ok := params.Attributes[attr]
		assert.True(t, ok, "expected params attribute %q", attr)
	}
}

func TestActionResource_Configure_NilProvider(t *testing.T) {
	r := &ActionResource{}
	ctx := context.Background()

	req := configureReq(nil)
	resp := configureResp()
	r.Configure(ctx, req, resp)

	assert.Nil(t, r.client)
	assert.False(t, resp.Diagnostics.HasError())
}

func TestActionResource_ImportState_Valid(t *testing.T) {
	r := &ActionResource{}
	resp := testImportState(t, r, "config123/a456")

	assert.False(t, resp.Diagnostics.HasError())
}

func TestActionResource_ImportState_Invalid(t *testing.T) {
	r := &ActionResource{}
	resp := testImportState(t, r, "invalid")

	assert.True(t, resp.Diagnostics.HasError())
}

func TestActionResource_ImportState_TooManyParts(t *testing.T) {
	r := &ActionResource{}
	resp := testImportState(t, r, "a/b/c")

	assert.True(t, resp.Diagnostics.HasError())
}

func TestActionParamsAttrTypes(t *testing.T) {
	m := actionParamsAttrTypes()
	for _, k := range []string{"content", "status", "headers"} {
		_, ok := m[k]
		assert.True(t, ok, "expected attr type %q", k)
	}
}
