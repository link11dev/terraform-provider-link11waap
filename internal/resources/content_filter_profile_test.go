package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/link11/terraform-provider-link11waap/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// emptySection returns a fully-populated zero-value section object value.
func emptySectionObject() types.Object {
	objType := types.ObjectType{AttrTypes: cfEntryMatchAttrTypes()}
	return types.ObjectValueMust(cfSectionAttrTypes(), map[string]attr.Value{
		"max_count":         types.Int64Value(1),
		"max_length":        types.Int64Value(1024),
		"enable_max_count":  types.BoolValue(false),
		"enable_max_length": types.BoolValue(false),
		"names":             types.ListNull(objType),
		"regex":             types.ListNull(objType),
	})
}

func emptyURLSectionObject() types.Object {
	objType := types.ObjectType{AttrTypes: cfEntryMatchURLPathAttrTypes()}
	return types.ObjectValueMust(cfURLSectionAttrTypes(), map[string]attr.Value{
		"max_count":         types.Int64Value(1),
		"max_length":        types.Int64Value(1024),
		"enable_max_count":  types.BoolValue(false),
		"enable_max_length": types.BoolValue(false),
		"regex":             types.ListNull(objType),
		"text":              types.ListNull(objType),
	})
}

func emptyPathSectionObject() types.Object {
	objType := types.ObjectType{AttrTypes: cfEntryMatchURLPathAttrTypes()}
	return types.ObjectValueMust(cfPathSectionAttrTypes(), map[string]attr.Value{
		"max_count":         types.Int64Value(1),
		"max_length":        types.Int64Value(1024),
		"enable_max_count":  types.BoolValue(false),
		"enable_max_length": types.BoolValue(false),
		"names":             types.ListNull(objType),
		"regex":             types.ListNull(objType),
		"text":              types.ListNull(objType),
	})
}

func emptyDecodingObject() types.Object {
	return types.ObjectValueMust(cfDecodingAttrTypes(), map[string]attr.Value{
		"base64":  types.BoolValue(true),
		"dual":    types.BoolValue(false),
		"html":    types.BoolValue(false),
		"unicode": types.BoolValue(false),
	})
}

func TestNewContentFilterProfileResource(t *testing.T) {
	r := NewContentFilterProfileResource()
	require.NotNil(t, r)
	_, ok := r.(*ContentFilterProfileResource)
	assert.True(t, ok)
}

func TestContentFilterProfileResource_Metadata(t *testing.T) {
	r := &ContentFilterProfileResource{}
	req := metadataReq("link11waap")
	resp := metadataResp()
	r.Metadata(context.Background(), req, resp)
	assert.Equal(t, "link11waap_content_filter_profile", resp.TypeName)
}

func TestContentFilterProfileResource_Schema(t *testing.T) {
	r := &ContentFilterProfileResource{}
	resp := schemaResp()
	r.Schema(context.Background(), schemaReq(), resp)

	s := resp.Schema
	assert.NotEmpty(t, s.Attributes)

	expectedAttrs := []string{
		"config_id", "id", "name", "description", "ignore_alphanum", "masking_seed",
		"content_type", "graphql_path", "ignore_body", "active", "report", "ignore",
		"tags", "action",
	}
	for _, attr := range expectedAttrs {
		_, ok := s.Attributes[attr]
		assert.True(t, ok, "expected attribute %q in schema", attr)
	}

	expectedBlocks := []string{"args", "headers", "cookies", "path", "url", "allsections", "decoding"}
	for _, block := range expectedBlocks {
		_, ok := s.Blocks[block]
		assert.True(t, ok, "expected block %q in schema", block)
	}
}

func TestContentFilterProfileResource_Schema_NameValidator(t *testing.T) {
	r := &ContentFilterProfileResource{}
	resp := schemaResp()
	r.Schema(context.Background(), schemaReq(), resp)

	name, ok := resp.Schema.Attributes["name"].(schema.StringAttribute)
	require.True(t, ok, "name attribute should be StringAttribute")
	assert.True(t, name.Required)
	assert.NotEmpty(t, name.Validators, "name should reject empty strings via LengthAtLeast(1)")
}

func TestContentFilterProfileResource_Schema_SectionNested(t *testing.T) {
	r := &ContentFilterProfileResource{}
	resp := schemaResp()
	r.Schema(context.Background(), schemaReq(), resp)

	args, ok := resp.Schema.Blocks["args"].(schema.SingleNestedBlock)
	require.True(t, ok, "args block should be SingleNestedBlock")
	assert.NotEmpty(t, args.Validators, "args block should have validators (IsRequired)")
	for _, attr := range []string{"max_count", "max_length", "enable_max_count", "enable_max_length"} {
		_, ok := args.Attributes[attr]
		assert.True(t, ok, "expected args attribute %q", attr)
	}
	for _, block := range []string{"names", "regex"} {
		_, ok := args.Blocks[block]
		assert.True(t, ok, "expected args sub-block %q", block)
	}
}

func TestContentFilterProfileResource_Configure_NilProvider(t *testing.T) {
	r := &ContentFilterProfileResource{}
	req := configureReq(nil)
	resp := configureResp()
	r.Configure(context.Background(), req, resp)

	assert.Nil(t, r.client)
	assert.False(t, resp.Diagnostics.HasError())
}

func TestContentFilterProfileResource_ImportState_Valid(t *testing.T) {
	r := &ContentFilterProfileResource{}
	resp := testImportState(t, r, "config123/cfp456")
	assert.False(t, resp.Diagnostics.HasError())
}

func TestContentFilterProfileResource_ImportState_Invalid(t *testing.T) {
	r := &ContentFilterProfileResource{}
	resp := testImportState(t, r, "invalid")
	assert.True(t, resp.Diagnostics.HasError())
}

func TestContentFilterProfileResource_ImportState_TooManyParts(t *testing.T) {
	r := &ContentFilterProfileResource{}
	resp := testImportState(t, r, "a/b/c")
	assert.True(t, resp.Diagnostics.HasError())
}

func TestCfSectionAttrTypes(t *testing.T) {
	m := cfSectionAttrTypes()
	for _, k := range []string{"max_count", "max_length", "enable_max_count", "enable_max_length", "names", "regex"} {
		_, ok := m[k]
		assert.True(t, ok, "expected attr type %q", k)
	}
}

func TestCfEntryMatchAttrTypes(t *testing.T) {
	m := cfEntryMatchAttrTypes()
	for _, k := range []string{"id", "parameter", "value", "restrict", "mask", "ignore_cf_rule_tags", "case_insensitive", "active"} {
		_, ok := m[k]
		assert.True(t, ok, "expected attr type %q", k)
	}
	// Type A must not contain domain/path.
	_, hasDomain := m["domain"]
	assert.False(t, hasDomain, "Type A should not contain domain")
	_, hasPath := m["path"]
	assert.False(t, hasPath, "Type A should not contain path")

	mb := cfEntryMatchURLPathAttrTypes()
	for _, k := range []string{"id", "restrict", "mask", "ignore_cf_rule_tags", "domain", "path", "case_insensitive", "active"} {
		_, ok := mb[k]
		assert.True(t, ok, "expected Type B attr type %q", k)
	}
	// Type B must not contain parameter/value.
	_, hasParam := mb["parameter"]
	assert.False(t, hasParam, "Type B should not contain parameter")
	_, hasValue := mb["value"]
	assert.False(t, hasValue, "Type B should not contain value")
}

func TestCfDecodingAttrTypes(t *testing.T) {
	m := cfDecodingAttrTypes()
	for _, k := range []string{"base64", "dual", "html", "unicode"} {
		_, ok := m[k]
		assert.True(t, ok, "expected attr type %q", k)
	}
}

// --- buildProfile unit tests ---

func TestContentFilterProfileResource_buildProfile_NullListsEmptySections(t *testing.T) {
	r := &ContentFilterProfileResource{}
	ctx := context.Background()

	plan := &ContentFilterProfileResourceModel{
		ID:             types.StringValue("id1"),
		ConfigID:       types.StringValue("cfg1"),
		Name:           types.StringValue("profile"),
		Description:    types.StringValue("desc"),
		IgnoreAlphanum: types.BoolValue(true),
		MaskingSeed:    types.StringValue("seed"),
		ContentType:    types.ListNull(types.StringType),
		GraphqlPath:    types.StringValue("$.q"),
		IgnoreBody:     types.BoolValue(false),
		Active:         types.ListNull(types.StringType),
		Report:         types.ListNull(types.StringType),
		Ignore:         types.ListNull(types.StringType),
		Tags:           types.ListNull(types.StringType),
		Action:         types.StringValue("act1"),
		Args:           emptySectionObject(),
		Headers:        emptySectionObject(),
		Cookies:        emptySectionObject(),
		Path:           emptyPathSectionObject(),
		URL:            emptyURLSectionObject(),
		AllSections:    emptySectionObject(),
		Decoding:       emptyDecodingObject(),
	}

	var diags diag.Diagnostics
	p := r.buildProfile(ctx, plan, &diags)

	require.False(t, diags.HasError(), "diags: %v", diags)
	assert.Equal(t, "id1", p.ID)
	assert.Equal(t, "profile", p.Name)
	assert.Equal(t, "desc", p.Description)
	assert.True(t, p.IgnoreAlphanum)
	assert.Equal(t, "seed", p.MaskingSeed)
	assert.Equal(t, "$.q", p.GraphqlPath)
	assert.Equal(t, "act1", p.Action)
	assert.Nil(t, p.ContentType)
	assert.Nil(t, p.Tags)
	assert.NotNil(t, p.Args.Names)
	assert.Empty(t, p.Args.Names)
	assert.True(t, p.Decoding.Base64)
}

func TestContentFilterProfileResource_buildProfile_PopulatedSections(t *testing.T) {
	r := &ContentFilterProfileResource{}
	ctx := context.Background()

	tags, d := types.ListValueFrom(ctx, types.StringType, []string{"t1", "t2"})
	require.False(t, d.HasError())

	entryObjType := types.ObjectType{AttrTypes: cfEntryMatchAttrTypes()}
	entry := types.ObjectValueMust(cfEntryMatchAttrTypes(), map[string]attr.Value{
		"id":                  types.StringValue(""),
		"parameter":           types.StringValue("mykey"),
		"value":               types.StringValue(".*"),
		"restrict":            types.BoolValue(true),
		"mask":                types.BoolValue(true),
		"ignore_cf_rule_tags": types.ListValueMust(types.StringType, []attr.Value{types.StringValue("skip")}),
		"case_insensitive":    types.BoolValue(true),
		"active":              types.BoolValue(true),
	})
	namesList := types.ListValueMust(entryObjType, []attr.Value{entry})

	section := types.ObjectValueMust(cfSectionAttrTypes(), map[string]attr.Value{
		"max_count":         types.Int64Value(5),
		"max_length":        types.Int64Value(50),
		"enable_max_count":  types.BoolValue(true),
		"enable_max_length": types.BoolValue(true),
		"names":             namesList,
		"regex":             types.ListNull(entryObjType),
	})

	decoding := types.ObjectValueMust(cfDecodingAttrTypes(), map[string]attr.Value{
		"base64":  types.BoolValue(false),
		"dual":    types.BoolValue(true),
		"html":    types.BoolValue(true),
		"unicode": types.BoolValue(true),
	})

	plan := &ContentFilterProfileResourceModel{
		ID:             types.StringValue("id1"),
		Name:           types.StringValue("profile"),
		Description:    types.StringValue(""),
		IgnoreAlphanum: types.BoolValue(false),
		MaskingSeed:    types.StringValue("seed"),
		ContentType:    types.ListNull(types.StringType),
		GraphqlPath:    types.StringValue(""),
		IgnoreBody:     types.BoolValue(false),
		Active:         types.ListNull(types.StringType),
		Report:         types.ListNull(types.StringType),
		Ignore:         types.ListNull(types.StringType),
		Tags:           tags,
		Action:         types.StringValue(""),
		Args:           section,
		Headers:        emptySectionObject(),
		Cookies:        emptySectionObject(),
		Path:           emptyPathSectionObject(),
		URL:            emptyURLSectionObject(),
		AllSections:    emptySectionObject(),
		Decoding:       decoding,
	}

	var diags diag.Diagnostics
	p := r.buildProfile(ctx, plan, &diags)

	require.False(t, diags.HasError(), "diags: %v", diags)
	assert.Equal(t, []string{"t1", "t2"}, p.Tags)
	assert.Equal(t, 5, p.Args.MaxCount)
	assert.Equal(t, 50, p.Args.MaxLength)
	assert.True(t, p.Args.EnableMaxCount)
	require.Len(t, p.Args.Names, 1)
	assert.Equal(t, "mykey", p.Args.Names[0].Key)
	assert.Equal(t, ".*", p.Args.Names[0].Reg)
	assert.True(t, p.Args.Names[0].Restrict)
	assert.True(t, p.Args.Names[0].Mask)
	assert.Equal(t, []string{"skip"}, p.Args.Names[0].IgnoreCFRuleTags)
	assert.Equal(t, "", p.Args.Names[0].Domain)
	assert.Equal(t, "", p.Args.Names[0].Path)
	assert.True(t, p.Args.Names[0].CaseInsensitive)
	assert.True(t, p.Args.Names[0].Active)
	assert.Equal(t, "names", p.Args.Names[0].Type)
	assert.NotEmpty(t, p.Args.Names[0].ID)
	// decoding round-trip
	assert.False(t, p.Decoding.Base64)
	assert.True(t, p.Decoding.Dual)
	assert.True(t, p.Decoding.HTML)
	assert.True(t, p.Decoding.Unicode)
}

// --- flattenProfile unit tests ---

func TestContentFilterProfileResource_flattenProfile_Populated(t *testing.T) {
	r := &ContentFilterProfileResource{}
	ctx := context.Background()

	p := &client.ContentFilterProfile{
		ID:             "id1",
		Name:           "profile",
		Description:    "desc",
		IgnoreAlphanum: true,
		MaskingSeed:    "seed",
		ContentType:    []string{"application/json"},
		GraphqlPath:    "$.q",
		IgnoreBody:     true,
		Tags:           []string{"t1"},
		Action:         "act1",
		Args: client.ContentFilterProfileSection{
			MaxCount: 5, MaxLength: 50, EnableMaxCount: true, EnableMaxLength: true,
			Names: []client.ContentFilterEntryMatch{
				{Key: "k", Reg: ".*", Restrict: true, Mask: true, IgnoreCFRuleTags: []string{"skip"}, Domain: "d", Path: "/p", CaseInsensitive: true, Active: true},
			},
		},
		Decoding: client.ContentFilterDecoding{Base64: true, Dual: true},
	}

	state := &ContentFilterProfileResourceModel{}
	var diags diag.Diagnostics
	r.flattenProfile(ctx, p, state, &diags)

	require.False(t, diags.HasError(), "diags: %v", diags)
	assert.Equal(t, "id1", state.ID.ValueString())
	assert.Equal(t, "profile", state.Name.ValueString())
	assert.True(t, state.IgnoreAlphanum.ValueBool())
	assert.Equal(t, "seed", state.MaskingSeed.ValueString())
	assert.True(t, state.IgnoreBody.ValueBool())
	assert.False(t, state.ContentType.IsNull())
	assert.False(t, state.Tags.IsNull())
	assert.False(t, state.Args.IsNull())
	assert.False(t, state.Decoding.IsNull())

	// Empty lists should be null.
	assert.True(t, state.Active.IsNull())
	assert.True(t, state.Report.IsNull())

	// Section names list populated.
	argsAttrs := state.Args.Attributes()
	names, ok := argsAttrs["names"].(types.List)
	require.True(t, ok)
	assert.False(t, names.IsNull())
	assert.Len(t, names.Elements(), 1)
}

func TestContentFilterProfileResource_flattenProfile_EmptyLists(t *testing.T) {
	r := &ContentFilterProfileResource{}
	ctx := context.Background()

	p := &client.ContentFilterProfile{
		ID:          "id1",
		Name:        "profile",
		MaskingSeed: "seed",
		Args:        client.ContentFilterProfileSection{},
	}

	state := &ContentFilterProfileResourceModel{}
	var diags diag.Diagnostics
	r.flattenProfile(ctx, p, state, &diags)

	require.False(t, diags.HasError(), "diags: %v", diags)
	assert.True(t, state.ContentType.IsNull())
	assert.True(t, state.Tags.IsNull())

	argsAttrs := state.Args.Attributes()
	names, ok := argsAttrs["names"].(types.List)
	require.True(t, ok)
	assert.True(t, names.IsNull(), "empty names slice should produce null list")
}
