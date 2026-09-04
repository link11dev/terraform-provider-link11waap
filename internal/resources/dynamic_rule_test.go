package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// tagFilterObjType returns the tftypes.Object type matching the include/exclude nested blocks.
func tagFilterObjType() tftypes.Object {
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"relation": tftypes.String,
		"tags":     tftypes.List{ElementType: tftypes.String},
	}}
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
		Include:            types.ObjectNull(tagFilterAttrTypes),
		Exclude:            types.ObjectNull(tagFilterAttrTypes),
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
	missing := tftypes.NewValue(objType, nil)
	present := tftypes.NewValue(objType, map[string]tftypes.Value{
		"relation": tftypes.NewValue(tftypes.String, "OR"),
		"tags":     tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{tftypes.NewValue(tftypes.String, "a")}),
	})

	tests := []struct {
		name      string
		include   tftypes.Value
		exclude   tftypes.Value
		expectErr bool
	}{
		{"only include", present, missing, true},
		{"only exclude", missing, present, true},
		{"neither", missing, missing, true},
		{"both", present, present, false},
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

func TestDynamicRuleResource_UpgradeState_V0ToV1(t *testing.T) {
	ctx := context.Background()
	r := &DynamicRuleResource{}

	upgraders := r.UpgradeState(ctx)
	upgrader, ok := upgraders[0]
	if !ok || upgrader.PriorSchema == nil {
		t.Fatal("expected a version 0 state upgrader with a prior schema")
	}

	objType := tagFilterObjType()
	listType := tftypes.List{ElementType: tftypes.String}

	priorTFType := upgrader.PriorSchema.Type().TerraformType(ctx)
	priorObjType, ok := priorTFType.(tftypes.Object)
	if !ok {
		t.Fatalf("expected prior schema to produce a tftypes.Object, got %T", priorTFType)
	}

	values := make(map[string]tftypes.Value, len(priorObjType.AttributeTypes))
	for name, at := range priorObjType.AttributeTypes {
		values[name] = tftypes.NewValue(at, nil)
	}
	values["config_id"] = tftypes.NewValue(tftypes.String, "cfg1")
	values["id"] = tftypes.NewValue(tftypes.String, "dr1")
	values["name"] = tftypes.NewValue(tftypes.String, "Burst Protection")
	values["description"] = tftypes.NewValue(tftypes.String, "")
	values["threshold"] = tftypes.NewValue(tftypes.Number, 100)
	values["timeframe"] = tftypes.NewValue(tftypes.Number, 60)
	values["ttl"] = tftypes.NewValue(tftypes.Number, 3600)
	values["active"] = tftypes.NewValue(tftypes.Bool, true)
	values["offload_ip_filtering"] = tftypes.NewValue(tftypes.Bool, false)
	values["target"] = tftypes.NewValue(tftypes.String, "ip")
	values["action"] = tftypes.NewValue(tftypes.String, "action-monitor")
	values["include"] = tftypes.NewValue(tftypes.Set{ElementType: objType}, []tftypes.Value{
		tftypes.NewValue(objType, map[string]tftypes.Value{
			"relation": tftypes.NewValue(tftypes.String, "OR"),
			"tags":     tftypes.NewValue(listType, []tftypes.Value{tftypes.NewValue(tftypes.String, "facebook")}),
		}),
	})
	values["exclude"] = tftypes.NewValue(tftypes.Set{ElementType: objType}, []tftypes.Value{})

	priorState := tfsdk.State{
		Schema: *upgrader.PriorSchema,
		Raw:    tftypes.NewValue(priorTFType, values),
	}

	sReq := schemaReq()
	sResp := schemaResp()
	r.Schema(ctx, sReq, sResp)

	req := resource.UpgradeStateRequest{State: &priorState}
	resp := &resource.UpgradeStateResponse{
		State: tfsdk.State{Schema: sResp.Schema},
	}

	upgrader.StateUpgrader(ctx, req, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diags: %v", resp.Diagnostics)
	}

	var upgraded DynamicRuleResourceModel
	if diags := resp.State.Get(ctx, &upgraded); diags.HasError() {
		t.Fatalf("unexpected diags reading upgraded state: %v", diags)
	}

	if upgraded.Include.IsNull() {
		t.Fatal("expected non-null include object after upgrade")
	}
	var includeModel RateLimitTagFilterModel
	if diags := upgraded.Include.As(ctx, &includeModel, basetypes.ObjectAsOptions{}); diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}
	if includeModel.Relation.ValueString() != "OR" {
		t.Errorf("expected relation='OR', got %q", includeModel.Relation.ValueString())
	}
	var includeTags []string
	includeModel.Tags.ElementsAs(ctx, &includeTags, false)
	if len(includeTags) != 1 || includeTags[0] != "facebook" {
		t.Errorf("expected tags=['facebook'], got %v", includeTags)
	}

	if !upgraded.Exclude.IsNull() {
		t.Error("expected exclude to be null after upgrading an empty legacy set")
	}
}
