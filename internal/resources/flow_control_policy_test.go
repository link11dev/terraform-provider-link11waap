package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/link11/terraform-provider-link11waap/internal/client"
)

func strPtr(s string) *string { return &s }

func TestNewFlowControlPolicyResource(t *testing.T) {
	r := NewFlowControlPolicyResource()
	if r == nil {
		t.Fatal("expected non-nil resource")
	}
	if _, ok := r.(*FlowControlPolicyResource); !ok {
		t.Fatal("expected *FlowControlPolicyResource")
	}
}

func TestFlowControlPolicyResource_Metadata(t *testing.T) {
	r := &FlowControlPolicyResource{}
	req := metadataReq("link11waap")
	resp := metadataResp()
	r.Metadata(context.Background(), req, resp)
	if resp.TypeName != "link11waap_flow_control_policy" {
		t.Errorf("expected 'link11waap_flow_control_policy', got %q", resp.TypeName)
	}
}

func TestFlowControlPolicyResource_Schema(t *testing.T) {
	r := &FlowControlPolicyResource{}
	sResp := schemaResp()
	r.Schema(context.Background(), schemaReq(), sResp)

	expectedAttrs := []string{"config_id", "id", "name", "description", "active", "timeframe", "tags", "include", "exclude"}
	for _, a := range expectedAttrs {
		if _, ok := sResp.Schema.Attributes[a]; !ok {
			t.Errorf("expected attribute %q in schema", a)
		}
	}
	for _, b := range []string{"key", "steps"} {
		if _, ok := sResp.Schema.Blocks[b]; !ok {
			t.Errorf("expected block %q in schema", b)
		}
	}
}

func TestFlowControlPolicyResource_Configure_NilProvider(t *testing.T) {
	r := &FlowControlPolicyResource{}
	r.Configure(context.Background(), configureReq(nil), configureResp())
	if r.client != nil {
		t.Error("expected nil client for nil provider data")
	}
}

func TestFlowControlPolicyResource_ImportState_Valid(t *testing.T) {
	r := &FlowControlPolicyResource{}
	resp := testImportState(t, r, "config123/fc456")
	if resp.Diagnostics.HasError() {
		t.Errorf("expected no errors, got: %v", resp.Diagnostics)
	}
}

func TestFlowControlPolicyResource_ImportState_Invalid(t *testing.T) {
	r := &FlowControlPolicyResource{}
	resp := testImportState(t, r, "invalid")
	if !resp.Diagnostics.HasError() {
		t.Error("expected error for invalid import ID")
	}
}

func TestFlowControlPolicyResource_ImportState_TooManyParts(t *testing.T) {
	r := &FlowControlPolicyResource{}
	resp := testImportState(t, r, "a/b/c")
	if !resp.Diagnostics.HasError() {
		t.Error("expected error for too many parts")
	}
}

func TestBuildFlowControlKeys_AllKeyTypes(t *testing.T) {
	keys := []FlowControlKeyModel{
		{Attrs: types.StringValue("ip"), Args: types.StringNull(), Plugins: types.StringNull(), Cookies: types.StringNull(), Headers: types.StringNull()},
		{Attrs: types.StringNull(), Args: types.StringValue("q"), Plugins: types.StringNull(), Cookies: types.StringNull(), Headers: types.StringNull()},
		{Attrs: types.StringNull(), Args: types.StringNull(), Plugins: types.StringValue("p"), Cookies: types.StringNull(), Headers: types.StringNull()},
		{Attrs: types.StringNull(), Args: types.StringNull(), Plugins: types.StringNull(), Cookies: types.StringValue("sid"), Headers: types.StringNull()},
		{Attrs: types.StringNull(), Args: types.StringNull(), Plugins: types.StringNull(), Cookies: types.StringNull(), Headers: types.StringValue("x-key")},
	}
	result := buildFlowControlKeys(keys)
	if len(result) != 5 {
		t.Fatalf("expected 5 entries, got %d", len(result))
	}
	if result[0].Attrs == nil || *result[0].Attrs != "ip" {
		t.Error("expected attrs='ip'")
	}
	if result[0].Args != nil || result[0].Plugins != nil || result[0].Cookies != nil || result[0].Headers != nil {
		t.Error("expected other fields nil for attrs entry")
	}
	if result[1].Args == nil || *result[1].Args != "q" {
		t.Error("expected args='q'")
	}
	if result[2].Plugins == nil || *result[2].Plugins != "p" {
		t.Error("expected plugins='p'")
	}
	if result[3].Cookies == nil || *result[3].Cookies != "sid" {
		t.Error("expected cookies='sid'")
	}
	if result[4].Headers == nil || *result[4].Headers != "x-key" {
		t.Error("expected headers='x-key'")
	}
}

func TestParseFlowControlKeys_AllKeyTypes(t *testing.T) {
	entries := []client.FlowControlKeyEntry{
		{Attrs: strPtr("session")},
		{Args: strPtr("q")},
		{Plugins: strPtr("p")},
		{Cookies: strPtr("sid")},
		{Headers: strPtr("x-key")},
	}
	result := parseFlowControlKeys(entries)
	if len(result) != 5 {
		t.Fatalf("expected 5 models, got %d", len(result))
	}
	if result[0].Attrs.ValueString() != "session" {
		t.Error("expected attrs='session'")
	}
	if !result[0].Args.IsNull() || !result[0].Plugins.IsNull() {
		t.Error("expected other fields null")
	}
	if result[1].Args.ValueString() != "q" {
		t.Error("expected args='q'")
	}
	if result[2].Plugins.ValueString() != "p" {
		t.Error("expected plugins='p'")
	}
	if result[3].Cookies.ValueString() != "sid" {
		t.Error("expected cookies='sid'")
	}
	if result[4].Headers.ValueString() != "x-key" {
		t.Error("expected headers='x-key'")
	}
}

func TestFlowControlKeys_RoundTrip(t *testing.T) {
	entries := []client.FlowControlKeyEntry{
		{Attrs: strPtr("ip")},
		{Cookies: strPtr("sid")},
	}
	parsed := parseFlowControlKeys(entries)
	built := buildFlowControlKeys(parsed)
	if len(built) != 2 {
		t.Fatalf("expected 2, got %d", len(built))
	}
	if built[0].Attrs == nil || *built[0].Attrs != "ip" {
		t.Error("round-trip attrs mismatch")
	}
	if built[1].Cookies == nil || *built[1].Cookies != "sid" {
		t.Error("round-trip cookies mismatch")
	}
}

func TestBuildFlowControlSteps(t *testing.T) {
	ctx := context.Background()
	headers, _ := types.MapValueFrom(ctx, types.StringType, map[string]string{"X-Test": "1"})
	steps := []FlowControlStepModel{
		{
			Method:  types.StringValue("GET"),
			URI:     types.StringValue("/login"),
			Headers: headers,
			Cookies: types.MapNull(types.StringType),
			Args:    types.MapNull(types.StringType),
			Plugins: types.MapNull(types.StringType),
		},
	}
	var diags diag.Diagnostics
	result := buildFlowControlSteps(ctx, steps, &diags)
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 step, got %d", len(result))
	}
	if result[0].Method != "GET" || result[0].URI != "/login" {
		t.Error("step method/uri mismatch")
	}
	if result[0].Headers["X-Test"] != "1" {
		t.Error("expected headers X-Test=1")
	}
	if result[0].Cookies != nil {
		t.Error("expected nil cookies for null map")
	}
}

func TestParseFlowControlSteps(t *testing.T) {
	ctx := context.Background()
	steps := []client.FlowStepItem{
		{Method: "POST", URI: "/checkout", Args: map[string]string{"step": "2"}},
	}
	var diags diag.Diagnostics
	result := parseFlowControlSteps(ctx, steps, &diags)
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 step, got %d", len(result))
	}
	if result[0].Method.ValueString() != "POST" || result[0].URI.ValueString() != "/checkout" {
		t.Error("step method/uri mismatch")
	}
	if result[0].Args.IsNull() {
		t.Error("expected non-null args")
	}
	if !result[0].Headers.IsNull() {
		t.Error("expected null headers")
	}
}

func TestBuildFlowControlPolicyAPIModel_BasicFields(t *testing.T) {
	ctx := context.Background()
	tags, _ := types.ListValueFrom(ctx, types.StringType, []string{"t1"})
	include, _ := types.ListValueFrom(ctx, types.StringType, []string{"in1"})

	plan := &FlowControlPolicyResourceModel{
		ID:          types.StringValue("fc-1"),
		Name:        types.StringValue("flow"),
		Description: types.StringValue("desc"),
		Active:      types.BoolValue(true),
		Timeframe:   types.Int64Value(120),
		Tags:        tags,
		Include:     include,
		Exclude:     types.ListNull(types.StringType),
		Key: []FlowControlKeyModel{
			{Attrs: types.StringValue("ip"), Args: types.StringNull(), Plugins: types.StringNull(), Cookies: types.StringNull(), Headers: types.StringNull()},
		},
		Steps: []FlowControlStepModel{
			{Method: types.StringValue("GET"), URI: types.StringValue("/a"), Headers: types.MapNull(types.StringType), Cookies: types.MapNull(types.StringType), Args: types.MapNull(types.StringType), Plugins: types.MapNull(types.StringType)},
		},
		ConfigID: types.StringValue("cfg1"),
	}

	policy, diags := buildFlowControlPolicyAPIModel(ctx, plan)
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}
	if policy.Name != "flow" || policy.Timeframe != 120 || !policy.Active {
		t.Error("basic fields mismatch")
	}
	if len(policy.Tags) != 1 {
		t.Errorf("expected 1 tag, got %d", len(policy.Tags))
	}
	if len(policy.Include) != 1 {
		t.Errorf("expected 1 include, got %d", len(policy.Include))
	}
	if policy.Exclude == nil || len(policy.Exclude) != 0 {
		t.Errorf("expected empty (non-nil) exclude, got %v", policy.Exclude)
	}
	if len(policy.Key) != 1 || policy.Key[0].Attrs == nil {
		t.Error("expected key with attrs set")
	}
	if len(policy.Steps) != 1 {
		t.Errorf("expected 1 step, got %d", len(policy.Steps))
	}
}
