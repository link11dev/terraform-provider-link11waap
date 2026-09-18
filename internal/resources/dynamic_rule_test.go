package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// tagFilterObjType returns the tftypes.Object type matching the include/exclude nested blocks.
func tagFilterObjType() tftypes.Object {
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"relation": tftypes.String,
		"tags":     tftypes.List{ElementType: tftypes.String},
	}}
}

// tagFilterSetValue builds a schema version 0 include/exclude value (a set of tag
// filter objects). Still used by the state upgrade tests.
func tagFilterSetValue(objType tftypes.Object, entries ...map[string]tftypes.Value) tftypes.Value {
	elems := make([]tftypes.Value, 0, len(entries))
	for _, e := range entries {
		elems = append(elems, tftypes.NewValue(objType, e))
	}
	return tftypes.NewValue(tftypes.Set{ElementType: objType}, elems)
}

// tagFilterEntry returns the raw attribute values of a tag filter block.
func tagFilterEntry(relation string, tags ...string) map[string]tftypes.Value {
	tagValues := make([]tftypes.Value, 0, len(tags))
	for _, tag := range tags {
		tagValues = append(tagValues, tftypes.NewValue(tftypes.String, tag))
	}
	return map[string]tftypes.Value{
		"relation": tftypes.NewValue(tftypes.String, relation),
		"tags":     tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, tagValues),
	}
}

// tagFilterObject builds the framework value of an include/exclude block.
func tagFilterObject(t *testing.T, relation string, tags ...string) types.Object {
	t.Helper()
	if tags == nil {
		tags = []string{}
	}
	tagsList, diags := types.ListValueFrom(context.Background(), types.StringType, tags)
	if diags.HasError() {
		t.Fatalf("unexpected diags building tags: %v", diags)
	}
	obj, diags := types.ObjectValue(tagFilterAttrTypes, map[string]attr.Value{
		"relation": types.StringValue(relation),
		"tags":     tagsList,
	})
	if diags.HasError() {
		t.Fatalf("unexpected diags building tag filter: %v", diags)
	}
	return obj
}

func TestNewDynamicRuleResource(t *testing.T) {
	r := NewDynamicRuleResource()
	if r == nil {
		t.Fatal("expected non-nil resource")
	}
	if _, ok := r.(*DynamicRuleResource); !ok {
		t.Fatal("expected *DynamicRuleResource")
	}
}

func TestDynamicRuleResource_Metadata(t *testing.T) {
	r := &DynamicRuleResource{}
	req := metadataReq("link11waap")
	resp := metadataResp()
	r.Metadata(context.Background(), req, resp)
	if resp.TypeName != "link11waap_dynamic_rule" {
		t.Errorf("expected 'link11waap_dynamic_rule', got %q", resp.TypeName)
	}
}

func TestDynamicRuleResource_Schema(t *testing.T) {
	r := &DynamicRuleResource{}
	sReq := schemaReq()
	sResp := schemaResp()
	r.Schema(context.Background(), sReq, sResp)

	expectedAttrs := []string{
		"config_id", "id", "name", "description", "threshold", "timeframe",
		"ttl", "active", "offload_ip_filtering", "target", "action", "tags",
	}
	for _, a := range expectedAttrs {
		if _, ok := sResp.Schema.Attributes[a]; !ok {
			t.Errorf("expected attribute %q in schema", a)
		}
	}
	for _, b := range []string{"include", "exclude"} {
		block, ok := sResp.Schema.Blocks[b]
		if !ok {
			t.Errorf("expected block %q in schema", b)
			continue
		}
		// WP-2552: the blocks must be single nested blocks, otherwise Terraform
		// identifies them by a hash of their whole value and a single tag change
		// re-renders the entire block.
		if _, ok := block.(schema.SingleNestedBlock); !ok {
			t.Errorf("expected block %q to be a SingleNestedBlock, got %T", b, block)
		}
	}

	if sResp.Schema.Version != dynamicRuleSchemaVersion {
		t.Errorf("expected schema version %d, got %d", dynamicRuleSchemaVersion, sResp.Schema.Version)
	}
}

func TestDynamicRuleResource_Configure_NilProvider(t *testing.T) {
	r := &DynamicRuleResource{}
	req := configureReq(nil)
	resp := configureResp()
	r.Configure(context.Background(), req, resp)
	if r.client != nil {
		t.Error("expected nil client for nil provider data")
	}
}

func TestDynamicRuleResource_ImportState_Valid(t *testing.T) {
	r := &DynamicRuleResource{}
	resp := testImportState(t, r, "config123/dr456")
	if resp.Diagnostics.HasError() {
		t.Errorf("expected no errors, got: %v", resp.Diagnostics)
	}
}

func TestDynamicRuleResource_ImportState_Invalid(t *testing.T) {
	r := &DynamicRuleResource{}
	resp := testImportState(t, r, "invalid")
	if !resp.Diagnostics.HasError() {
		t.Error("expected error for invalid import ID")
	}
}

func TestDynamicRuleResource_ImportState_TooManyParts(t *testing.T) {
	r := &DynamicRuleResource{}
	resp := testImportState(t, r, "a/b/c")
	if !resp.Diagnostics.HasError() {
		t.Error("expected error for too many parts")
	}
}

func TestBuildDynamicRuleAPIModel_BasicFields(t *testing.T) {
	ctx := context.Background()
	plan := &DynamicRuleResourceModel{
		ID:                 types.StringValue("dr1"),
		Name:               types.StringValue("dyn"),
		Description:        types.StringValue("desc"),
		Threshold:          types.Int64Value(100),
		Timeframe:          types.Int64Value(60),
		TTL:                types.Int64Value(300),
		Active:             types.BoolValue(true),
		OffloadIPFiltering: types.BoolValue(false),
		Target:             types.StringValue("ip"),
		Action:             types.StringValue("action-monitor"),
		Tags:               types.ListNull(types.StringType),
		Include:            tagFilterObject(t, "OR", "facebook"),
		Exclude:            tagFilterObject(t, "AND"),
	}
	rule, diags := buildDynamicRuleAPIModel(ctx, plan)
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}
	if rule.Target != "ip" || rule.Threshold != 100 || rule.TTL != 300 {
		t.Errorf("unexpected mapping: %+v", rule)
	}
	if rule.Include.Relation != "OR" || len(rule.Include.Tags) != 1 || rule.Include.Tags[0] != "facebook" {
		t.Errorf("unexpected include mapping: %+v", rule.Include)
	}
	if rule.Exclude.Relation != "AND" || len(rule.Exclude.Tags) != 0 {
		t.Errorf("unexpected exclude mapping: %+v", rule.Exclude)
	}
}

// A null include/exclude cannot be rejected by ValidateConfig when it only becomes
// known at plan time (dynamic blocks), so buildDynamicRuleAPIModel enforces the
// "exactly one block" rule as well.
func TestBuildDynamicRuleAPIModel_RejectsMissingBlocks(t *testing.T) {
	ctx := context.Background()
	nullFilter := types.ObjectNull(tagFilterAttrTypes)

	tests := []struct {
		name    string
		include types.Object
		exclude types.Object
	}{
		{"include missing", nullFilter, tagFilterObject(t, "OR")},
		{"exclude missing", tagFilterObject(t, "OR"), nullFilter},
		{"both missing", nullFilter, nullFilter},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			plan := &DynamicRuleResourceModel{
				ID:      types.StringValue("dr1"),
				Name:    types.StringValue("dyn"),
				Target:  types.StringValue("ip"),
				Tags:    types.ListNull(types.StringType),
				Include: tc.include,
				Exclude: tc.exclude,
			}
			if _, diags := buildDynamicRuleAPIModel(ctx, plan); !diags.HasError() {
				t.Error("expected an error for a missing include/exclude block")
			}
		})
	}
}

func TestDynamicRuleResource_ValidateConfig_RequiresBothIncludeAndExclude(t *testing.T) {
	ctx := context.Background()
	r := &DynamicRuleResource{}
	objType := tagFilterObjType()
	nullBlock := tftypes.NewValue(objType, nil)
	unknownBlock := tftypes.NewValue(objType, tftypes.UnknownValue)
	oneEntry := tftypes.NewValue(objType, tagFilterEntry("OR", "a"))
	missingTags := tftypes.NewValue(objType, map[string]tftypes.Value{
		"relation": tftypes.NewValue(tftypes.String, "OR"),
		"tags":     tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, nil),
	})
	missingRelation := tftypes.NewValue(objType, map[string]tftypes.Value{
		"relation": tftypes.NewValue(tftypes.String, nil),
		"tags":     tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{}),
	})

	tests := []struct {
		name      string
		include   tftypes.Value
		exclude   tftypes.Value
		expectErr bool
	}{
		{"only include", oneEntry, nullBlock, true},
		{"only exclude", nullBlock, oneEntry, true},
		{"neither", nullBlock, nullBlock, true},
		{"both", oneEntry, oneEntry, false},
		{"tags omitted", oneEntry, missingTags, true},
		{"relation omitted", missingRelation, oneEntry, true},
		// Values produced by a dynamic block are only known at apply time; they
		// are re-checked in buildDynamicRuleAPIModel.
		{"unknown blocks", unknownBlock, unknownBlock, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			config := buildConfig(ctx, t, r, map[string]tftypes.Value{
				"include": tc.include,
				"exclude": tc.exclude,
			})
			req := resource.ValidateConfigRequest{Config: config}
			resp := &resource.ValidateConfigResponse{}
			r.ValidateConfig(ctx, req, resp)
			if resp.Diagnostics.HasError() != tc.expectErr {
				t.Errorf("expected HasError=%v, got diags: %v", tc.expectErr, resp.Diagnostics)
			}
		})
	}
}
