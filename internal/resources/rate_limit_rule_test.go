package resources

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/link11/terraform-provider-link11waap/internal/client"
)

func TestParseRateLimitKeys_Nil(t *testing.T) {
	result := parseRateLimitKeys(nil)
	if len(result) != 0 {
		t.Fatalf("expected empty slice for nil input, got %d items", len(result))
	}
}

func TestParseRateLimitKeys_InvalidType(t *testing.T) {
	result := parseRateLimitKeys("not a slice")
	if len(result) != 0 {
		t.Fatalf("expected empty slice for invalid type, got %d items", len(result))
	}
}

func TestParseRateLimitKeys_AllKeyTypes(t *testing.T) {
	raw := []interface{}{
		map[string]interface{}{"attrs": "session"},
		map[string]interface{}{"args": "query_param"},
		map[string]interface{}{"plugins": "jwt.claim"},
		map[string]interface{}{"cookies": "session_id"},
		map[string]interface{}{"headers": "x-api-key"},
	}

	result := parseRateLimitKeys(raw)
	if len(result) != 5 {
		t.Fatalf("expected 5 keys, got %d", len(result))
	}

	// attrs
	if result[0].Attrs.ValueString() != "session" {
		t.Errorf("expected attrs='session', got '%s'", result[0].Attrs.ValueString())
	}
	if !result[0].Args.IsNull() || !result[0].Plugins.IsNull() || !result[0].Cookies.IsNull() || !result[0].Headers.IsNull() {
		t.Error("expected other fields to be null for attrs key")
	}

	// args
	if result[1].Args.ValueString() != "query_param" {
		t.Errorf("expected args='query_param', got '%s'", result[1].Args.ValueString())
	}
	if !result[1].Attrs.IsNull() {
		t.Error("expected attrs to be null for args key")
	}

	// plugins
	if result[2].Plugins.ValueString() != "jwt.claim" {
		t.Errorf("expected plugins='jwt.claim', got '%s'", result[2].Plugins.ValueString())
	}
	if !result[2].Attrs.IsNull() {
		t.Error("expected attrs to be null for plugins key")
	}

	// cookies
	if result[3].Cookies.ValueString() != "session_id" {
		t.Errorf("expected cookies='session_id', got '%s'", result[3].Cookies.ValueString())
	}
	if !result[3].Attrs.IsNull() {
		t.Error("expected attrs to be null for cookies key")
	}

	// headers
	if result[4].Headers.ValueString() != "x-api-key" {
		t.Errorf("expected headers='x-api-key', got '%s'", result[4].Headers.ValueString())
	}
	if !result[4].Attrs.IsNull() {
		t.Error("expected attrs to be null for headers key")
	}
}

func TestParseRateLimitKeys_DuplicateKeyTypes(t *testing.T) {
	raw := []interface{}{
		map[string]interface{}{"attrs": "session"},
		map[string]interface{}{"attrs": "ip"},
	}

	result := parseRateLimitKeys(raw)
	if len(result) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(result))
	}
	if result[0].Attrs.ValueString() != "session" {
		t.Errorf("expected first attrs='session', got '%s'", result[0].Attrs.ValueString())
	}
	if result[1].Attrs.ValueString() != "ip" {
		t.Errorf("expected second attrs='ip', got '%s'", result[1].Attrs.ValueString())
	}
}

func TestParseRateLimitKeys_SkipsInvalidEntries(t *testing.T) {
	raw := []interface{}{
		"not a map",
		map[string]interface{}{"attrs": "session"},
	}

	result := parseRateLimitKeys(raw)
	if len(result) != 1 {
		t.Fatalf("expected 1 key (skipping invalid), got %d", len(result))
	}
	if result[0].Attrs.ValueString() != "session" {
		t.Errorf("expected attrs='session', got '%s'", result[0].Attrs.ValueString())
	}
}

func TestParseRateLimitKeys_UnknownKeyType(t *testing.T) {
	raw := []interface{}{
		map[string]interface{}{"unknown_field": "value"},
	}

	result := parseRateLimitKeys(raw)
	if len(result) != 1 {
		t.Fatalf("expected 1 key, got %d", len(result))
	}
	// All fields should be null since "unknown_field" is not recognized
	if !result[0].Attrs.IsNull() || !result[0].Args.IsNull() || !result[0].Plugins.IsNull() || !result[0].Cookies.IsNull() || !result[0].Headers.IsNull() {
		t.Error("expected all fields to be null for unknown key type")
	}
}

func TestBuildRateLimitKeys_AllKeyTypes(t *testing.T) {
	keys := []RateLimitKeyModel{
		{Attrs: types.StringValue("session"), Args: types.StringNull(), Plugins: types.StringNull(), Cookies: types.StringNull(), Headers: types.StringNull()},
		{Attrs: types.StringNull(), Args: types.StringValue("q"), Plugins: types.StringNull(), Cookies: types.StringNull(), Headers: types.StringNull()},
		{Attrs: types.StringNull(), Args: types.StringNull(), Plugins: types.StringValue("jwt.sub"), Cookies: types.StringNull(), Headers: types.StringNull()},
		{Attrs: types.StringNull(), Args: types.StringNull(), Plugins: types.StringNull(), Cookies: types.StringValue("sid"), Headers: types.StringNull()},
		{Attrs: types.StringNull(), Args: types.StringNull(), Plugins: types.StringNull(), Cookies: types.StringNull(), Headers: types.StringValue("x-key")},
	}

	result := buildRateLimitKeys(keys)
	if len(result) != 5 {
		t.Fatalf("expected 5 results, got %d", len(result))
	}

	expected := []struct {
		key   string
		value string
	}{
		{"attrs", "session"},
		{"args", "q"},
		{"plugins", "jwt.sub"},
		{"cookies", "sid"},
		{"headers", "x-key"},
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
		if len(result[i]) != 1 {
			t.Errorf("result[%d]: expected exactly 1 key in map, got %d", i, len(result[i]))
		}
	}
}

func TestBuildRateLimitKeys_Empty(t *testing.T) {
	result := buildRateLimitKeys([]RateLimitKeyModel{})
	if len(result) != 0 {
		t.Fatalf("expected empty result, got %d items", len(result))
	}
}

func TestBuildRateLimitKeys_AllNull(t *testing.T) {
	keys := []RateLimitKeyModel{
		{Attrs: types.StringNull(), Args: types.StringNull(), Plugins: types.StringNull(), Cookies: types.StringNull(), Headers: types.StringNull()},
	}
	result := buildRateLimitKeys(keys)
	if len(result) != 0 {
		t.Fatalf("expected empty result for all-null key, got %d items", len(result))
	}
}

func TestRoundTrip_ParseAndBuild(t *testing.T) {
	// Simulate API response
	apiResponse := []interface{}{
		map[string]interface{}{"attrs": "session"},
		map[string]interface{}{"args": "ssss"},
		map[string]interface{}{"plugins": "jwt.ddddd"},
		map[string]interface{}{"cookies": "ddddd"},
		map[string]interface{}{"headers": "zzzzz"},
	}

	// Parse from API
	parsed := parseRateLimitKeys(apiResponse)
	if len(parsed) != 5 {
		t.Fatalf("expected 5 parsed keys, got %d", len(parsed))
	}

	// Build back to API format
	built := buildRateLimitKeys(parsed)
	if len(built) != 5 {
		t.Fatalf("expected 5 built keys, got %d", len(built))
	}

	// Verify round-trip
	expectations := []struct {
		key   string
		value string
	}{
		{"attrs", "session"},
		{"args", "ssss"},
		{"plugins", "jwt.ddddd"},
		{"cookies", "ddddd"},
		{"headers", "zzzzz"},
	}

	for i, exp := range expectations {
		v, ok := built[i][exp.key]
		if !ok {
			t.Errorf("round-trip[%d]: expected key '%s' not found", i, exp.key)
			continue
		}
		if v != exp.value {
			t.Errorf("round-trip[%d]: expected %s='%s', got '%s'", i, exp.key, exp.value, v)
		}
	}
}

func TestPluginsValidation_ValidPrefix(t *testing.T) {
	validValues := []string{"jwt.claim", "jwt.sub", "jwt.iss", "jwt.a"}
	for _, v := range validValues {
		if !strings.HasPrefix(v, "jwt.") {
			t.Errorf("expected '%s' to have jwt. prefix", v)
		}
	}
}

func TestPluginsValidation_InvalidPrefix(t *testing.T) {
	invalidValues := []string{"", "jwt", "jw.claim", "JWT.claim", "other.value", "jwtclaim"}
	for _, v := range invalidValues {
		if strings.HasPrefix(v, "jwt.") {
			t.Errorf("expected '%s' to NOT have jwt. prefix", v)
		}
	}
}

func TestKeyBlockValidation_ExactlyOneField(t *testing.T) {
	tests := []struct {
		name      string
		key       RateLimitKeyModel
		wantValid bool
	}{
		{
			name:      "only attrs set",
			key:       RateLimitKeyModel{Attrs: types.StringValue("session"), Args: types.StringNull(), Plugins: types.StringNull(), Cookies: types.StringNull(), Headers: types.StringNull()},
			wantValid: true,
		},
		{
			name:      "only args set",
			key:       RateLimitKeyModel{Attrs: types.StringNull(), Args: types.StringValue("q"), Plugins: types.StringNull(), Cookies: types.StringNull(), Headers: types.StringNull()},
			wantValid: true,
		},
		{
			name:      "only plugins set",
			key:       RateLimitKeyModel{Attrs: types.StringNull(), Args: types.StringNull(), Plugins: types.StringValue("jwt.x"), Cookies: types.StringNull(), Headers: types.StringNull()},
			wantValid: true,
		},
		{
			name:      "only cookies set",
			key:       RateLimitKeyModel{Attrs: types.StringNull(), Args: types.StringNull(), Plugins: types.StringNull(), Cookies: types.StringValue("sid"), Headers: types.StringNull()},
			wantValid: true,
		},
		{
			name:      "only headers set",
			key:       RateLimitKeyModel{Attrs: types.StringNull(), Args: types.StringNull(), Plugins: types.StringNull(), Cookies: types.StringNull(), Headers: types.StringValue("x-key")},
			wantValid: true,
		},
		{
			name:      "none set",
			key:       RateLimitKeyModel{Attrs: types.StringNull(), Args: types.StringNull(), Plugins: types.StringNull(), Cookies: types.StringNull(), Headers: types.StringNull()},
			wantValid: false,
		},
		{
			name:      "two set - attrs and args",
			key:       RateLimitKeyModel{Attrs: types.StringValue("ip"), Args: types.StringValue("q"), Plugins: types.StringNull(), Cookies: types.StringNull(), Headers: types.StringNull()},
			wantValid: false,
		},
		{
			name:      "all set",
			key:       RateLimitKeyModel{Attrs: types.StringValue("ip"), Args: types.StringValue("q"), Plugins: types.StringValue("jwt.x"), Cookies: types.StringValue("s"), Headers: types.StringValue("h")},
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setCount := countSetFields(tt.key)
			isValid := setCount == 1
			if isValid != tt.wantValid {
				t.Errorf("expected valid=%v, got valid=%v (setCount=%d)", tt.wantValid, isValid, setCount)
			}
		})
	}
}

// countSetFields counts the number of non-null, non-unknown fields in a RateLimitKeyModel.
// This mirrors the validation logic in ValidateConfig.
func countSetFields(k RateLimitKeyModel) int {
	count := 0
	if !k.Attrs.IsNull() && !k.Attrs.IsUnknown() {
		count++
	}
	if !k.Args.IsNull() && !k.Args.IsUnknown() {
		count++
	}
	if !k.Plugins.IsNull() && !k.Plugins.IsUnknown() {
		count++
	}
	if !k.Cookies.IsNull() && !k.Cookies.IsUnknown() {
		count++
	}
	if !k.Headers.IsNull() && !k.Headers.IsUnknown() {
		count++
	}
	return count
}

// Additional rate limit rule tests

func TestNewRateLimitRuleResource(t *testing.T) {
	r := NewRateLimitRuleResource()
	if r == nil {
		t.Fatal("expected non-nil resource")
	}
	_, ok := r.(*RateLimitRuleResource)
	if !ok {
		t.Fatal("expected *RateLimitRuleResource")
	}
}

func TestRateLimitRuleResource_Metadata(t *testing.T) {
	r := &RateLimitRuleResource{}
	ctx := context.Background()

	req := metadataReq("link11waap")
	resp := metadataResp()
	r.Metadata(ctx, req, resp)

	if resp.TypeName != "link11waap_rate_limit_rule" {
		t.Errorf("expected 'link11waap_rate_limit_rule', got %q", resp.TypeName)
	}
}

func TestRateLimitRuleResource_Schema(t *testing.T) {
	r := &RateLimitRuleResource{}
	ctx := context.Background()

	sReq := schemaReq()
	sResp := schemaResp()
	r.Schema(ctx, sReq, sResp)

	schema := sResp.Schema
	if len(schema.Attributes) == 0 {
		t.Fatal("expected non-empty attributes")
	}

	expectedAttrs := []string{
		"config_id", "id", "name", "description", "global", "active",
		"timeframe", "threshold", "ttl", "action", "is_action_ban",
		"tags", "pairwith",
	}
	for _, a := range expectedAttrs {
		if _, ok := schema.Attributes[a]; !ok {
			t.Errorf("expected attribute %q in schema", a)
		}
	}

	expectedBlocks := []string{"key", "include", "exclude"}
	for _, b := range expectedBlocks {
		if _, ok := schema.Blocks[b]; !ok {
			t.Errorf("expected block %q in schema", b)
		}
	}
}

func TestRateLimitRuleResource_Configure_NilProvider(t *testing.T) {
	r := &RateLimitRuleResource{}
	ctx := context.Background()

	req := configureReq(nil)
	resp := configureResp()
	r.Configure(ctx, req, resp)

	if r.client != nil {
		t.Error("expected nil client for nil provider data")
	}
}

func TestRateLimitRuleResource_ImportState_Valid(t *testing.T) {
	r := &RateLimitRuleResource{}
	resp := testImportState(t, r, "config123/rl456")

	if resp.Diagnostics.HasError() {
		t.Errorf("expected no errors, got: %v", resp.Diagnostics)
	}
}

func TestRateLimitRuleResource_ImportState_Invalid(t *testing.T) {
	r := &RateLimitRuleResource{}
	resp := testImportState(t, r, "invalid")

	if !resp.Diagnostics.HasError() {
		t.Error("expected error for invalid import ID")
	}
}

func TestRateLimitRuleResource_ImportState_TooManyParts(t *testing.T) {
	r := &RateLimitRuleResource{}
	resp := testImportState(t, r, "a/b/c")

	if !resp.Diagnostics.HasError() {
		t.Error("expected error for too many parts")
	}
}

func TestTagFilterToSet_WithTags(t *testing.T) {
	ctx := context.Background()
	filter := client.RateLimitTagFilter{
		Relation: "OR",
		Tags:     []string{"tag1", "tag2"},
	}

	result, diags := tagFilterToSet(ctx, filter)

	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	if result.IsNull() {
		t.Fatal("expected non-null set")
	}
}

func TestTagFilterToSet_NilTags(t *testing.T) {
	ctx := context.Background()
	filter := client.RateLimitTagFilter{
		Relation: "AND",
		Tags:     nil,
	}

	result, diags := tagFilterToSet(ctx, filter)

	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	if result.IsNull() {
		t.Fatal("expected non-null set even with nil tags")
	}
}

func TestTagFilterToSet_EmptyTags(t *testing.T) {
	ctx := context.Background()
	filter := client.RateLimitTagFilter{
		Relation: "OR",
		Tags:     []string{},
	}

	result, diags := tagFilterToSet(ctx, filter)

	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	if result.IsNull() {
		t.Fatal("expected non-null set")
	}
}

func TestBuildRateLimitRuleAPIModel_BasicFields(t *testing.T) {
	ctx := context.Background()

	// Create include/exclude sets
	tagsList, _ := types.ListValueFrom(ctx, types.StringType, []string{"tag1"})
	includeObj, _ := types.ObjectValue(tagFilterAttrTypes, map[string]attr.Value{
		"relation": types.StringValue("OR"),
		"tags":     tagsList,
	})
	includeSet, _ := types.SetValue(types.ObjectType{AttrTypes: tagFilterAttrTypes}, []attr.Value{includeObj})
	excludeSet, _ := types.SetValue(types.ObjectType{AttrTypes: tagFilterAttrTypes}, []attr.Value{includeObj})

	plan := &RateLimitRuleResourceModel{
		ID:          types.StringValue("rl-1"),
		Name:        types.StringValue("test-rule"),
		Description: types.StringValue("desc"),
		Global:      types.BoolValue(false),
		Active:      types.BoolValue(true),
		Timeframe:   types.Int64Value(60),
		Threshold:   types.Int64Value(100),
		TTL:         types.Int64Value(300),
		Action:      types.StringValue("action-monitor"),
		IsActionBan: types.BoolValue(false),
		Tags:        types.ListNull(types.StringType),
		Key: []RateLimitKeyModel{
			{Attrs: types.StringValue("ip"), Args: types.StringNull(), Plugins: types.StringNull(), Cookies: types.StringNull(), Headers: types.StringNull()},
		},
		Pairwith: types.StringValue(`{"self":"self"}`),
		Include:  includeSet,
		Exclude:  excludeSet,
		ConfigID: types.StringValue("cfg1"),
	}

	rule, diags := buildRateLimitRuleAPIModel(ctx, plan)

	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if rule.Name != "test-rule" {
		t.Errorf("expected Name='test-rule', got '%s'", rule.Name)
	}
	if rule.Timeframe != 60 {
		t.Errorf("expected Timeframe=60, got %d", rule.Timeframe)
	}
	if rule.Threshold != 100 {
		t.Errorf("expected Threshold=100, got %d", rule.Threshold)
	}
	if rule.TTL != 300 {
		t.Errorf("expected TTL=300, got %d", rule.TTL)
	}
	if rule.Active != true {
		t.Errorf("expected Active=true, got %v", rule.Active)
	}
	if rule.Key == nil {
		t.Fatal("expected non-nil Key")
	}
}

func TestBuildRateLimitRuleAPIModel_NullPairwith(t *testing.T) {
	ctx := context.Background()

	emptyTagsList, _ := types.ListValueFrom(ctx, types.StringType, []string{})
	emptyObj, _ := types.ObjectValue(tagFilterAttrTypes, map[string]attr.Value{
		"relation": types.StringValue("OR"),
		"tags":     emptyTagsList,
	})
	emptySet, _ := types.SetValue(types.ObjectType{AttrTypes: tagFilterAttrTypes}, []attr.Value{emptyObj})

	plan := &RateLimitRuleResourceModel{
		ID:          types.StringValue("rl-2"),
		Name:        types.StringValue("test"),
		Description: types.StringValue(""),
		Global:      types.BoolValue(false),
		Active:      types.BoolValue(true),
		Timeframe:   types.Int64Value(60),
		Threshold:   types.Int64Value(100),
		TTL:         types.Int64Value(0),
		Action:      types.StringValue("action-monitor"),
		IsActionBan: types.BoolValue(false),
		Tags:        types.ListNull(types.StringType),
		Key: []RateLimitKeyModel{
			{Attrs: types.StringValue("ip"), Args: types.StringNull(), Plugins: types.StringNull(), Cookies: types.StringNull(), Headers: types.StringNull()},
		},
		Pairwith: types.StringNull(),
		Include:  emptySet,
		Exclude:  emptySet,
		ConfigID: types.StringValue("cfg1"),
	}

	rule, diags := buildRateLimitRuleAPIModel(ctx, plan)

	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	// When pairwith is null, it should default to {"self": "self"}
	pairwithMap, ok := rule.Pairwith.(map[string]string)
	if !ok {
		t.Fatalf("expected pairwith to be map[string]string, got %T", rule.Pairwith)
	}
	if pairwithMap["self"] != "self" {
		t.Errorf("expected pairwith[self]='self', got '%s'", pairwithMap["self"])
	}
}

func TestBuildRateLimitRuleAPIModel_WithTags(t *testing.T) {
	ctx := context.Background()

	tags, _ := types.ListValueFrom(ctx, types.StringType, []string{"tag1", "tag2"})

	emptyTagsList, _ := types.ListValueFrom(ctx, types.StringType, []string{})
	emptyObj, _ := types.ObjectValue(tagFilterAttrTypes, map[string]attr.Value{
		"relation": types.StringValue("OR"),
		"tags":     emptyTagsList,
	})
	emptySet, _ := types.SetValue(types.ObjectType{AttrTypes: tagFilterAttrTypes}, []attr.Value{emptyObj})

	plan := &RateLimitRuleResourceModel{
		ID:          types.StringValue("rl-3"),
		Name:        types.StringValue("test"),
		Description: types.StringValue(""),
		Global:      types.BoolValue(false),
		Active:      types.BoolValue(true),
		Timeframe:   types.Int64Value(60),
		Threshold:   types.Int64Value(100),
		TTL:         types.Int64Value(0),
		Action:      types.StringValue("action-monitor"),
		IsActionBan: types.BoolValue(false),
		Tags:        tags,
		Key: []RateLimitKeyModel{
			{Attrs: types.StringValue("ip"), Args: types.StringNull(), Plugins: types.StringNull(), Cookies: types.StringNull(), Headers: types.StringNull()},
		},
		Pairwith: types.StringValue(`{"self":"self"}`),
		Include:  emptySet,
		Exclude:  emptySet,
		ConfigID: types.StringValue("cfg1"),
	}

	rule, diags := buildRateLimitRuleAPIModel(ctx, plan)

	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}

	if len(rule.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(rule.Tags))
	}
}

func TestTagFilterAttrTypes(t *testing.T) {
	if _, ok := tagFilterAttrTypes["relation"]; !ok {
		t.Error("expected 'relation' in tagFilterAttrTypes")
	}
	if _, ok := tagFilterAttrTypes["tags"]; !ok {
		t.Error("expected 'tags' in tagFilterAttrTypes")
	}
	if len(tagFilterAttrTypes) != 2 {
		t.Errorf("expected 2 attr types, got %d", len(tagFilterAttrTypes))
	}
}
