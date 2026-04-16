package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
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

	_, inBlocks := s.Blocks["rule"]
	assert.True(t, inBlocks, "rule should be in Blocks")

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
			Entries: []EntryModel{
				{
					Type:    types.StringValue("path"),
					Name:    types.StringNull(),
					Value:   types.StringValue("/api/"),
					Comment: types.StringValue("API path"),
				},
			},
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
			Entries: []EntryModel{
				{
					Type:    types.StringValue("headers"),
					Name:    types.StringValue("content-type"),
					Value:   types.StringValue("application/json"),
					Comment: types.StringValue("JSON content type"),
				},
			},
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
			Entries: []EntryModel{
				{
					Type:    types.StringValue("path"),
					Name:    types.StringNull(),
					Value:   types.StringValue("/api/"),
					Comment: types.StringValue(""),
				},
			},
			Groups: []GroupModel{
				{
					Relation: types.StringValue("AND"),
					Entries: []EntryModel{
						{
							Type:    types.StringValue("asn"),
							Name:    types.StringNull(),
							Value:   types.StringValue("100"),
							Comment: types.StringValue(""),
						},
					},
				},
			},
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
	raw := map[string]interface{}{
		"relation": "OR",
		"entries": []interface{}{
			[]interface{}{"path", "/api/", "API path"},
			[]interface{}{"asn", "100", ""},
		},
	}

	model, err := apiRuleToModel(raw)
	require.NoError(t, err)
	require.NotNil(t, model)

	assert.Equal(t, "OR", model.Relation.ValueString())
	require.Len(t, model.Entries, 2)
	assert.Equal(t, "path", model.Entries[0].Type.ValueString())
	assert.Equal(t, "/api/", model.Entries[0].Value.ValueString())
	assert.Equal(t, "API path", model.Entries[0].Comment.ValueString())
	assert.True(t, model.Entries[0].Name.IsNull())

	assert.Equal(t, "asn", model.Entries[1].Type.ValueString())
	assert.Equal(t, "100", model.Entries[1].Value.ValueString())
}

func TestApiRuleToModel_NamedEntry(t *testing.T) {
	raw := map[string]interface{}{
		"relation": "AND",
		"entries": []interface{}{
			[]interface{}{"headers", []interface{}{"content-type", "application/json"}, "JSON type"},
		},
	}

	model, err := apiRuleToModel(raw)
	require.NoError(t, err)
	require.Len(t, model.Entries, 1)

	entry := model.Entries[0]
	assert.Equal(t, "headers", entry.Type.ValueString())
	assert.Equal(t, "content-type", entry.Name.ValueString())
	assert.Equal(t, "application/json", entry.Value.ValueString())
	assert.Equal(t, "JSON type", entry.Comment.ValueString())
}

func TestApiRuleToModel_WithGroup(t *testing.T) {
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

	model, err := apiRuleToModel(raw)
	require.NoError(t, err)
	require.NotNil(t, model)

	assert.Equal(t, "OR", model.Relation.ValueString())
	require.Len(t, model.Entries, 1)
	require.Len(t, model.Groups, 1)

	group := model.Groups[0]
	assert.Equal(t, "AND", group.Relation.ValueString())
	require.Len(t, group.Entries, 2)

	assert.Equal(t, "asn", group.Entries[0].Type.ValueString())
	assert.Equal(t, "100", group.Entries[0].Value.ValueString())

	assert.Equal(t, "cookies", group.Entries[1].Type.ValueString())
	assert.Equal(t, "session", group.Entries[1].Name.ValueString())
	assert.Equal(t, "abc", group.Entries[1].Value.ValueString())
	assert.Equal(t, "session cookie", group.Entries[1].Comment.ValueString())
}

func TestApiRuleToModel_Nil(t *testing.T) {
	model, err := apiRuleToModel(nil)
	require.NoError(t, err)
	assert.Nil(t, model)
}

func TestApiRuleToModel_InvalidType(t *testing.T) {
	_, err := apiRuleToModel("not a map")
	assert.Error(t, err)
}

func TestApiEntryToModel_TooFewElements(t *testing.T) {
	_, err := apiEntryToModel([]interface{}{"path"})
	assert.Error(t, err)
}

func TestRuleRoundTrip(t *testing.T) {
	original := &RuleModel{
		Relation: types.StringValue("OR"),
		Entries: []EntryModel{
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
		},
		Groups: []GroupModel{
			{
				Relation: types.StringValue("AND"),
				Entries: []EntryModel{
					{
						Type:    types.StringValue("asn"),
						Name:    types.StringNull(),
						Value:   types.StringValue("100"),
						Comment: types.StringValue(""),
					},
				},
			},
		},
	}

	// Convert to API format
	apiData := ruleModelToAPI(original)
	require.NotNil(t, apiData)

	// Convert back to model
	result, err := apiRuleToModel(apiData)
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, original.Relation.ValueString(), result.Relation.ValueString())
	require.Len(t, result.Entries, 2)
	require.Len(t, result.Groups, 1)

	// Verify simple entry round-trips
	assert.Equal(t, "path", result.Entries[0].Type.ValueString())
	assert.Equal(t, "/api/", result.Entries[0].Value.ValueString())
	assert.True(t, result.Entries[0].Name.IsNull())

	// Verify named entry round-trips
	assert.Equal(t, "headers", result.Entries[1].Type.ValueString())
	assert.Equal(t, "content-type", result.Entries[1].Name.ValueString())
	assert.Equal(t, "application/json", result.Entries[1].Value.ValueString())

	// Verify group round-trips
	assert.Equal(t, "AND", result.Groups[0].Relation.ValueString())
	require.Len(t, result.Groups[0].Entries, 1)
	assert.Equal(t, "asn", result.Groups[0].Entries[0].Type.ValueString())
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
			Entries: []EntryModel{
				{
					Type:    types.StringValue("cookies"),
					Name:    types.StringValue("test"),
					Value:   types.StringValue("ddddd"),
					Comment: types.StringValue("dddd"),
				},
			},
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
