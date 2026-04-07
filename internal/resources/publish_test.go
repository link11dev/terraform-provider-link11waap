package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPublishResource(t *testing.T) {
	r := NewPublishResource()
	require.NotNil(t, r)
	_, ok := r.(*PublishResource)
	assert.True(t, ok)
}

func TestPublishResource_Metadata(t *testing.T) {
	r := &PublishResource{}
	ctx := context.Background()

	req := metadataReq("link11waap")
	resp := metadataResp()
	r.Metadata(ctx, req, resp)

	assert.Equal(t, "link11waap_publish", resp.TypeName)
}

func TestPublishResource_Schema(t *testing.T) {
	r := &PublishResource{}
	ctx := context.Background()

	req := schemaReq()
	resp := schemaResp()
	r.Schema(ctx, req, resp)

	schema := resp.Schema
	assert.NotEmpty(t, schema.Attributes)

	expectedAttrs := []string{
		"config_id", "buckets", "triggers", "id",
	}
	for _, attr := range expectedAttrs {
		_, ok := schema.Attributes[attr]
		assert.True(t, ok, "expected attribute %q in schema", attr)
	}
}

func TestPublishResource_Configure_NilProvider(t *testing.T) {
	r := &PublishResource{}
	ctx := context.Background()

	req := configureReq(nil)
	resp := configureResp()
	r.Configure(ctx, req, resp)

	assert.Nil(t, r.client)
	assert.False(t, resp.Diagnostics.HasError())
}

func TestPublishResource_GetBuckets_NullBuckets(t *testing.T) {
	r := &PublishResource{}
	ctx := context.Background()
	var diags diag.Diagnostics

	model := &PublishResourceModel{
		Buckets: types.ListNull(types.ObjectType{}),
	}

	buckets := r.getBuckets(ctx, model, &diags)

	assert.False(t, diags.HasError())
	assert.Empty(t, buckets)
}

func TestPublishResource_GetBuckets_UnknownBuckets(t *testing.T) {
	r := &PublishResource{}
	ctx := context.Background()
	var diags diag.Diagnostics

	model := &PublishResourceModel{
		Buckets: types.ListUnknown(types.ObjectType{}),
	}

	buckets := r.getBuckets(ctx, model, &diags)

	assert.False(t, diags.HasError())
	assert.Empty(t, buckets)
}

func TestPublishResourceModel_Fields(t *testing.T) {
	model := PublishResourceModel{
		ConfigID: types.StringValue("cfg1"),
		ID:       types.StringValue("id1"),
	}

	assert.Equal(t, "cfg1", model.ConfigID.ValueString())
	assert.Equal(t, "id1", model.ID.ValueString())
}

func TestBucketModel_Fields(t *testing.T) {
	model := BucketModel{
		Name: types.StringValue("bucket1"),
		URL:  types.StringValue("https://bucket.example.com"),
	}

	assert.Equal(t, "bucket1", model.Name.ValueString())
	assert.Equal(t, "https://bucket.example.com", model.URL.ValueString())
}
