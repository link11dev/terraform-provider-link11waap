package resources

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// entryObjectType is the tftypes object shape for a leaf entry block.
var entryObjectType = tftypes.Object{AttributeTypes: map[string]tftypes.Type{
	"type":    tftypes.String,
	"name":    tftypes.String,
	"value":   tftypes.String,
	"comment": tftypes.String,
}}

// groupObjectType is the tftypes object shape for a group block.
var groupObjectType = tftypes.Object{AttributeTypes: map[string]tftypes.Type{
	"relation":     tftypes.String,
	"entries_json": tftypes.String,
	"entry":        tftypes.List{ElementType: entryObjectType},
}}

// ruleObjectType is the tftypes object shape for the rule block.
var ruleObjectType = tftypes.Object{AttributeTypes: map[string]tftypes.Type{
	"relation":     tftypes.String,
	"entries_json": tftypes.String,
	"entry":        tftypes.List{ElementType: entryObjectType},
	"group":        tftypes.List{ElementType: groupObjectType},
}}

// newEntryValue builds a tftypes value for a leaf entry block.
func newEntryValue(typ, value, comment string) tftypes.Value {
	return tftypes.NewValue(entryObjectType, map[string]tftypes.Value{
		"type":    tftypes.NewValue(tftypes.String, typ),
		"name":    tftypes.NewValue(tftypes.String, nil),
		"value":   tftypes.NewValue(tftypes.String, value),
		"comment": tftypes.NewValue(tftypes.String, comment),
	})
}

// mustEntryList builds a types.List of EntryModel for use in RuleModel/GroupModel literals.
// RuleModel.Entries and GroupModel.Entries are types.List (not native slices) so the
// framework can represent unknown values produced by `dynamic` blocks.
func mustEntryList(t *testing.T, entries []EntryModel) types.List {
	t.Helper()
	l, diags := entryModelsToList(context.Background(), entries)
	require.False(t, diags.HasError(), "%v", diags)
	return l
}

// mustGroupList builds a types.List of GroupModel for use in RuleModel literals.
func mustGroupList(t *testing.T, groups []GroupModel) types.List {
	t.Helper()
	l, diags := groupModelsToList(context.Background(), groups)
	require.False(t, diags.HasError(), "%v", diags)
	return l
}

// entriesOf extracts []EntryModel from a types.List for assertions.
func entriesOf(t *testing.T, l types.List) []EntryModel {
	t.Helper()
	if l.IsNull() || l.IsUnknown() {
		return nil
	}
	var models []EntryModel
	diags := l.ElementsAs(context.Background(), &models, false)
	require.False(t, diags.HasError(), "%v", diags)
	return models
}

// groupsOf extracts []GroupModel from a types.List for assertions.
func groupsOf(t *testing.T, l types.List) []GroupModel {
	t.Helper()
	if l.IsNull() || l.IsUnknown() {
		return nil
	}
	var models []GroupModel
	diags := l.ElementsAs(context.Background(), &models, false)
	require.False(t, diags.HasError(), "%v", diags)
	return models
}

// newRuleValue builds a tftypes value for the rule block given entries_json and entry blocks.
func newRuleValue(relation string, entriesJSON *string, entries []tftypes.Value) tftypes.Value {
	var ejVal tftypes.Value
	if entriesJSON == nil {
		ejVal = tftypes.NewValue(tftypes.String, nil)
	} else {
		ejVal = tftypes.NewValue(tftypes.String, *entriesJSON)
	}
	if entries == nil {
		entries = []tftypes.Value{}
	}
	return tftypes.NewValue(ruleObjectType, map[string]tftypes.Value{
		"relation":     tftypes.NewValue(tftypes.String, relation),
		"entries_json": ejVal,
		"entry":        tftypes.NewValue(tftypes.List{ElementType: entryObjectType}, entries),
		"group":        tftypes.NewValue(tftypes.List{ElementType: groupObjectType}, []tftypes.Value{}),
	})
}

func TestGlobalFilterResource_ValidateConfig_MutualExclusion(t *testing.T) {
	r := &GlobalFilterResource{}
	ctx := context.Background()

	ej := `[["path","/api/",""]]`

	testCases := []struct {
		name      string
		rule      tftypes.Value
		expectErr bool
	}{
		{
			name:      "only entries_json",
			rule:      newRuleValue("OR", &ej, nil),
			expectErr: false,
		},
		{
			name:      "only entry blocks",
			rule:      newRuleValue("OR", nil, []tftypes.Value{newEntryValue("path", "/api/", "")}),
			expectErr: false,
		},
		{
			name:      "both entries_json and entry blocks",
			rule:      newRuleValue("OR", &ej, []tftypes.Value{newEntryValue("path", "/api/", "")}),
			expectErr: true,
		},
		{
			name: "entries_json set, zero entry blocks, one group block",
			rule: tftypes.NewValue(ruleObjectType, map[string]tftypes.Value{
				"relation":     tftypes.NewValue(tftypes.String, "OR"),
				"entries_json": tftypes.NewValue(tftypes.String, ej),
				"entry":        tftypes.NewValue(tftypes.List{ElementType: entryObjectType}, []tftypes.Value{}),
				"group": tftypes.NewValue(tftypes.List{ElementType: groupObjectType}, []tftypes.Value{
					tftypes.NewValue(groupObjectType, map[string]tftypes.Value{
						"relation":     tftypes.NewValue(tftypes.String, "AND"),
						"entries_json": tftypes.NewValue(tftypes.String, nil),
						"entry": tftypes.NewValue(tftypes.List{ElementType: entryObjectType}, []tftypes.Value{
							newEntryValue("path", "/api/", ""),
						}),
					}),
				}),
			}),
			expectErr: true,
		},
		{
			name:      "neither set",
			rule:      newRuleValue("OR", nil, nil),
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := buildConfig(ctx, t, r, map[string]tftypes.Value{
				"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
				"name":      tftypes.NewValue(tftypes.String, "test"),
				"rule":      tc.rule,
			})

			req := resource.ValidateConfigRequest{Config: config}
			resp := &resource.ValidateConfigResponse{}
			r.ValidateConfig(ctx, req, resp)

			if tc.expectErr {
				assert.True(t, resp.Diagnostics.HasError(), "expected mutual-exclusion error")
			} else {
				assert.False(t, resp.Diagnostics.HasError(), "did not expect an error")
			}
		})
	}
}

// TestGlobalFilterResource_ValidateConfig_UnknownEntryAndGroup reproduces the
// crash scenario: a `dynamic "entry"`/`dynamic "group"` block whose
// for_each cannot be resolved at plan time makes the whole `entry`/`group`
// collection unknown (not just individual elements' values). ValidateConfig
// must defer validation (mutual-exclusion and group checks) instead of
// erroring or panicking on the unknown collection.
func TestGlobalFilterResource_ValidateConfig_UnknownEntryAndGroup(t *testing.T) {
	r := &GlobalFilterResource{}
	ctx := context.Background()

	rule := tftypes.NewValue(ruleObjectType, map[string]tftypes.Value{
		"relation":     tftypes.NewValue(tftypes.String, "OR"),
		"entries_json": tftypes.NewValue(tftypes.String, nil),
		"entry":        tftypes.NewValue(tftypes.List{ElementType: entryObjectType}, tftypes.UnknownValue),
		"group":        tftypes.NewValue(tftypes.List{ElementType: groupObjectType}, tftypes.UnknownValue),
	})

	config := buildConfig(ctx, t, r, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"name":      tftypes.NewValue(tftypes.String, "test"),
		"rule":      rule,
	})

	req := resource.ValidateConfigRequest{Config: config}
	resp := &resource.ValidateConfigResponse{}

	assert.NotPanics(t, func() {
		r.ValidateConfig(ctx, req, resp)
	})
	assert.False(t, resp.Diagnostics.HasError(), "unknown entry/group collections should defer validation, not error: %v", resp.Diagnostics)
}

// TestRuleModelToAPI_UnknownEntryAndGroup ensures the API-model builder
// tolerates a wholly-unknown entry/group collection (as produced by an
// unresolved `dynamic` block) without panicking. Only reachable at plan time;
// by apply time Terraform guarantees fully known values.
func TestRuleModelToAPI_UnknownEntryAndGroup(t *testing.T) {
	ctx := context.Background()

	rule := &RuleModel{
		Relation:    types.StringValue("OR"),
		EntriesJSON: jsontypes.NewNormalizedNull(),
		Entries:     types.ListUnknown(entryModelType()),
		Groups:      types.ListUnknown(groupModelType()),
	}

	var api interface{}
	var diags diag.Diagnostics
	assert.NotPanics(t, func() {
		api, diags = ruleModelToAPI(ctx, rule)
	})
	require.False(t, diags.HasError(), "unexpected errors: %v", diags)

	m, ok := api.(map[string]interface{})
	require.True(t, ok, "expected map[string]interface{}, got %T", api)
	entries, ok := m["entries"].([]interface{})
	require.True(t, ok, "expected entries to be []interface{}, got %T", m["entries"])
	assert.Empty(t, entries, "entries should be empty when both collections are unknown")
}

func TestGlobalFilterResource_ValidateConfig_NilRule(t *testing.T) {
	r := &GlobalFilterResource{}
	ctx := context.Background()

	config := buildConfig(ctx, t, r, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"name":      tftypes.NewValue(tftypes.String, "test"),
		"rule":      tftypes.NewValue(ruleObjectType, nil),
	})

	req := resource.ValidateConfigRequest{Config: config}
	resp := &resource.ValidateConfigResponse{}
	r.ValidateConfig(ctx, req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "nil rule should not produce errors")
}

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

	s := resp.Schema
	assert.NotEmpty(t, s.Attributes)

	expectedAttrs := []string{
		"config_id", "id", "name", "description", "source",
		"mdate", "active", "tags", "action",
	}
	for _, attr := range expectedAttrs {
		_, ok := s.Attributes[attr]
		assert.True(t, ok, "expected attribute %q in schema", attr)
	}

	// rule should be a block, not an attribute
	_, inAttrs := s.Attributes["rule"]
	assert.False(t, inAttrs, "rule should not be in Attributes")

	ruleBlock, inBlocks := s.Blocks["rule"]
	assert.True(t, inBlocks, "rule should be in Blocks")

	// entries_json should be an Optional-only StringAttribute inside the rule block
	ruleNested, ok := ruleBlock.(schema.SingleNestedBlock)
	require.True(t, ok, "rule should be a SingleNestedBlock")
	entriesJSONAttr, ok := ruleNested.Attributes["entries_json"]
	require.True(t, ok, "expected entries_json attribute in rule block")
	ejStr, ok := entriesJSONAttr.(schema.StringAttribute)
	require.True(t, ok, "entries_json should be a StringAttribute")
	assert.True(t, ejStr.Optional, "entries_json should be Optional")
	assert.False(t, ejStr.Computed, "entries_json should not be Computed")
	assert.False(t, ejStr.Required, "entries_json should not be Required")
	_, isNormalized := ejStr.CustomType.(jsontypes.NormalizedType)
	assert.True(t, isNormalized, "entries_json should use jsontypes.NormalizedType")

	// source should be Computed only (not Required, not Optional)
	sourceAttr, ok := s.Attributes["source"]
	require.True(t, ok)
	strAttr, ok := sourceAttr.(schema.StringAttribute)
	require.True(t, ok)
	assert.True(t, strAttr.Computed, "source should be Computed")
	assert.False(t, strAttr.Required, "source should not be Required")
	assert.False(t, strAttr.Optional, "source should not be Optional")
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
		Source:      types.StringValue("anything"),
		Active:      types.BoolValue(true),
		Tags:        tagsList,
		Action:      types.StringValue("action-monitor"),
		Rule: &RuleModel{
			Relation: types.StringValue("OR"),
			Entries: mustEntryList(t, []EntryModel{
				{
					Type:    types.StringValue("path"),
					Name:    types.StringNull(),
					Value:   types.StringValue("/api/"),
					Comment: types.StringValue("API path"),
				},
			}),
		},
	}

	filter, buildDiags := buildGlobalFilterAPIModel(ctx, plan)
	require.False(t, buildDiags.HasError())
	require.NotNil(t, filter)

	assert.Equal(t, "gf1", filter.ID)
	assert.Equal(t, "test-filter", filter.Name)
	assert.Equal(t, "A test filter", filter.Description)
	assert.Equal(t, "self-managed", filter.Source)
	assert.True(t, filter.Active)
	assert.Equal(t, []string{"tag1", "tag2"}, filter.Tags)
	assert.Equal(t, "action-monitor", filter.Action)

	// Rule should be a map with relation and entries
	ruleMap, ok := filter.Rule.(map[string]interface{})
	require.True(t, ok, "rule should be a map")
	assert.Equal(t, "OR", ruleMap["relation"])

	entries, ok := ruleMap["entries"].([]interface{})
	require.True(t, ok)
	require.Len(t, entries, 1)

	entry, ok := entries[0].([]interface{})
	require.True(t, ok)
	assert.Equal(t, "path", entry[0])
	assert.Equal(t, "/api/", entry[1])
	assert.Equal(t, "API path", entry[2])
}

func TestBuildGlobalFilterAPIModel_NamedEntry(t *testing.T) {
	ctx := context.Background()

	plan := &GlobalFilterResourceModel{
		ConfigID:    types.StringValue("cfg1"),
		ID:          types.StringValue("gf1"),
		Name:        types.StringValue("test"),
		Description: types.StringValue(""),
		Active:      types.BoolValue(true),
		Tags:        types.ListNull(types.StringType),
		Action:      types.StringValue("action-monitor"),
		Rule: &RuleModel{
			Relation: types.StringValue("AND"),
			Entries: mustEntryList(t, []EntryModel{
				{
					Type:    types.StringValue("headers"),
					Name:    types.StringValue("content-type"),
					Value:   types.StringValue("application/json"),
					Comment: types.StringValue("JSON content type"),
				},
			}),
		},
	}

	filter, diags := buildGlobalFilterAPIModel(ctx, plan)
	require.False(t, diags.HasError())

	ruleMap := filter.Rule.(map[string]interface{})
	entries := ruleMap["entries"].([]interface{})
	require.Len(t, entries, 1)

	entry := entries[0].([]interface{})
	assert.Equal(t, "headers", entry[0])
	nameVal := entry[1].([]interface{})
	assert.Equal(t, "content-type", nameVal[0])
	assert.Equal(t, "application/json", nameVal[1])
	assert.Equal(t, "JSON content type", entry[2])
}

func TestBuildGlobalFilterAPIModel_WithGroups(t *testing.T) {
	ctx := context.Background()

	plan := &GlobalFilterResourceModel{
		ConfigID:    types.StringValue("cfg1"),
		ID:          types.StringValue("gf1"),
		Name:        types.StringValue("test"),
		Description: types.StringValue(""),
		Active:      types.BoolValue(true),
		Tags:        types.ListNull(types.StringType),
		Action:      types.StringValue("action-monitor"),
		Rule: &RuleModel{
			Relation: types.StringValue("OR"),
			Entries: mustEntryList(t, []EntryModel{
				{
					Type:    types.StringValue("path"),
					Name:    types.StringNull(),
					Value:   types.StringValue("/api/"),
					Comment: types.StringValue(""),
				},
			}),
			Groups: mustGroupList(t, []GroupModel{
				{
					Relation: types.StringValue("AND"),
					Entries: mustEntryList(t, []EntryModel{
						{
							Type:    types.StringValue("asn"),
							Name:    types.StringNull(),
							Value:   types.StringValue("100"),
							Comment: types.StringValue(""),
						},
					}),
				},
			}),
		},
	}

	filter, diags := buildGlobalFilterAPIModel(ctx, plan)
	require.False(t, diags.HasError())

	ruleMap := filter.Rule.(map[string]interface{})
	entries := ruleMap["entries"].([]interface{})
	// 1 entry + 1 group
	require.Len(t, entries, 2)

	// First is entry
	entry := entries[0].([]interface{})
	assert.Equal(t, "path", entry[0])

	// Second is group
	group := entries[1].(map[string]interface{})
	assert.Equal(t, "AND", group["relation"])
	groupEntries := group["entries"].([]interface{})
	require.Len(t, groupEntries, 1)
}

func TestBuildGlobalFilterAPIModel_NilRule(t *testing.T) {
	ctx := context.Background()

	plan := &GlobalFilterResourceModel{
		ConfigID:    types.StringValue("cfg1"),
		ID:          types.StringValue("gf1"),
		Name:        types.StringValue("test"),
		Description: types.StringValue(""),
		Active:      types.BoolValue(true),
		Tags:        types.ListNull(types.StringType),
		Action:      types.StringValue("action-monitor"),
		Rule:        nil,
	}

	filter, diags := buildGlobalFilterAPIModel(ctx, plan)
	require.False(t, diags.HasError())
	assert.Nil(t, filter.Rule)
}

func TestGlobalFilterResource_Source_AlwaysSelfManaged(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name   string
		source types.String
	}{
		{"empty source", types.StringValue("")},
		{"custom source", types.StringValue("https://example.com/filter.json")},
		{"null source", types.StringNull()},
		{"unknown source", types.StringUnknown()},
		{"self-managed source", types.StringValue("self-managed")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			plan := &GlobalFilterResourceModel{
				ConfigID:    types.StringValue("cfg1"),
				ID:          types.StringValue("gf1"),
				Name:        types.StringValue("test"),
				Description: types.StringValue(""),
				Source:      tc.source,
				Active:      types.BoolValue(true),
				Tags:        types.ListNull(types.StringType),
				Action:      types.StringValue("action-monitor"),
				Rule:        nil,
			}

			filter, diags := buildGlobalFilterAPIModel(ctx, plan)
			require.False(t, diags.HasError())
			assert.Equal(t, "self-managed", filter.Source, "source should always be self-managed")
		})
	}
}

func TestApiRuleToModel_SimpleEntries(t *testing.T) {
	ctx := context.Background()
	raw := map[string]interface{}{
		"relation": "OR",
		"entries": []interface{}{
			[]interface{}{"path", "/api/", "API path"},
			[]interface{}{"asn", "100", ""},
		},
	}

	model, err := apiRuleToModel(ctx, raw, nil)
	require.NoError(t, err)
	require.NotNil(t, model)

	assert.Equal(t, "OR", model.Relation.ValueString())
	entries := entriesOf(t, model.Entries)
	require.Len(t, entries, 2)
	assert.Equal(t, "path", entries[0].Type.ValueString())
	assert.Equal(t, "/api/", entries[0].Value.ValueString())
	assert.Equal(t, "API path", entries[0].Comment.ValueString())
	assert.True(t, entries[0].Name.IsNull())

	assert.Equal(t, "asn", entries[1].Type.ValueString())
	assert.Equal(t, "100", entries[1].Value.ValueString())
}

func TestApiRuleToModel_NamedEntry(t *testing.T) {
	ctx := context.Background()
	raw := map[string]interface{}{
		"relation": "AND",
		"entries": []interface{}{
			[]interface{}{"headers", []interface{}{"content-type", "application/json"}, "JSON type"},
		},
	}

	model, err := apiRuleToModel(ctx, raw, nil)
	require.NoError(t, err)
	entries := entriesOf(t, model.Entries)
	require.Len(t, entries, 1)

	entry := entries[0]
	assert.Equal(t, "headers", entry.Type.ValueString())
	assert.Equal(t, "content-type", entry.Name.ValueString())
	assert.Equal(t, "application/json", entry.Value.ValueString())
	assert.Equal(t, "JSON type", entry.Comment.ValueString())
}

func TestApiRuleToModel_WithGroup(t *testing.T) {
	ctx := context.Background()
	raw := map[string]interface{}{
		"relation": "OR",
		"entries": []interface{}{
			[]interface{}{"path", "/api/", ""},
			map[string]interface{}{
				"relation": "AND",
				"entries": []interface{}{
					[]interface{}{"asn", "100", ""},
					[]interface{}{"cookies", []interface{}{"session", "abc"}, "session cookie"},
				},
			},
		},
	}

	model, err := apiRuleToModel(ctx, raw, nil)
	require.NoError(t, err)
	require.NotNil(t, model)

	assert.Equal(t, "OR", model.Relation.ValueString())
	entries := entriesOf(t, model.Entries)
	require.Len(t, entries, 1)
	groups := groupsOf(t, model.Groups)
	require.Len(t, groups, 1)

	group := groups[0]
	assert.Equal(t, "AND", group.Relation.ValueString())
	groupEntries := entriesOf(t, group.Entries)
	require.Len(t, groupEntries, 2)

	assert.Equal(t, "asn", groupEntries[0].Type.ValueString())
	assert.Equal(t, "100", groupEntries[0].Value.ValueString())

	assert.Equal(t, "cookies", groupEntries[1].Type.ValueString())
	assert.Equal(t, "session", groupEntries[1].Name.ValueString())
	assert.Equal(t, "abc", groupEntries[1].Value.ValueString())
	assert.Equal(t, "session cookie", groupEntries[1].Comment.ValueString())
}

func TestApiRuleToModel_Nil(t *testing.T) {
	model, err := apiRuleToModel(context.Background(), nil, nil)
	require.NoError(t, err)
	assert.Nil(t, model)
}

func TestApiRuleToModel_InvalidType(t *testing.T) {
	_, err := apiRuleToModel(context.Background(), "not a map", nil)
	assert.Error(t, err)
}

func TestApiEntryToModel_TooFewElements(t *testing.T) {
	_, err := apiEntryToModel([]interface{}{"path"})
	assert.Error(t, err)
}

func TestRuleRoundTrip(t *testing.T) {
	ctx := context.Background()
	original := &RuleModel{
		Relation: types.StringValue("OR"),
		Entries: mustEntryList(t, []EntryModel{
			{
				Type:    types.StringValue("path"),
				Name:    types.StringNull(),
				Value:   types.StringValue("/api/"),
				Comment: types.StringValue("API path"),
			},
			{
				Type:    types.StringValue("headers"),
				Name:    types.StringValue("content-type"),
				Value:   types.StringValue("application/json"),
				Comment: types.StringValue(""),
			},
		}),
		Groups: mustGroupList(t, []GroupModel{
			{
				Relation: types.StringValue("AND"),
				Entries: mustEntryList(t, []EntryModel{
					{
						Type:    types.StringValue("asn"),
						Name:    types.StringNull(),
						Value:   types.StringValue("100"),
						Comment: types.StringValue(""),
					},
				}),
			},
		}),
	}

	// Convert to API format
	apiData, apiDiags := ruleModelToAPI(ctx, original)
	require.False(t, apiDiags.HasError())
	require.NotNil(t, apiData)

	// Convert back to model
	result, err := apiRuleToModel(ctx, apiData, nil)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, original.Relation.ValueString(), result.Relation.ValueString())
	entries := entriesOf(t, result.Entries)
	groups := groupsOf(t, result.Groups)
	require.Len(t, entries, 2)
	require.Len(t, groups, 1)

	// Verify simple entry round-trips
	assert.Equal(t, "path", entries[0].Type.ValueString())
	assert.Equal(t, "/api/", entries[0].Value.ValueString())
	assert.True(t, entries[0].Name.IsNull())

	// Verify named entry round-trips
	assert.Equal(t, "headers", entries[1].Type.ValueString())
	assert.Equal(t, "content-type", entries[1].Name.ValueString())
	assert.Equal(t, "application/json", entries[1].Value.ValueString())

	// Verify group round-trips
	assert.Equal(t, "AND", groups[0].Relation.ValueString())
	groupEntries := entriesOf(t, groups[0].Entries)
	require.Len(t, groupEntries, 1)
	assert.Equal(t, "asn", groupEntries[0].Type.ValueString())
}

func TestRuleModelToAPI_EntriesJSON(t *testing.T) {
	rule := &RuleModel{
		Relation:    types.StringValue("OR"),
		EntriesJSON: jsontypes.NewNormalizedValue(`[["path","/api/","API path"],["ip","203.0.113.0/24","Suspicious subnet"]]`),
	}

	apiData, diags := ruleModelToAPI(context.Background(), rule)
	require.False(t, diags.HasError())
	require.NotNil(t, apiData)

	ruleMap, ok := apiData.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "OR", ruleMap["relation"])

	entries, ok := ruleMap["entries"].([]interface{})
	require.True(t, ok)
	require.Len(t, entries, 2)

	first, ok := entries[0].([]interface{})
	require.True(t, ok)
	assert.Equal(t, "path", first[0])
	assert.Equal(t, "/api/", first[1])
	assert.Equal(t, "API path", first[2])

	second, ok := entries[1].([]interface{})
	require.True(t, ok)
	assert.Equal(t, "ip", second[0])
	assert.Equal(t, "203.0.113.0/24", second[1])
	assert.Equal(t, "Suspicious subnet", second[2])
}

func TestRuleModelToAPI_EntriesJSON_EmptyArray(t *testing.T) {
	rule := &RuleModel{
		Relation:    types.StringValue("OR"),
		EntriesJSON: jsontypes.NewNormalizedValue(`[]`),
	}

	apiData, diags := ruleModelToAPI(context.Background(), rule)
	require.False(t, diags.HasError())

	ruleMap := apiData.(map[string]interface{})
	entries, ok := ruleMap["entries"].([]interface{})
	require.True(t, ok)
	assert.Len(t, entries, 0)
}

func TestRuleModelToAPI_EntriesJSON_Invalid(t *testing.T) {
	rule := &RuleModel{
		Relation:    types.StringValue("OR"),
		EntriesJSON: jsontypes.NewNormalizedValue(`not valid json`),
	}

	apiData, diags := ruleModelToAPI(context.Background(), rule)
	assert.True(t, diags.HasError(), "invalid entries_json should produce a diagnostic")
	assert.Nil(t, apiData)
}

func diagMessages(resp *validator.StringResponse) string {
	var parts []string
	for _, d := range resp.Diagnostics {
		parts = append(parts, d.Summary()+": "+d.Detail())
	}
	return strings.Join(parts, "\n")
}

func TestEntriesJSONLeafValidator(t *testing.T) {
	v := entriesJSONLeafValidator{}
	ctx := context.Background()

	tests := []struct {
		name        string
		input       string
		null        bool
		wantError   bool
		errContains string
	}{
		{
			name:  "valid flat entries",
			input: `[["path","/api/","API path"],["ip","203.0.113.0/24",""]]`,
		},
		{
			name:  "empty array",
			input: `[]`,
		},
		{
			name: "null value is skipped",
			null: true,
		},
		{
			name:        "not a JSON array — object",
			input:       `{"key":"value"}`,
			wantError:   true,
			errContains: "must be a JSON array",
		},
		{
			name:        "contains a group object",
			input:       `[["path","/api/","ok"],{"relation":"AND","entries":[]}]`,
			wantError:   true,
			errContains: "entry[1]: must be a JSON array",
		},
		{
			name:        "tuple has wrong length — too short",
			input:       `[["path","/api/"]]`,
			wantError:   true,
			errContains: "must have exactly 3 elements",
		},
		{
			name:        "tuple has wrong length — too long",
			input:       `[["path","/api/","comment","extra"]]`,
			wantError:   true,
			errContains: "must have exactly 3 elements",
		},
		{
			name:        "value field is not a string",
			input:       `[["path",123,""]]`,
			wantError:   true,
			errContains: "must be a string",
		},
		{
			name:        "type field is a nested array",
			input:       `[[["nested"],"value",""]]`,
			wantError:   true,
			errContains: "must be a string",
		},
		{
			name:        "unknown type",
			input:       `[["not-a-type","value",""]]`,
			wantError:   true,
			errContains: "unknown type",
		},
		{
			name:  "all known types accepted",
			input: `[["asn","100",""],["ip","1.2.3.4",""],["path","/",""],["country","DE",""],["method","GET",""],["headers","",""],["cookies","",""],["uri","/x",""],["query","q",""],["region","EU",""],["subregion","x",""],["tag","t",""],["session","s",""],["network","net",""],["authority","a",""],["company","c",""],["organization","o",""],["securitypolicy","sp",""],["securitypolicyentry","spe",""],["securitypolicyentryid","spei",""],["securitypolicyentryname","spen",""],["securitypolicyid","spi",""],["securitypolicyname","spn",""],["secpolentryid","sei",""],["secpolid","si",""],["secpolentryname","sen",""],["secpolname","sn",""]]`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var configValue types.String
			if tc.null {
				configValue = types.StringNull()
			} else {
				configValue = types.StringValue(tc.input)
			}
			req := validator.StringRequest{
				Path:        path.Root("entries_json"),
				ConfigValue: configValue,
			}
			resp := &validator.StringResponse{}
			v.ValidateString(ctx, req, resp)

			if tc.wantError {
				require.True(t, resp.Diagnostics.HasError(), "expected validation error for input: %s", tc.input)
				if tc.errContains != "" {
					assert.Contains(t, diagMessages(resp), tc.errContains)
				}
			} else {
				assert.False(t, resp.Diagnostics.HasError(), "unexpected error: %s", diagMessages(resp))
			}
		})
	}
}

func TestEntriesJSONLeafValidator_StopsAfterMaxErrors(t *testing.T) {
	v := entriesJSONLeafValidator{}
	ctx := context.Background()

	// Build an array with 25 bad entries (objects instead of arrays).
	parts := make([]string, 25)
	for i := range parts {
		parts[i] = fmt.Sprintf(`{"bad":%d}`, i)
	}
	input := "[" + strings.Join(parts, ",") + "]"

	req := validator.StringRequest{
		Path:        path.Root("entries_json"),
		ConfigValue: types.StringValue(input),
	}
	resp := &validator.StringResponse{}
	v.ValidateString(ctx, req, resp)

	require.True(t, resp.Diagnostics.HasError())
	// Should have exactly entriesJSONMaxErrors entry errors + 1 stop message.
	assert.Equal(t, entriesJSONMaxErrors+1, len(resp.Diagnostics))
	assert.Contains(t, diagMessages(resp), "validation stopped after")
}

func TestRuleModelToAPI_EntriesJSON_EmptyStringUsesBlocks(t *testing.T) {
	rule := &RuleModel{
		Relation:    types.StringValue("OR"),
		EntriesJSON: jsontypes.NewNormalizedValue(""),
		Entries: mustEntryList(t, []EntryModel{
			{
				Type:    types.StringValue("path"),
				Name:    types.StringNull(),
				Value:   types.StringValue("/api/"),
				Comment: types.StringValue(""),
			},
		}),
	}

	apiData, diags := ruleModelToAPI(context.Background(), rule)
	require.False(t, diags.HasError())

	ruleMap := apiData.(map[string]interface{})
	entries := ruleMap["entries"].([]interface{})
	require.Len(t, entries, 1)
	entry := entries[0].([]interface{})
	assert.Equal(t, "path", entry[0])
}

func TestRuleModelToAPI_NilRule(t *testing.T) {
	apiData, diags := ruleModelToAPI(context.Background(), nil)
	require.False(t, diags.HasError())
	assert.Nil(t, apiData)
}

func TestApiRuleToModel_PriorUsedJSON(t *testing.T) {
	raw := map[string]interface{}{
		"relation": "OR",
		"entries": []interface{}{
			[]interface{}{"path", "/api/", "API path"},
			map[string]interface{}{
				"relation": "AND",
				"entries": []interface{}{
					[]interface{}{"asn", "100", ""},
				},
			},
		},
	}

	prior := &RuleModel{
		EntriesJSON: jsontypes.NewNormalizedValue(`[["path","/api/","API path"]]`),
	}

	model, err := apiRuleToModel(context.Background(), raw, prior)
	require.NoError(t, err)
	require.NotNil(t, model)

	// Block fields should be empty; entries_json should be populated.
	assert.Empty(t, entriesOf(t, model.Entries))
	assert.Empty(t, groupsOf(t, model.Groups))
	assert.False(t, model.EntriesJSON.IsNull())

	// entries_json should semantically match the API entries array.
	expected := jsontypes.NewNormalizedValue(`[["path","/api/","API path"],{"relation":"AND","entries":[["asn","100",""]]}]`)
	eq, eqDiags := model.EntriesJSON.StringSemanticEquals(context.Background(), expected)
	require.False(t, eqDiags.HasError())
	assert.True(t, eq, "entries_json should semantically equal the round-tripped entries")
}

func TestApiRuleToModel_PriorUsedJSON_EmptyEntries(t *testing.T) {
	raw := map[string]interface{}{
		"relation": "OR",
	}

	prior := &RuleModel{
		EntriesJSON: jsontypes.NewNormalizedValue(`[]`),
	}

	model, err := apiRuleToModel(context.Background(), raw, prior)
	require.NoError(t, err)
	require.NotNil(t, model)

	assert.Empty(t, entriesOf(t, model.Entries))
	assert.Empty(t, groupsOf(t, model.Groups))
	assert.Equal(t, "[]", model.EntriesJSON.ValueString())
}

func TestApiRuleToModel_PriorNilUsesBlocks(t *testing.T) {
	raw := map[string]interface{}{
		"relation": "OR",
		"entries": []interface{}{
			[]interface{}{"path", "/api/", ""},
		},
	}

	model, err := apiRuleToModel(context.Background(), raw, nil)
	require.NoError(t, err)
	require.NotNil(t, model)

	// With nil prior, block representation is used and entries_json stays null.
	require.Len(t, entriesOf(t, model.Entries), 1)
	assert.True(t, model.EntriesJSON.IsNull())
}

func TestApiRuleToModel_PriorNoJSONUsesBlocks(t *testing.T) {
	raw := map[string]interface{}{
		"relation": "OR",
		"entries": []interface{}{
			[]interface{}{"path", "/api/", ""},
		},
	}

	prior := &RuleModel{EntriesJSON: jsontypes.NewNormalizedNull()}

	model, err := apiRuleToModel(context.Background(), raw, prior)
	require.NoError(t, err)
	require.NotNil(t, model)

	require.Len(t, entriesOf(t, model.Entries), 1)
	assert.True(t, model.EntriesJSON.IsNull())
}

func TestBuildGlobalFilterAPIModel_ActionAlwaysSet(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name           string
		action         types.String
		expectedAction string
	}{
		{"explicit action-monitor", types.StringValue("action-monitor"), "action-monitor"},
		{"action-challenge", types.StringValue("action-challenge"), "action-challenge"},
		{"action-skip", types.StringValue("action-skip"), "action-skip"},
		{"action-global-filter-block", types.StringValue("action-global-filter-block"), "action-global-filter-block"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			plan := &GlobalFilterResourceModel{
				ConfigID:    types.StringValue("cfg1"),
				ID:          types.StringValue("gf1"),
				Name:        types.StringValue("test"),
				Description: types.StringValue(""),
				Active:      types.BoolValue(true),
				Tags:        types.ListNull(types.StringType),
				Action:      tc.action,
				Rule:        nil,
			}

			filter, diags := buildGlobalFilterAPIModel(ctx, plan)
			require.False(t, diags.HasError())
			assert.Equal(t, tc.expectedAction, filter.Action, "action should always be set")
		})
	}
}

func TestBuildGlobalFilterAPIModel_CookiesNamedEntry(t *testing.T) {
	ctx := context.Background()

	plan := &GlobalFilterResourceModel{
		ConfigID:    types.StringValue("cfg1"),
		ID:          types.StringValue("gf1"),
		Name:        types.StringValue("test"),
		Description: types.StringValue(""),
		Active:      types.BoolValue(true),
		Tags:        types.ListNull(types.StringType),
		Action:      types.StringValue("action-monitor"),
		Rule: &RuleModel{
			Relation: types.StringValue("OR"),
			Entries: mustEntryList(t, []EntryModel{
				{
					Type:    types.StringValue("cookies"),
					Name:    types.StringValue("test"),
					Value:   types.StringValue("ddddd"),
					Comment: types.StringValue("dddd"),
				},
			}),
		},
	}

	filter, diags := buildGlobalFilterAPIModel(ctx, plan)
	require.False(t, diags.HasError())

	ruleMap := filter.Rule.(map[string]interface{})
	entries := ruleMap["entries"].([]interface{})
	require.Len(t, entries, 1)

	entry := entries[0].([]interface{})
	assert.Equal(t, "cookies", entry[0])
	nameVal := entry[1].([]interface{})
	assert.Equal(t, "test", nameVal[0])
	assert.Equal(t, "ddddd", nameVal[1])
	assert.Equal(t, "dddd", entry[2])
}

func TestBuildGlobalFilterAPIModel_GroupWithEntriesJSON(t *testing.T) {
	ctx := context.Background()

	plan := &GlobalFilterResourceModel{
		ConfigID:    types.StringValue("cfg1"),
		ID:          types.StringValue("gf1"),
		Name:        types.StringValue("test"),
		Description: types.StringValue(""),
		Active:      types.BoolValue(true),
		Tags:        types.ListNull(types.StringType),
		Action:      types.StringValue("action-monitor"),
		Rule: &RuleModel{
			Relation: types.StringValue("AND"),
			Groups: mustGroupList(t, []GroupModel{
				{
					Relation:    types.StringValue("OR"),
					Entries:     types.ListNull(entryModelType()),
					EntriesJSON: jsontypes.NewNormalizedValue(`[["path","/admin/","Admin paths"],["uri","/.+\\.php","PHP files"]]`),
				},
			}),
		},
	}

	filter, diags := buildGlobalFilterAPIModel(ctx, plan)
	require.False(t, diags.HasError())

	ruleMap := filter.Rule.(map[string]interface{})
	entries := ruleMap["entries"].([]interface{})
	require.Len(t, entries, 1)

	group := entries[0].(map[string]interface{})
	assert.Equal(t, "OR", group["relation"])
	groupEntries := group["entries"].([]interface{})
	require.Len(t, groupEntries, 2)

	e0 := groupEntries[0].([]interface{})
	assert.Equal(t, "path", e0[0])
	assert.Equal(t, "/admin/", e0[1])

	e1 := groupEntries[1].([]interface{})
	assert.Equal(t, "uri", e1[0])
}

func TestApiRuleToModel_GroupPriorUsedJSON(t *testing.T) {
	raw := map[string]interface{}{
		"relation": "AND",
		"entries": []interface{}{
			map[string]interface{}{
				"relation": "OR",
				"entries": []interface{}{
					[]interface{}{"path", "/admin/", "Admin paths"},
					[]interface{}{"uri", `/.+\.php`, "PHP files"},
				},
			},
		},
	}

	prior := &RuleModel{
		Relation: types.StringValue("AND"),
		Groups: mustGroupList(t, []GroupModel{
			{
				Relation:    types.StringValue("OR"),
				Entries:     types.ListNull(entryModelType()),
				EntriesJSON: jsontypes.NewNormalizedValue(`[["path","/old/",""]]`),
			},
		}),
	}

	model, err := apiRuleToModel(context.Background(), raw, prior)
	require.NoError(t, err)
	groups := groupsOf(t, model.Groups)
	require.Len(t, groups, 1)

	g := groups[0]
	assert.Equal(t, "OR", g.Relation.ValueString())
	assert.Empty(t, entriesOf(t, g.Entries))
	assert.False(t, g.EntriesJSON.IsNull())

	expected := jsontypes.NewNormalizedValue(`[["path","/admin/","Admin paths"],["uri","/.+\\.php","PHP files"]]`)
	eq, eqDiags := g.EntriesJSON.StringSemanticEquals(context.Background(), expected)
	require.False(t, eqDiags.HasError())
	assert.True(t, eq)
}

func TestApiRuleToModel_GroupPriorUsedBlocksKeepsBlocks(t *testing.T) {
	raw := map[string]interface{}{
		"relation": "AND",
		"entries": []interface{}{
			map[string]interface{}{
				"relation": "OR",
				"entries": []interface{}{
					[]interface{}{"path", "/admin/", ""},
				},
			},
		},
	}

	prior := &RuleModel{
		Relation: types.StringValue("AND"),
		Groups: mustGroupList(t, []GroupModel{
			{
				Relation:    types.StringValue("OR"),
				Entries:     types.ListNull(entryModelType()),
				EntriesJSON: jsontypes.NewNormalizedNull(),
			},
		}),
	}

	model, err := apiRuleToModel(context.Background(), raw, prior)
	require.NoError(t, err)
	groups := groupsOf(t, model.Groups)
	require.Len(t, groups, 1)

	g := groups[0]
	assert.True(t, g.EntriesJSON.IsNull())
	gEntries := entriesOf(t, g.Entries)
	require.Len(t, gEntries, 1)
	assert.Equal(t, "path", gEntries[0].Type.ValueString())
}

func TestGlobalFilterResource_ValidateConfig_GroupMutualExclusion(t *testing.T) {
	r := &GlobalFilterResource{}
	ctx := context.Background()

	ej := `[["path","/api/",""]]`

	groupWithBoth := tftypes.NewValue(groupObjectType, map[string]tftypes.Value{
		"relation":     tftypes.NewValue(tftypes.String, "OR"),
		"entries_json": tftypes.NewValue(tftypes.String, ej),
		"entry": tftypes.NewValue(tftypes.List{ElementType: entryObjectType}, []tftypes.Value{
			newEntryValue("path", "/api/", ""),
		}),
	})

	ruleVal := tftypes.NewValue(ruleObjectType, map[string]tftypes.Value{
		"relation":     tftypes.NewValue(tftypes.String, "AND"),
		"entries_json": tftypes.NewValue(tftypes.String, nil),
		"entry":        tftypes.NewValue(tftypes.List{ElementType: entryObjectType}, []tftypes.Value{}),
		"group":        tftypes.NewValue(tftypes.List{ElementType: groupObjectType}, []tftypes.Value{groupWithBoth}),
	})

	config := buildConfig(ctx, t, r, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"name":      tftypes.NewValue(tftypes.String, "test"),
		"rule":      ruleVal,
	})

	req := resource.ValidateConfigRequest{Config: config}
	resp := &resource.ValidateConfigResponse{}
	r.ValidateConfig(ctx, req, resp)

	assert.True(t, resp.Diagnostics.HasError(), "expected mutual-exclusion error for group entries_json + entry blocks")
}

func TestGlobalFilterResource_ValidateConfig_GroupEntriesJSONOnly(t *testing.T) {
	r := &GlobalFilterResource{}
	ctx := context.Background()

	ej := `[["path","/api/",""]]`

	groupWithJSON := tftypes.NewValue(groupObjectType, map[string]tftypes.Value{
		"relation":     tftypes.NewValue(tftypes.String, "OR"),
		"entries_json": tftypes.NewValue(tftypes.String, ej),
		"entry":        tftypes.NewValue(tftypes.List{ElementType: entryObjectType}, []tftypes.Value{}),
	})

	ruleVal := tftypes.NewValue(ruleObjectType, map[string]tftypes.Value{
		"relation":     tftypes.NewValue(tftypes.String, "AND"),
		"entries_json": tftypes.NewValue(tftypes.String, nil),
		"entry":        tftypes.NewValue(tftypes.List{ElementType: entryObjectType}, []tftypes.Value{}),
		"group":        tftypes.NewValue(tftypes.List{ElementType: groupObjectType}, []tftypes.Value{groupWithJSON}),
	})

	config := buildConfig(ctx, t, r, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"name":      tftypes.NewValue(tftypes.String, "test"),
		"rule":      ruleVal,
	})

	req := resource.ValidateConfigRequest{Config: config}
	resp := &resource.ValidateConfigResponse{}
	r.ValidateConfig(ctx, req, resp)

	assert.False(t, resp.Diagnostics.HasError(), "group with only entries_json should be valid: %v", resp.Diagnostics)
}

func TestGlobalFilterResource_ValidateConfig_RuleEntriesJSONNotArray(t *testing.T) {
	r := &GlobalFilterResource{}
	ctx := context.Background()

	// ValidateConfig only checks empty/whitespace and mutual-exclusivity.
	// JSON structure (array-of-3-string-tuples) is enforced by entriesJSONLeafValidator
	// at schema level; see TestEntriesJSONLeafValidator for those cases.
	testCases := []struct {
		name        string
		entriesJSON string
		expectErr   bool
	}{
		{
			name:        "valid array",
			entriesJSON: `[["path","/api/",""]]`,
			expectErr:   false,
		},
		{
			name:        "empty array",
			entriesJSON: `[]`,
			expectErr:   false,
		},
		{
			name:        "object instead of array — caught by schema validator not ValidateConfig",
			entriesJSON: `{"key":"value"}`,
			expectErr:   false,
		},
		{
			name:        "plain string — caught by schema validator not ValidateConfig",
			entriesJSON: `"just a string"`,
			expectErr:   false,
		},
		{
			name:        "empty string",
			entriesJSON: "",
			expectErr:   true,
		},
		{
			name:        "whitespace string",
			entriesJSON: "   \n\t  ",
			expectErr:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := buildConfig(ctx, t, r, map[string]tftypes.Value{
				"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
				"name":      tftypes.NewValue(tftypes.String, "test"),
				"rule":      newRuleValue("OR", &tc.entriesJSON, nil),
			})

			req := resource.ValidateConfigRequest{Config: config}
			resp := &resource.ValidateConfigResponse{}
			r.ValidateConfig(ctx, req, resp)

			if tc.expectErr {
				assert.True(t, resp.Diagnostics.HasError(), "expected error for entries_json=%q", tc.entriesJSON)
			} else {
				assert.False(t, resp.Diagnostics.HasError(), "unexpected error for entries_json=%q: %v", tc.entriesJSON, resp.Diagnostics)
			}
		})
	}
}

func TestUnmarshalEntriesJSON(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantLen   int
		wantError bool
	}{
		{
			name:    "valid array of tuples",
			input:   `[["path","/api/","comment"],["ip","1.2.3.4",""]]`,
			wantLen: 2,
		},
		{
			name:    "empty array",
			input:   `[]`,
			wantLen: 0,
		},
		{
			name:      "invalid JSON",
			input:     `not valid json`,
			wantError: true,
		},
		{
			name:      "JSON object instead of array",
			input:     `{"key":"value"}`,
			wantError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			entries, err := unmarshalEntriesJSON(tc.input)
			if tc.wantError {
				assert.Error(t, err)
				assert.Nil(t, entries)
			} else {
				require.NoError(t, err)
				assert.Len(t, entries, tc.wantLen)
			}
		})
	}
}

func TestEvalEntriesJSON(t *testing.T) {
	tests := []struct {
		name      string
		v         jsontypes.Normalized
		wantEmpty bool
		wantSet   bool
	}{
		{
			name:      "null — not empty, not set",
			v:         jsontypes.NewNormalizedNull(),
			wantEmpty: false,
			wantSet:   false,
		},
		{
			name:      "unknown — not empty, not set",
			v:         jsontypes.NewNormalizedUnknown(),
			wantEmpty: false,
			wantSet:   false,
		},
		{
			name:      "non-empty value — not empty, is set",
			v:         jsontypes.NewNormalizedValue(`[["path","/api/",""]]`),
			wantEmpty: false,
			wantSet:   true,
		},
		{
			name:      "empty string — is empty, not set",
			v:         jsontypes.NewNormalizedValue(""),
			wantEmpty: true,
			wantSet:   false,
		},
		{
			name:      "whitespace-only string — is empty, not set",
			v:         jsontypes.NewNormalizedValue("   \n\t  "),
			wantEmpty: true,
			wantSet:   false,
		},
		{
			name:      "valid JSON array — not empty, is set",
			v:         jsontypes.NewNormalizedValue(`[]`),
			wantEmpty: false,
			wantSet:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			isEmpty, isSet := evalEntriesJSON(tc.v)
			assert.Equal(t, tc.wantEmpty, isEmpty, "isEmpty")
			assert.Equal(t, tc.wantSet, isSet, "isSet")
		})
	}
}

func TestGlobalFilterResource_ValidateConfig_GroupEntriesJSONNotArray(t *testing.T) {
	r := &GlobalFilterResource{}
	ctx := context.Background()

	buildRuleWithGroupJSON := func(ej string) tftypes.Value {
		group := tftypes.NewValue(groupObjectType, map[string]tftypes.Value{
			"relation":     tftypes.NewValue(tftypes.String, "OR"),
			"entries_json": tftypes.NewValue(tftypes.String, ej),
			"entry":        tftypes.NewValue(tftypes.List{ElementType: entryObjectType}, []tftypes.Value{}),
		})
		return tftypes.NewValue(ruleObjectType, map[string]tftypes.Value{
			"relation":     tftypes.NewValue(tftypes.String, "AND"),
			"entries_json": tftypes.NewValue(tftypes.String, nil),
			"entry":        tftypes.NewValue(tftypes.List{ElementType: entryObjectType}, []tftypes.Value{}),
			"group":        tftypes.NewValue(tftypes.List{ElementType: groupObjectType}, []tftypes.Value{group}),
		})
	}

	testCases := []struct {
		name        string
		entriesJSON string
		expectErr   bool
	}{
		// ValidateConfig only checks empty/whitespace and mutual-exclusivity.
		// JSON structure is enforced by entriesJSONLeafValidator at schema level.
		{
			name:        "valid array",
			entriesJSON: `[["path","/api/",""]]`,
			expectErr:   false,
		},
		{
			name:        "object instead of array — caught by schema validator not ValidateConfig",
			entriesJSON: `{"key":"value"}`,
			expectErr:   false,
		},
		{
			name:        "empty string",
			entriesJSON: "",
			expectErr:   true,
		},
		{
			name:        "whitespace string",
			entriesJSON: "   ",
			expectErr:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := buildConfig(ctx, t, r, map[string]tftypes.Value{
				"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
				"name":      tftypes.NewValue(tftypes.String, "test"),
				"rule":      buildRuleWithGroupJSON(tc.entriesJSON),
			})

			req := resource.ValidateConfigRequest{Config: config}
			resp := &resource.ValidateConfigResponse{}
			r.ValidateConfig(ctx, req, resp)

			if tc.expectErr {
				assert.True(t, resp.Diagnostics.HasError(), "expected error for group entries_json=%q", tc.entriesJSON)
			} else {
				assert.False(t, resp.Diagnostics.HasError(), "unexpected error for group entries_json=%q: %v", tc.entriesJSON, resp.Diagnostics)
			}
		})
	}
}
