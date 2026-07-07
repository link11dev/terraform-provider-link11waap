package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
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

func tagFilterSetValue(objType tftypes.Object, entries ...map[string]tftypes.Value) tftypes.Value {
	elems := make([]tftypes.Value, 0, len(entries))
	for _, e := range entries {
		elems = append(elems, tftypes.NewValue(objType, e))
	}
	return tftypes.NewValue(tftypes.Set{ElementType: objType}, elems)
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
		if _, ok := sResp.Schema.Blocks[b]; !ok {
			t.Errorf("expected block %q in schema", b)
		}
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
		Include:            types.SetNull(types.ObjectType{AttrTypes: tagFilterAttrTypes}),
		Exclude:            types.SetNull(types.ObjectType{AttrTypes: tagFilterAttrTypes}),
	}
	rule, diags := buildDynamicRuleAPIModel(ctx, plan)
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}
	if rule.Target != "ip" || rule.Threshold != 100 || rule.TTL != 300 {
		t.Errorf("unexpected mapping: %+v", rule)
	}
}

func TestDynamicRuleResource_ValidateConfig_RequiresBothIncludeAndExclude(t *testing.T) {
	ctx := context.Background()
	r := &DynamicRuleResource{}
	objType := tagFilterObjType()
	emptySet := tagFilterSetValue(objType)
	oneEntry := map[string]tftypes.Value{
		"relation": tftypes.NewValue(tftypes.String, "OR"),
		"tags":     tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{tftypes.NewValue(tftypes.String, "a")}),
	}

	tests := []struct {
		name      string
		include   tftypes.Value
		exclude   tftypes.Value
		expectErr bool
	}{
		{"only include", tagFilterSetValue(objType, oneEntry), emptySet, true},
		{"only exclude", emptySet, tagFilterSetValue(objType, oneEntry), true},
		{"neither", emptySet, emptySet, true},
		{"both", tagFilterSetValue(objType, oneEntry), tagFilterSetValue(objType, oneEntry), false},
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
