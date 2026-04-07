package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMobileApplicationGroupResource(t *testing.T) {
	r := NewMobileApplicationGroupResource()
	require.NotNil(t, r)
	_, ok := r.(*MobileApplicationGroupResource)
	assert.True(t, ok)
}

func TestMobileApplicationGroupResource_Metadata(t *testing.T) {
	r := &MobileApplicationGroupResource{}
	ctx := context.Background()

	req := metadataReq("link11waap")
	resp := metadataResp()
	r.Metadata(ctx, req, resp)

	assert.Equal(t, "link11waap_mobile_application_group", resp.TypeName)
}

func TestMobileApplicationGroupResource_Schema(t *testing.T) {
	r := &MobileApplicationGroupResource{}
	ctx := context.Background()

	req := schemaReq()
	resp := schemaResp()
	r.Schema(ctx, req, resp)

	schema := resp.Schema
	assert.NotEmpty(t, schema.Attributes)

	expectedAttrs := []string{
		"config_id", "id", "name", "description", "uid_header", "grace",
	}
	for _, attr := range expectedAttrs {
		_, ok := schema.Attributes[attr]
		assert.True(t, ok, "expected attribute %q in schema", attr)
	}

	// Check blocks
	expectedBlocks := []string{"active_config", "signatures"}
	for _, block := range expectedBlocks {
		_, ok := schema.Blocks[block]
		assert.True(t, ok, "expected block %q in schema", block)
	}
}

func TestMobileApplicationGroupResource_Configure_NilProvider(t *testing.T) {
	r := &MobileApplicationGroupResource{}
	ctx := context.Background()

	req := configureReq(nil)
	resp := configureResp()
	r.Configure(ctx, req, resp)

	assert.Nil(t, r.client)
	assert.False(t, resp.Diagnostics.HasError())
}

func TestMobileApplicationGroupResource_ImportState_Valid(t *testing.T) {
	r := &MobileApplicationGroupResource{}
	resp := testImportState(t, r, "config123/mag456")

	assert.False(t, resp.Diagnostics.HasError())
}

func TestMobileApplicationGroupResource_ImportState_Invalid(t *testing.T) {
	r := &MobileApplicationGroupResource{}
	resp := testImportState(t, r, "invalid")

	assert.True(t, resp.Diagnostics.HasError())
}

func TestMobileApplicationGroupResource_ImportState_TooManyParts(t *testing.T) {
	r := &MobileApplicationGroupResource{}
	resp := testImportState(t, r, "a/b/c")

	assert.True(t, resp.Diagnostics.HasError())
}

func TestBuildMobileApplicationGroupAPIModel_BasicFields(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	plan := &MobileApplicationGroupResourceModel{
		ID:           types.StringValue("mag-1"),
		Name:         types.StringValue("test-mag"),
		Description:  types.StringValue("desc"),
		UIDHeader:    types.StringValue("X-UID"),
		Grace:        types.StringValue("5m"),
		ActiveConfig: types.SetNull(types.ObjectType{AttrTypes: activeConfigAttrTypes()}),
		Signatures:   types.SetNull(types.ObjectType{AttrTypes: signatureAttrTypes()}),
	}

	mag := buildMobileApplicationGroupAPIModel(ctx, plan, &diags)

	assert.False(t, diags.HasError())
	assert.Equal(t, "mag-1", mag.ID)
	assert.Equal(t, "test-mag", mag.Name)
	assert.Equal(t, "desc", mag.Description)
	assert.Equal(t, "X-UID", mag.UIDHeader)
	assert.Equal(t, "5m", mag.Grace)
	assert.Nil(t, mag.ActiveConfig)
	assert.Nil(t, mag.Signatures)
}

func TestBuildMobileApplicationGroupAPIModel_WithActiveConfig(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	acModels := []ActiveConfigModel{
		{Active: types.BoolValue(true), JSON: types.StringValue(`{"key":"val"}`), Name: types.StringValue("config1")},
		{Active: types.BoolValue(false), JSON: types.StringValue(`{}`), Name: types.StringValue("config2")},
	}
	acSet, d := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: activeConfigAttrTypes()}, acModels)
	require.False(t, d.HasError())

	plan := &MobileApplicationGroupResourceModel{
		ID:           types.StringValue("mag-2"),
		Name:         types.StringValue("test"),
		Description:  types.StringValue(""),
		UIDHeader:    types.StringValue(""),
		Grace:        types.StringValue(""),
		ActiveConfig: acSet,
		Signatures:   types.SetNull(types.ObjectType{AttrTypes: signatureAttrTypes()}),
	}

	mag := buildMobileApplicationGroupAPIModel(ctx, plan, &diags)

	assert.False(t, diags.HasError())
	require.Len(t, mag.ActiveConfig, 2)
	assert.True(t, mag.ActiveConfig[0].Active || mag.ActiveConfig[1].Active)
}

func TestBuildMobileApplicationGroupAPIModel_WithSignatures(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	sigModels := []SignatureModel{
		{Active: types.BoolValue(true), Hash: types.StringValue("abc123"), Name: types.StringValue("sig1")},
	}
	sigSet, d := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: signatureAttrTypes()}, sigModels)
	require.False(t, d.HasError())

	plan := &MobileApplicationGroupResourceModel{
		ID:           types.StringValue("mag-3"),
		Name:         types.StringValue("test"),
		Description:  types.StringValue(""),
		UIDHeader:    types.StringValue(""),
		Grace:        types.StringValue(""),
		ActiveConfig: types.SetNull(types.ObjectType{AttrTypes: activeConfigAttrTypes()}),
		Signatures:   sigSet,
	}

	mag := buildMobileApplicationGroupAPIModel(ctx, plan, &diags)

	assert.False(t, diags.HasError())
	require.Len(t, mag.Signatures, 1)
	assert.True(t, mag.Signatures[0].Active)
	assert.Equal(t, "abc123", mag.Signatures[0].Hash)
	assert.Equal(t, "sig1", mag.Signatures[0].Name)
}

func TestActiveConfigAttrTypes(t *testing.T) {
	attrTypes := activeConfigAttrTypes()
	assert.Contains(t, attrTypes, "active")
	assert.Contains(t, attrTypes, "json")
	assert.Contains(t, attrTypes, "name")
	assert.Len(t, attrTypes, 3)
}

func TestSignatureAttrTypes(t *testing.T) {
	attrTypes := signatureAttrTypes()
	assert.Contains(t, attrTypes, "active")
	assert.Contains(t, attrTypes, "hash")
	assert.Contains(t, attrTypes, "name")
	assert.Len(t, attrTypes, 3)
}
