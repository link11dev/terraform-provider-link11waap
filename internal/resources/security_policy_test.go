package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/link11/terraform-provider-link11waap/internal/client"
)

func TestBuildSecProfileMapEntry_AllFields(t *testing.T) {
	ctx := context.Background()

	rl, _ := types.ListValueFrom(ctx, types.StringType, []string{"rl-1", "rl-2"})
	ef, _ := types.ListValueFrom(ctx, types.StringType, []string{"ef-1"})

	m := SecProfileMapModel{
		ID:                         types.StringValue("entry-1"),
		Name:                       types.StringValue("Default"),
		Match:                      types.StringValue("/"),
		ACLProfile:                 types.StringValue("acl-default"),
		ACLProfileActive:           types.BoolValue(true),
		ContentFilterProfile:       types.StringValue("cf-default"),
		ContentFilterProfileActive: types.BoolValue(false),
		BackendService:             types.StringValue("backend-1"),
		Description:                types.StringValue("Default entry"),
		RateLimitRules:             rl,
		EdgeFunctions:              ef,
	}

	result := buildSecProfileMapEntry(ctx, m)

	if result.ID != "entry-1" {
		t.Errorf("expected ID='entry-1', got '%s'", result.ID)
	}
	if result.Name != "Default" {
		t.Errorf("expected Name='Default', got '%s'", result.Name)
	}
	if result.Match != "/" {
		t.Errorf("expected Match='/', got '%s'", result.Match)
	}
	if result.ACLProfile != "acl-default" {
		t.Errorf("expected ACLProfile='acl-default', got '%s'", result.ACLProfile)
	}
	if result.ACLProfileActive != true {
		t.Errorf("expected ACLProfileActive=true, got %v", result.ACLProfileActive)
	}
	if result.ContentFilterProfile != "cf-default" {
		t.Errorf("expected ContentFilterProfile='cf-default', got '%s'", result.ContentFilterProfile)
	}
	if result.ContentFilterProfileActive != false {
		t.Errorf("expected ContentFilterProfileActive=false, got %v", result.ContentFilterProfileActive)
	}
	if result.BackendService != "backend-1" {
		t.Errorf("expected BackendService='backend-1', got '%s'", result.BackendService)
	}
	if result.Description != "Default entry" {
		t.Errorf("expected Description='Default entry', got '%s'", result.Description)
	}
	if len(result.RateLimitRules) != 2 || result.RateLimitRules[0] != "rl-1" || result.RateLimitRules[1] != "rl-2" {
		t.Errorf("expected RateLimitRules=['rl-1','rl-2'], got %v", result.RateLimitRules)
	}
	if len(result.EdgeFunctions) != 1 || result.EdgeFunctions[0] != "ef-1" {
		t.Errorf("expected EdgeFunctions=['ef-1'], got %v", result.EdgeFunctions)
	}
}

func TestBuildSecProfileMapEntry_NullRateLimitRulesAndEdgeFunctions(t *testing.T) {
	ctx := context.Background()

	m := SecProfileMapModel{
		ID:                         types.StringValue("entry-2"),
		Name:                       types.StringValue("API"),
		Match:                      types.StringValue("/api"),
		ACLProfile:                 types.StringValue("acl-api"),
		ACLProfileActive:           types.BoolValue(true),
		ContentFilterProfile:       types.StringValue("cf-api"),
		ContentFilterProfileActive: types.BoolValue(true),
		BackendService:             types.StringValue("backend-2"),
		Description:                types.StringValue(""),
		RateLimitRules:             types.ListNull(types.StringType),
		EdgeFunctions:              types.ListNull(types.StringType),
	}

	result := buildSecProfileMapEntry(ctx, m)

	if result.RateLimitRules != nil {
		t.Errorf("expected RateLimitRules=nil, got %v", result.RateLimitRules)
	}
	if result.EdgeFunctions != nil {
		t.Errorf("expected EdgeFunctions=nil, got %v", result.EdgeFunctions)
	}
}

func TestBuildSecProfileMapEntry_EmptySlice(t *testing.T) {
	ctx := context.Background()

	models := []SecProfileMapModel{}

	if len(models) != 0 {
		t.Fatalf("expected empty slice, got %d items", len(models))
	}

	// Verify that buildSecurityPolicyAPIModel handles empty map correctly
	plan := &SecurityPolicyResourceModel{
		ID:          types.StringValue("sp-1"),
		Name:        types.StringValue("test"),
		Description: types.StringValue(""),
		Tags:        types.ListNull(types.StringType),
		Map:         models,
		ConfigID:    types.StringValue("config-1"),
	}

	sp := buildSecurityPolicyAPIModel(ctx, plan)
	if sp.Map != nil {
		t.Errorf("expected nil Map for empty slice, got %v", sp.Map)
	}
}

func TestBuildSecProfileMapEntries_MultipleEntries(t *testing.T) {
	ctx := context.Background()

	rl, _ := types.ListValueFrom(ctx, types.StringType, []string{"rl-1"})

	models := []SecProfileMapModel{
		{
			ID:                         types.StringValue("entry-1"),
			Name:                       types.StringValue("Default"),
			Match:                      types.StringValue("/"),
			ACLProfile:                 types.StringValue("acl-1"),
			ACLProfileActive:           types.BoolValue(true),
			ContentFilterProfile:       types.StringValue("cf-1"),
			ContentFilterProfileActive: types.BoolValue(true),
			BackendService:             types.StringValue("be-1"),
			Description:                types.StringValue("first"),
			RateLimitRules:             rl,
			EdgeFunctions:              types.ListNull(types.StringType),
		},
		{
			ID:                         types.StringValue("entry-2"),
			Name:                       types.StringValue("API"),
			Match:                      types.StringValue("/api"),
			ACLProfile:                 types.StringValue("acl-2"),
			ACLProfileActive:           types.BoolValue(false),
			ContentFilterProfile:       types.StringValue("cf-2"),
			ContentFilterProfileActive: types.BoolValue(false),
			BackendService:             types.StringValue("be-2"),
			Description:                types.StringValue("second"),
			RateLimitRules:             types.ListNull(types.StringType),
			EdgeFunctions:              types.ListNull(types.StringType),
		},
	}

	plan := &SecurityPolicyResourceModel{
		ID:          types.StringValue("sp-1"),
		Name:        types.StringValue("test"),
		Description: types.StringValue(""),
		Tags:        types.ListNull(types.StringType),
		Map:         models,
		// Session:     types.StringNull(),
		// c:  types.StringNull(),
		ConfigID: types.StringValue("config-1"),
	}

	sp := buildSecurityPolicyAPIModel(ctx, plan)
	if len(sp.Map) != 2 {
		t.Fatalf("expected 2 map entries, got %d", len(sp.Map))
	}

	if sp.Map[0].ID != "entry-1" {
		t.Errorf("expected first entry ID='entry-1', got '%s'", sp.Map[0].ID)
	}
	if sp.Map[0].Name != "Default" {
		t.Errorf("expected first entry Name='Default', got '%s'", sp.Map[0].Name)
	}
	if len(sp.Map[0].RateLimitRules) != 1 || sp.Map[0].RateLimitRules[0] != "rl-1" {
		t.Errorf("expected first entry RateLimitRules=['rl-1'], got %v", sp.Map[0].RateLimitRules)
	}
	if sp.Map[0].EdgeFunctions != nil {
		t.Errorf("expected first entry EdgeFunctions=nil, got %v", sp.Map[0].EdgeFunctions)
	}

	if sp.Map[1].ID != "entry-2" {
		t.Errorf("expected second entry ID='entry-2', got '%s'", sp.Map[1].ID)
	}
	if sp.Map[1].ACLProfileActive != false {
		t.Errorf("expected second entry ACLProfileActive=false, got %v", sp.Map[1].ACLProfileActive)
	}
}

func TestRoundTrip_APIToModelToAPI(t *testing.T) {
	ctx := context.Background()

	// Simulate API response
	apiEntries := []client.SecProfileMap{
		{
			ID:                         "rt-entry-1",
			Name:                       "RoundTrip",
			Match:                      "/test",
			ACLProfile:                 "acl-rt",
			ACLProfileActive:           true,
			ContentFilterProfile:       "cf-rt",
			ContentFilterProfileActive: false,
			BackendService:             "be-rt",
			Description:                "round-trip test",
			RateLimitRules:             []string{"rl-a", "rl-b"},
			EdgeFunctions:              []string{"ef-a"},
		},
		{
			ID:                         "rt-entry-2",
			Name:                       "RoundTrip2",
			Match:                      "/test2",
			ACLProfile:                 "acl-rt2",
			ACLProfileActive:           false,
			ContentFilterProfile:       "cf-rt2",
			ContentFilterProfileActive: true,
			BackendService:             "be-rt2",
			Description:                "",
			RateLimitRules:             nil,
			EdgeFunctions:              nil,
		},
	}

	// Convert API -> Model (simulating Read with nil-normalization)
	models := make([]SecProfileMapModel, len(apiEntries))
	for i, m := range apiEntries {
		model := SecProfileMapModel{
			ID:                         types.StringValue(m.ID),
			Name:                       types.StringValue(m.Name),
			Match:                      types.StringValue(m.Match),
			ACLProfile:                 types.StringValue(m.ACLProfile),
			ACLProfileActive:           types.BoolValue(m.ACLProfileActive),
			ContentFilterProfile:       types.StringValue(m.ContentFilterProfile),
			ContentFilterProfileActive: types.BoolValue(m.ContentFilterProfileActive),
			BackendService:             types.StringValue(m.BackendService),
			Description:                types.StringValue(m.Description),
		}

		// Normalize nil -> empty slice, matching the updated Read logic.
		rateLimitRules := m.RateLimitRules
		if rateLimitRules == nil {
			rateLimitRules = []string{}
		}
		rl, _ := types.ListValueFrom(ctx, types.StringType, rateLimitRules)
		model.RateLimitRules = rl

		edgeFunctions := m.EdgeFunctions
		if edgeFunctions == nil {
			edgeFunctions = []string{}
		}
		ef, _ := types.ListValueFrom(ctx, types.StringType, edgeFunctions)
		model.EdgeFunctions = ef

		models[i] = model
	}

	// Convert Model -> API (simulating Create/Update)
	results := make([]client.SecProfileMap, len(models))
	for i, m := range models {
		results[i] = buildSecProfileMapEntry(ctx, m)
	}

	// Verify round-trip for entry 1
	if results[0].ID != "rt-entry-1" {
		t.Errorf("round-trip[0]: expected ID='rt-entry-1', got '%s'", results[0].ID)
	}
	if results[0].Name != "RoundTrip" {
		t.Errorf("round-trip[0]: expected Name='RoundTrip', got '%s'", results[0].Name)
	}
	if results[0].ACLProfileActive != true {
		t.Errorf("round-trip[0]: expected ACLProfileActive=true, got %v", results[0].ACLProfileActive)
	}
	if len(results[0].RateLimitRules) != 2 || results[0].RateLimitRules[0] != "rl-a" || results[0].RateLimitRules[1] != "rl-b" {
		t.Errorf("round-trip[0]: expected RateLimitRules=['rl-a','rl-b'], got %v", results[0].RateLimitRules)
	}
	if len(results[0].EdgeFunctions) != 1 || results[0].EdgeFunctions[0] != "ef-a" {
		t.Errorf("round-trip[0]: expected EdgeFunctions=['ef-a'], got %v", results[0].EdgeFunctions)
	}

	// Verify round-trip for entry 2 (nil -> empty lists after normalization)
	if results[1].ID != "rt-entry-2" {
		t.Errorf("round-trip[1]: expected ID='rt-entry-2', got '%s'", results[1].ID)
	}
	if results[1].RateLimitRules == nil || len(results[1].RateLimitRules) != 0 {
		t.Errorf("round-trip[1]: expected RateLimitRules=[] (empty, non-nil), got %v", results[1].RateLimitRules)
	}
	if results[1].EdgeFunctions == nil || len(results[1].EdgeFunctions) != 0 {
		t.Errorf("round-trip[1]: expected EdgeFunctions=[] (empty, non-nil), got %v", results[1].EdgeFunctions)
	}
	if results[1].ContentFilterProfileActive != true {
		t.Errorf("round-trip[1]: expected ContentFilterProfileActive=true, got %v", results[1].ContentFilterProfileActive)
	}
}

func TestMergeDefaultMaps_AddsSiteLevelToPlan(t *testing.T) {
	siteLevel := SecProfileMapModel{
		ID:                         types.StringValue("__site_level__"),
		Name:                       types.StringValue("Site Level"),
		Match:                      types.StringValue("__site_level__"),
		ACLProfile:                 types.StringValue("__acldefault__"),
		ACLProfileActive:           types.BoolValue(false),
		ContentFilterProfile:       types.StringValue("__defaultcontentfilter__"),
		ContentFilterProfileActive: types.BoolValue(false),
		BackendService:             types.StringValue("__default__"),
		Description:                types.StringValue(""),
		RateLimitRules:             types.ListNull(types.StringType),
		EdgeFunctions:              types.ListNull(types.StringType),
	}
	apiEntry := SecProfileMapModel{
		ID:                         types.StringValue("api-entry"),
		Name:                       types.StringValue("API"),
		Match:                      types.StringValue("/api/"),
		ACLProfile:                 types.StringValue("acl-1"),
		ACLProfileActive:           types.BoolValue(true),
		ContentFilterProfile:       types.StringValue("cf-1"),
		ContentFilterProfileActive: types.BoolValue(true),
		BackendService:             types.StringValue("be-1"),
		Description:                types.StringValue(""),
		RateLimitRules:             types.ListNull(types.StringType),
		EdgeFunctions:              types.ListNull(types.StringType),
	}

	stateMaps := []SecProfileMapModel{apiEntry, siteLevel}
	planMaps := []SecProfileMapModel{apiEntry}

	result := mergeDefaultMaps(stateMaps, planMaps)

	if result == nil {
		t.Fatal("expected non-nil merged maps, got nil")
	}
	if len(result) != 2 {
		t.Errorf("expected 2 maps, got %d", len(result))
	}

	found := false
	for _, m := range result {
		if m.ID.ValueString() == "__site_level__" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected __site_level__ map to be present in merged result")
	}
}

func TestMergeDefaultMaps_NoChangeWhenAlreadyInPlan(t *testing.T) {
	siteLevel := SecProfileMapModel{
		ID:                   types.StringValue("__site_level__"),
		Name:                 types.StringValue("Site Level"),
		Match:                types.StringValue("__site_level__"),
		ACLProfile:           types.StringValue("__acldefault__"),
		BackendService:       types.StringValue("__default__"),
		ContentFilterProfile: types.StringValue("__defaultcontentfilter__"),
		RateLimitRules:       types.ListNull(types.StringType),
		EdgeFunctions:        types.ListNull(types.StringType),
	}

	stateMaps := []SecProfileMapModel{siteLevel}
	planMaps := []SecProfileMapModel{siteLevel}

	result := mergeDefaultMaps(stateMaps, planMaps)
	if result != nil {
		t.Errorf("expected nil (no changes needed), got %v", result)
	}
}

func TestMergeDefaultMaps_NoChangeWhenNotInState(t *testing.T) {
	apiEntry := SecProfileMapModel{
		ID:                   types.StringValue("api-entry"),
		Name:                 types.StringValue("API"),
		Match:                types.StringValue("/api/"),
		ACLProfile:           types.StringValue("acl-1"),
		BackendService:       types.StringValue("be-1"),
		ContentFilterProfile: types.StringValue("cf-1"),
		RateLimitRules:       types.ListNull(types.StringType),
		EdgeFunctions:        types.ListNull(types.StringType),
	}

	stateMaps := []SecProfileMapModel{apiEntry}
	planMaps := []SecProfileMapModel{apiEntry}

	result := mergeDefaultMaps(stateMaps, planMaps)
	if result != nil {
		t.Errorf("expected nil (no default map in state), got %v", result)
	}
}

func TestMergeDefaultMaps_EmptyStateMaps(t *testing.T) {
	result := mergeDefaultMaps(nil, nil)
	if result != nil {
		t.Errorf("expected nil for empty state, got %v", result)
	}
}

// TestReadNormalization_EmptyAPIListsMatchSchemaDefault verifies that when the
// API returns empty slices (or nil) for edge_functions/rate_limit_rules, the
// Read logic produces an empty (non-null) types.List. This must match the schema
// default so that the set-nested-block hash is identical and no spurious diff
// is generated.
func TestReadNormalization_EmptyAPIListsMatchSchemaDefault(t *testing.T) {
	ctx := context.Background()

	// Simulate API response where fields are empty slices (common) or nil.
	cases := []struct {
		name           string
		rateLimitRules []string
		edgeFunctions  []string
	}{
		{"both empty slices", []string{}, []string{}},
		{"both nil", nil, nil},
		{"mixed nil and empty", nil, []string{}},
	}

	emptyList := types.ListValueMust(types.StringType, []attr.Value{})

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Apply the same normalization as the Read function.
			rlRules := tc.rateLimitRules
			if rlRules == nil {
				rlRules = []string{}
			}
			rl, _ := types.ListValueFrom(ctx, types.StringType, rlRules)

			efFuncs := tc.edgeFunctions
			if efFuncs == nil {
				efFuncs = []string{}
			}
			ef, _ := types.ListValueFrom(ctx, types.StringType, efFuncs)

			if !rl.Equal(emptyList) {
				t.Errorf("rate_limit_rules: expected empty list (equal to schema default), got %v", rl)
			}
			if !ef.Equal(emptyList) {
				t.Errorf("edge_functions: expected empty list (equal to schema default), got %v", ef)
			}

			// Verify they are NOT null (which would cause the set-hash mismatch).
			if rl.IsNull() {
				t.Error("rate_limit_rules should not be null after normalization")
			}
			if ef.IsNull() {
				t.Error("edge_functions should not be null after normalization")
			}
		})
	}
}

// Additional security policy tests

func TestNewSecurityPolicyResource(t *testing.T) {
	r := NewSecurityPolicyResource()
	if r == nil {
		t.Fatal("expected non-nil resource")
	}
	_, ok := r.(*SecurityPolicyResource)
	if !ok {
		t.Fatal("expected *SecurityPolicyResource")
	}
}

func TestSecurityPolicyResource_Metadata(t *testing.T) {
	r := &SecurityPolicyResource{}
	ctx := context.Background()

	req := metadataReq("link11waap")
	resp := metadataResp()
	r.Metadata(ctx, req, resp)

	if resp.TypeName != "link11waap_security_policy" {
		t.Errorf("expected 'link11waap_security_policy', got %q", resp.TypeName)
	}
}

func TestSecurityPolicyResource_Schema(t *testing.T) {
	r := &SecurityPolicyResource{}
	ctx := context.Background()

	sReq := schemaReq()
	sResp := schemaResp()
	r.Schema(ctx, sReq, sResp)

	schema := sResp.Schema
	if len(schema.Attributes) == 0 {
		t.Fatal("expected non-empty attributes")
	}

	expectedAttrs := []string{"config_id", "id", "name", "description", "tags"}
	for _, a := range expectedAttrs {
		if _, ok := schema.Attributes[a]; !ok {
			t.Errorf("expected attribute %q in schema", a)
		}
	}

	expectedBlocks := []string{"map", "session", "session_ids"}
	for _, b := range expectedBlocks {
		if _, ok := schema.Blocks[b]; !ok {
			t.Errorf("expected block %q in schema", b)
		}
	}
}

func TestSecurityPolicyResource_Configure_NilProvider(t *testing.T) {
	r := &SecurityPolicyResource{}
	ctx := context.Background()

	req := configureReq(nil)
	resp := configureResp()
	r.Configure(ctx, req, resp)

	if r.client != nil {
		t.Error("expected nil client for nil provider data")
	}
}

func TestSecurityPolicyResource_ImportState_Valid(t *testing.T) {
	r := &SecurityPolicyResource{}
	resp := testImportState(t, r, "config123/sp456")

	if resp.Diagnostics.HasError() {
		t.Errorf("expected no errors, got: %v", resp.Diagnostics)
	}
}

func TestSecurityPolicyResource_ImportState_Invalid(t *testing.T) {
	r := &SecurityPolicyResource{}
	resp := testImportState(t, r, "invalid")

	if !resp.Diagnostics.HasError() {
		t.Error("expected error for invalid import ID")
	}
}

func TestSecurityPolicyResource_ImportState_TooManyParts(t *testing.T) {
	r := &SecurityPolicyResource{}
	resp := testImportState(t, r, "a/b/c")

	if !resp.Diagnostics.HasError() {
		t.Error("expected error for too many parts")
	}
}

func TestParseSessionKeys_Nil(t *testing.T) {
	result := parseSessionKeys(nil)
	if len(result) != 0 {
		t.Fatalf("expected empty slice for nil input, got %d items", len(result))
	}
}

func TestParseSessionKeys_InvalidType(t *testing.T) {
	result := parseSessionKeys("not a slice")
	if len(result) != 0 {
		t.Fatalf("expected empty slice for invalid type, got %d items", len(result))
	}
}

func TestParseSessionKeys_AllKeyTypes(t *testing.T) {
	raw := []interface{}{
		map[string]interface{}{"attrs": "ip"},
		map[string]interface{}{"args": "user_id"},
		map[string]interface{}{"plugins": "jwt.sub"},
		map[string]interface{}{"cookies": "session"},
		map[string]interface{}{"headers": "authorization"},
	}

	result := parseSessionKeys(raw)
	if len(result) != 5 {
		t.Fatalf("expected 5 keys, got %d", len(result))
	}

	if result[0].Attrs.ValueString() != "ip" {
		t.Errorf("expected attrs='ip', got '%s'", result[0].Attrs.ValueString())
	}
	if result[1].Args.ValueString() != "user_id" {
		t.Errorf("expected args='user_id', got '%s'", result[1].Args.ValueString())
	}
	if result[2].Plugins.ValueString() != "jwt.sub" {
		t.Errorf("expected plugins='jwt.sub', got '%s'", result[2].Plugins.ValueString())
	}
	if result[3].Cookies.ValueString() != "session" {
		t.Errorf("expected cookies='session', got '%s'", result[3].Cookies.ValueString())
	}
	if result[4].Headers.ValueString() != "authorization" {
		t.Errorf("expected headers='authorization', got '%s'", result[4].Headers.ValueString())
	}
}

func TestParseSessionKeys_SkipsInvalidEntries(t *testing.T) {
	raw := []interface{}{
		"not a map",
		42,
		map[string]interface{}{"attrs": "ip"},
	}

	result := parseSessionKeys(raw)
	if len(result) != 1 {
		t.Fatalf("expected 1 key (skipping invalid), got %d", len(result))
	}
}

func TestParseSessionKeys_UnknownKeyType(t *testing.T) {
	raw := []interface{}{
		map[string]interface{}{"unknown": "value"},
	}

	result := parseSessionKeys(raw)
	if len(result) != 1 {
		t.Fatalf("expected 1 key, got %d", len(result))
	}
	if !result[0].Attrs.IsNull() || !result[0].Args.IsNull() || !result[0].Plugins.IsNull() ||
		!result[0].Cookies.IsNull() || !result[0].Headers.IsNull() {
		t.Error("expected all fields to be null for unknown key type")
	}
}

func TestBuildSessionKeys_AllKeyTypes(t *testing.T) {
	keys := []SessionKeyModel{
		{Attrs: types.StringValue("ip"), Args: types.StringNull(), Plugins: types.StringNull(), Cookies: types.StringNull(), Headers: types.StringNull()},
		{Attrs: types.StringNull(), Args: types.StringValue("q"), Plugins: types.StringNull(), Cookies: types.StringNull(), Headers: types.StringNull()},
		{Attrs: types.StringNull(), Args: types.StringNull(), Plugins: types.StringValue("jwt.sub"), Cookies: types.StringNull(), Headers: types.StringNull()},
		{Attrs: types.StringNull(), Args: types.StringNull(), Plugins: types.StringNull(), Cookies: types.StringValue("sid"), Headers: types.StringNull()},
		{Attrs: types.StringNull(), Args: types.StringNull(), Plugins: types.StringNull(), Cookies: types.StringNull(), Headers: types.StringValue("auth")},
	}

	result := buildSessionKeys(keys)
	if len(result) != 5 {
		t.Fatalf("expected 5 results, got %d", len(result))
	}

	expected := []struct {
		key   string
		value string
	}{
		{"attrs", "ip"},
		{"args", "q"},
		{"plugins", "jwt.sub"},
		{"cookies", "sid"},
		{"headers", "auth"},
	}

	for i, exp := range expected {
		v, ok := result[i][exp.key]
		if !ok {
			t.Errorf("result[%d]: expected key '%s' not found", i, exp.key)
			continue
		}
		if v != exp.value {
			t.Errorf("result[%d]: expected %s='%s', got '%s'", i, exp.key, exp.value, v)
		}
	}
}

func TestBuildSessionKeys_Empty(t *testing.T) {
	result := buildSessionKeys([]SessionKeyModel{})
	if len(result) != 0 {
		t.Fatalf("expected empty result, got %d items", len(result))
	}
}

func TestBuildSessionKeys_AllNull(t *testing.T) {
	keys := []SessionKeyModel{
		{Attrs: types.StringNull(), Args: types.StringNull(), Plugins: types.StringNull(), Cookies: types.StringNull(), Headers: types.StringNull()},
	}
	result := buildSessionKeys(keys)
	if len(result) != 0 {
		t.Fatalf("expected empty result for all-null key, got %d items", len(result))
	}
}

func TestCountSessionKeyFields(t *testing.T) {
	tests := []struct {
		name string
		key  SessionKeyModel
		want int
	}{
		{
			name: "none set",
			key:  SessionKeyModel{Attrs: types.StringNull(), Args: types.StringNull(), Plugins: types.StringNull(), Cookies: types.StringNull(), Headers: types.StringNull()},
			want: 0,
		},
		{
			name: "only attrs",
			key:  SessionKeyModel{Attrs: types.StringValue("ip"), Args: types.StringNull(), Plugins: types.StringNull(), Cookies: types.StringNull(), Headers: types.StringNull()},
			want: 1,
		},
		{
			name: "only args",
			key:  SessionKeyModel{Attrs: types.StringNull(), Args: types.StringValue("q"), Plugins: types.StringNull(), Cookies: types.StringNull(), Headers: types.StringNull()},
			want: 1,
		},
		{
			name: "only plugins",
			key:  SessionKeyModel{Attrs: types.StringNull(), Args: types.StringNull(), Plugins: types.StringValue("jwt.x"), Cookies: types.StringNull(), Headers: types.StringNull()},
			want: 1,
		},
		{
			name: "only cookies",
			key:  SessionKeyModel{Attrs: types.StringNull(), Args: types.StringNull(), Plugins: types.StringNull(), Cookies: types.StringValue("sid"), Headers: types.StringNull()},
			want: 1,
		},
		{
			name: "only headers",
			key:  SessionKeyModel{Attrs: types.StringNull(), Args: types.StringNull(), Plugins: types.StringNull(), Cookies: types.StringNull(), Headers: types.StringValue("auth")},
			want: 1,
		},
		{
			name: "two set",
			key:  SessionKeyModel{Attrs: types.StringValue("ip"), Args: types.StringValue("q"), Plugins: types.StringNull(), Cookies: types.StringNull(), Headers: types.StringNull()},
			want: 2,
		},
		{
			name: "all set",
			key:  SessionKeyModel{Attrs: types.StringValue("ip"), Args: types.StringValue("q"), Plugins: types.StringValue("jwt.x"), Cookies: types.StringValue("s"), Headers: types.StringValue("h")},
			want: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := countSessionKeyFields(tt.key)
			if got != tt.want {
				t.Errorf("expected %d, got %d", tt.want, got)
			}
		})
	}
}

func TestSessionKeysRoundTrip_ParseAndBuild(t *testing.T) {
	apiResponse := []interface{}{
		map[string]interface{}{"attrs": "ip"},
		map[string]interface{}{"cookies": "session_id"},
	}

	parsed := parseSessionKeys(apiResponse)
	if len(parsed) != 2 {
		t.Fatalf("expected 2 parsed keys, got %d", len(parsed))
	}

	built := buildSessionKeys(parsed)
	if len(built) != 2 {
		t.Fatalf("expected 2 built keys, got %d", len(built))
	}

	if v, ok := built[0]["attrs"]; !ok || v != "ip" {
		t.Errorf("roundtrip[0]: expected attrs='ip', got %v", built[0])
	}
	if v, ok := built[1]["cookies"]; !ok || v != "session_id" {
		t.Errorf("roundtrip[1]: expected cookies='session_id', got %v", built[1])
	}
}

func TestBuildSecurityPolicyAPIModel_WithTags(t *testing.T) {
	ctx := context.Background()

	tags, _ := types.ListValueFrom(ctx, types.StringType, []string{"tag1", "tag2"})
	plan := &SecurityPolicyResourceModel{
		ID:          types.StringValue("sp-1"),
		Name:        types.StringValue("test-sp"),
		Description: types.StringValue("desc"),
		Tags:        tags,
		Map:         []SecProfileMapModel{},
		Session: []SessionKeyModel{
			{Attrs: types.StringValue("ip"), Args: types.StringNull(), Plugins: types.StringNull(), Cookies: types.StringNull(), Headers: types.StringNull()},
		},
		SessionIDs: []SessionKeyModel{},
		ConfigID:   types.StringValue("cfg1"),
	}

	sp := buildSecurityPolicyAPIModel(ctx, plan)

	if sp.Name != "test-sp" {
		t.Errorf("expected Name='test-sp', got '%s'", sp.Name)
	}
	if len(sp.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(sp.Tags))
	}
	if len(sp.Session.([]map[string]string)) != 1 {
		t.Errorf("expected 1 session key, got %d", len(sp.Session.([]map[string]string)))
	}
	// SessionIDs should be empty (not nil) when no session_ids blocks
	if len(sp.SessionIDs.([]map[string]string)) != 0 {
		t.Errorf("expected empty SessionIDs, got %v", sp.SessionIDs)
	}
}

func TestBuildSecurityPolicyAPIModel_NullTags(t *testing.T) {
	ctx := context.Background()

	plan := &SecurityPolicyResourceModel{
		ID:          types.StringValue("sp-2"),
		Name:        types.StringValue("test"),
		Description: types.StringValue(""),
		Tags:        types.ListNull(types.StringType),
		Map:         []SecProfileMapModel{},
		Session:     []SessionKeyModel{},
		SessionIDs:  []SessionKeyModel{},
		ConfigID:    types.StringValue("cfg1"),
	}

	sp := buildSecurityPolicyAPIModel(ctx, plan)

	if sp.Tags != nil {
		t.Errorf("expected nil tags, got %v", sp.Tags)
	}
}

func TestBuildSecurityPolicyAPIModel_WithSessionIDs(t *testing.T) {
	ctx := context.Background()

	plan := &SecurityPolicyResourceModel{
		ID:          types.StringValue("sp-3"),
		Name:        types.StringValue("test"),
		Description: types.StringValue(""),
		Tags:        types.ListNull(types.StringType),
		Map:         []SecProfileMapModel{},
		Session: []SessionKeyModel{
			{Attrs: types.StringValue("ip"), Args: types.StringNull(), Plugins: types.StringNull(), Cookies: types.StringNull(), Headers: types.StringNull()},
		},
		SessionIDs: []SessionKeyModel{
			{Attrs: types.StringNull(), Args: types.StringNull(), Plugins: types.StringNull(), Cookies: types.StringValue("sid"), Headers: types.StringNull()},
		},
		ConfigID: types.StringValue("cfg1"),
	}

	sp := buildSecurityPolicyAPIModel(ctx, plan)

	sessionIDs := sp.SessionIDs.([]map[string]string)
	if len(sessionIDs) != 1 {
		t.Fatalf("expected 1 session_id key, got %d", len(sessionIDs))
	}
	if v, ok := sessionIDs[0]["cookies"]; !ok || v != "sid" {
		t.Errorf("expected session_id cookies='sid', got %v", sessionIDs[0])
	}
}
