package resources

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEdgeFunctionResource(t *testing.T) {
	r := NewEdgeFunctionResource()
	require.NotNil(t, r)
	_, ok := r.(*EdgeFunctionResource)
	assert.True(t, ok)
}

func TestEdgeFunctionResource_Metadata(t *testing.T) {
	r := &EdgeFunctionResource{}
	ctx := context.Background()

	req := metadataReq("link11waap")
	resp := metadataResp()
	r.Metadata(ctx, req, resp)

	assert.Equal(t, "link11waap_edge_function", resp.TypeName)
}

func TestEdgeFunctionResource_Schema(t *testing.T) {
	r := &EdgeFunctionResource{}
	ctx := context.Background()

	req := schemaReq()
	resp := schemaResp()
	r.Schema(ctx, req, resp)

	schema := resp.Schema
	assert.NotEmpty(t, schema.Attributes)

	expectedAttrs := []string{
		"config_id", "id", "name", "description", "code", "phase",
	}
	for _, attr := range expectedAttrs {
		_, ok := schema.Attributes[attr]
		assert.True(t, ok, "expected attribute %q in schema", attr)
	}
}

func TestEdgeFunctionResource_Configure_NilProvider(t *testing.T) {
	r := &EdgeFunctionResource{}
	ctx := context.Background()

	req := configureReq(nil)
	resp := configureResp()
	r.Configure(ctx, req, resp)

	assert.Nil(t, r.client)
	assert.False(t, resp.Diagnostics.HasError())
}

func TestEdgeFunctionResource_ImportState_Valid(t *testing.T) {
	r := &EdgeFunctionResource{}
	resp := testImportState(t, r, "config123/ef456")

	assert.False(t, resp.Diagnostics.HasError())
}

func TestEdgeFunctionResource_ImportState_Invalid(t *testing.T) {
	r := &EdgeFunctionResource{}
	resp := testImportState(t, r, "invalid")

	assert.True(t, resp.Diagnostics.HasError())
}

func TestEdgeFunctionResource_ImportState_TooManyParts(t *testing.T) {
	r := &EdgeFunctionResource{}
	resp := testImportState(t, r, "a/b/c")

	assert.True(t, resp.Diagnostics.HasError())
}
