package resources

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/link11/terraform-provider-link11waap/internal/client"
	"github.com/link11/terraform-provider-link11waap/internal/providerutil"
	"github.com/stretchr/testify/assert"
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
	keys := []providerutil.FlowControlKeyModel{
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

// mustFlowControlKeyList builds a types.List of FlowControlKeyModel for use in
// FlowControlPolicyResourceModel literals.
func mustFlowControlKeyList(t *testing.T, keys []providerutil.FlowControlKeyModel) types.List {
	t.Helper()
	l, diags := types.ListValueFrom(context.Background(), providerutil.FlowControlKeyModelType(), keys)
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}
	return l
}

// mustFlowControlStepList builds a types.List of FlowControlStepModel for use in
// FlowControlPolicyResourceModel literals.
func mustFlowControlStepList(t *testing.T, steps []providerutil.FlowControlStepModel) types.List {
	t.Helper()
	l, diags := types.ListValueFrom(context.Background(), providerutil.FlowControlStepModelType(), steps)
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}
	return l
}

// flowControlKeysOf decodes a types.List of providerutil.FlowControlKeyModel for assertions.
func flowControlKeysOf(t *testing.T, l types.List) []providerutil.FlowControlKeyModel {
	t.Helper()
	if l.IsNull() || l.IsUnknown() {
		return nil
	}
	var models []providerutil.FlowControlKeyModel
	diags := l.ElementsAs(context.Background(), &models, false)
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}
	return models
}

// flowControlStepsOf decodes a types.List of providerutil.FlowControlStepModel for assertions.
func flowControlStepsOf(t *testing.T, l types.List) []providerutil.FlowControlStepModel {
	t.Helper()
	if l.IsNull() || l.IsUnknown() {
		return nil
	}
	var models []providerutil.FlowControlStepModel
	diags := l.ElementsAs(context.Background(), &models, false)
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}
	return models
}

func TestParseFlowControlKeys_AllKeyTypes(t *testing.T) {
	entries := []client.FlowControlKeyEntry{
		{Attrs: strPtr("session")},
		{Args: strPtr("q")},
		{Plugins: strPtr("p")},
		{Cookies: strPtr("sid")},
		{Headers: strPtr("x-key")},
	}
	list, diags := providerutil.ParseFlowControlKeys(context.Background(), entries)
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}
	result := flowControlKeysOf(t, list)
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
	list, diags := providerutil.ParseFlowControlKeys(context.Background(), entries)
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}
	parsed := flowControlKeysOf(t, list)
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
	steps := []providerutil.FlowControlStepModel{
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
	if result[0].Cookies == nil {
		t.Error("expected empty map cookies for null map")
	}
	if len(result[0].Cookies) != 0 {
		t.Error("expected empty cookies map")
	}
}

func TestParseFlowControlSteps(t *testing.T) {
	ctx := context.Background()
	steps := []client.FlowStepItem{
		{Method: "POST", URI: "/checkout", Args: map[string]string{"step": "2"}},
	}
	list, diags := providerutil.ParseFlowControlSteps(ctx, steps)
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}
	result := flowControlStepsOf(t, list)
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

// fcKeyObjType returns the tftypes.Object type matching the key nested block.
func fcKeyObjType() tftypes.Object {
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"attrs": tftypes.String, "args": tftypes.String, "plugins": tftypes.String,
		"cookies": tftypes.String, "headers": tftypes.String,
	}}
}

// fcStepObjType returns the tftypes.Object type matching the steps nested block.
func fcStepObjType() tftypes.Object {
	return tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"method":  tftypes.String,
		"uri":     tftypes.String,
		"headers": tftypes.Map{ElementType: tftypes.String},
		"cookies": tftypes.Map{ElementType: tftypes.String},
		"args":    tftypes.Map{ElementType: tftypes.String},
		"plugins": tftypes.Map{ElementType: tftypes.String},
	}}
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
		Key: mustFlowControlKeyList(t, []providerutil.FlowControlKeyModel{
			{Attrs: types.StringValue("ip"), Args: types.StringNull(), Plugins: types.StringNull(), Cookies: types.StringNull(), Headers: types.StringNull()},
		}),
		Steps: mustFlowControlStepList(t, []providerutil.FlowControlStepModel{
			{Method: types.StringValue("GET"), URI: types.StringValue("/a"), Headers: types.MapNull(types.StringType), Cookies: types.MapNull(types.StringType), Args: types.MapNull(types.StringType), Plugins: types.MapNull(types.StringType)},
		}),
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

func TestBuildFlowControlPolicyAPIModel_NonNullExclude(t *testing.T) {
	ctx := context.Background()
	exclude, _ := types.ListValueFrom(ctx, types.StringType, []string{"excl1"})

	plan := &FlowControlPolicyResourceModel{
		ID:          types.StringValue("fc-1"),
		Name:        types.StringValue("flow"),
		Description: types.StringValue(""),
		Active:      types.BoolValue(false),
		Timeframe:   types.Int64Value(30),
		Tags:        types.ListNull(types.StringType),
		Include:     types.ListNull(types.StringType),
		Exclude:     exclude,
		Key:         mustFlowControlKeyList(t, []providerutil.FlowControlKeyModel{}),
		Steps:       mustFlowControlStepList(t, []providerutil.FlowControlStepModel{}),
		ConfigID:    types.StringValue("cfg1"),
	}

	policy, diags := buildFlowControlPolicyAPIModel(ctx, plan)
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}
	if len(policy.Exclude) != 1 || policy.Exclude[0] != "excl1" {
		t.Errorf("expected exclude=['excl1'], got %v", policy.Exclude)
	}
}

func TestBuildFlowControlSteps_AllMaps(t *testing.T) {
	ctx := context.Background()
	cookies, _ := types.MapValueFrom(ctx, types.StringType, map[string]string{"sid": "abc"})
	args, _ := types.MapValueFrom(ctx, types.StringType, map[string]string{"step": "2"})
	plugins, _ := types.MapValueFrom(ctx, types.StringType, map[string]string{"key": "val"})

	steps := []providerutil.FlowControlStepModel{
		{
			Method:  types.StringValue("POST"),
			URI:     types.StringValue("/submit"),
			Headers: types.MapNull(types.StringType),
			Cookies: cookies,
			Args:    args,
			Plugins: plugins,
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
	if result[0].Cookies["sid"] != "abc" {
		t.Error("expected cookies sid=abc")
	}
	if result[0].Args["step"] != "2" {
		t.Error("expected args step=2")
	}
	if result[0].Plugins["key"] != "val" {
		t.Error("expected plugins key=val")
	}
}

func TestParseFlowControlSteps_AllMaps(t *testing.T) {
	ctx := context.Background()
	steps := []client.FlowStepItem{
		{
			Method:  "GET",
			URI:     "/page",
			Headers: map[string]string{"X-Custom": "val"},
			Cookies: map[string]string{"sid": "123"},
			Args:    map[string]string{},
			Plugins: map[string]string{"plugin-key": "plugin-val"},
		},
	}
	list, diags := providerutil.ParseFlowControlSteps(ctx, steps)
	if diags.HasError() {
		t.Fatalf("unexpected diags: %v", diags)
	}
	result := flowControlStepsOf(t, list)
	if len(result) != 1 {
		t.Fatalf("expected 1 step, got %d", len(result))
	}
	if result[0].Headers.IsNull() {
		t.Error("expected non-null headers")
	}
	if result[0].Cookies.IsNull() {
		t.Error("expected non-null cookies")
	}
	if result[0].Plugins.IsNull() {
		t.Error("expected non-null plugins")
	}
	if !result[0].Args.IsNull() {
		t.Error("expected null args for empty map")
	}
}

// ---- CRUD with failing client ----

func TestFlowControlPolicyResource_CRUD_WithFailingClient(t *testing.T) {
	r := &FlowControlPolicyResource{}
	planVals := map[string]tftypes.Value{
		"config_id":   tftypes.NewValue(tftypes.String, "cfg1"),
		"id":          tftypes.NewValue(tftypes.String, nil),
		"name":        tftypes.NewValue(tftypes.String, "test-flow"),
		"description": tftypes.NewValue(tftypes.String, ""),
		"active":      tftypes.NewValue(tftypes.Bool, true),
		"timeframe":   tftypes.NewValue(tftypes.Number, 60),
	}
	stateVals := map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"id":        tftypes.NewValue(tftypes.String, "fc1"),
	}

	t.Run("Create", func(t *testing.T) { crudCreateWithClient(t, r, planVals) })
	t.Run("Read", func(t *testing.T) { crudReadWithClient(t, r, stateVals) })
	t.Run("Update", func(t *testing.T) { crudUpdateWithClient(t, r, planVals) })
	t.Run("Delete", func(t *testing.T) { crudDeleteWithClient(t, r, stateVals) })
}

// ---- Read with mock ----

func TestFlowControlPolicyResource_Read_WithMock(t *testing.T) {
	r := &FlowControlPolicyResource{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.FlowControl{
			ID:          "fc1",
			Name:        "test-flow",
			Description: "a flow",
			Active:      true,
			Timeframe:   60,
			Tags:        []string{"tag1"},
			Include:     []string{"inc1"},
			Exclude:     []string{"exc1"},
			Key:         []client.FlowControlKeyEntry{{Attrs: strPtr("ip")}},
			Steps: []client.FlowStepItem{
				{Method: "GET", URI: "/test", Headers: map[string]string{}, Cookies: map[string]string{}, Args: map[string]string{}},
			},
		})
	})
	configureResourceWithMock(t, r, handler)

	resp := readWithMock(t, r, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"id":        tftypes.NewValue(tftypes.String, "fc1"),
	})

	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}

func TestFlowControlPolicyResource_Read_NotFound(t *testing.T) {
	r := &FlowControlPolicyResource{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"not found"}`))
	})
	configureResourceWithMock(t, r, handler)

	resp := readWithMock(t, r, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"id":        tftypes.NewValue(tftypes.String, "missing"),
	})

	assert.False(t, resp.Diagnostics.HasError())
}

// ---- ValidateConfig ----

func TestFlowControlPolicyResource_ValidateConfig_Valid(t *testing.T) {
	ctx := context.Background()
	r := &FlowControlPolicyResource{}
	keyObjType := fcKeyObjType()
	stepObjType := fcStepObjType()

	config := buildConfig(ctx, t, r, map[string]tftypes.Value{
		"key": tftypes.NewValue(
			tftypes.List{ElementType: keyObjType},
			[]tftypes.Value{
				tftypes.NewValue(keyObjType, map[string]tftypes.Value{
					"attrs": tftypes.NewValue(tftypes.String, "ip"), "args": tftypes.NewValue(tftypes.String, nil),
					"plugins": tftypes.NewValue(tftypes.String, nil), "cookies": tftypes.NewValue(tftypes.String, nil),
					"headers": tftypes.NewValue(tftypes.String, nil),
				}),
			},
		),
		"steps": tftypes.NewValue(
			tftypes.List{ElementType: stepObjType},
			[]tftypes.Value{
				tftypes.NewValue(stepObjType, map[string]tftypes.Value{
					"method": tftypes.NewValue(tftypes.String, "GET"), "uri": tftypes.NewValue(tftypes.String, "/test"),
					"headers": tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
					"cookies": tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
					"args":    tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
					"plugins": tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
				}),
			},
		),
	})
	req := resource.ValidateConfigRequest{Config: config}
	resp := &resource.ValidateConfigResponse{}
	r.ValidateConfig(ctx, req, resp)
	assert.False(t, resp.Diagnostics.HasError(), "valid config should not produce errors: %v", resp.Diagnostics)
}

func TestFlowControlPolicyResource_ValidateConfig_NoKeys(t *testing.T) {
	ctx := context.Background()
	r := &FlowControlPolicyResource{}
	keyObjType := fcKeyObjType()
	stepObjType := fcStepObjType()

	config := buildConfig(ctx, t, r, map[string]tftypes.Value{
		"key": tftypes.NewValue(tftypes.List{ElementType: keyObjType}, []tftypes.Value{}),
		"steps": tftypes.NewValue(
			tftypes.List{ElementType: stepObjType},
			[]tftypes.Value{
				tftypes.NewValue(stepObjType, map[string]tftypes.Value{
					"method": tftypes.NewValue(tftypes.String, "GET"), "uri": tftypes.NewValue(tftypes.String, "/test"),
					"headers": tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
					"cookies": tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
					"args":    tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
					"plugins": tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
				}),
			},
		),
	})
	req := resource.ValidateConfigRequest{Config: config}
	resp := &resource.ValidateConfigResponse{}
	r.ValidateConfig(ctx, req, resp)
	assert.True(t, resp.Diagnostics.HasError(), "empty key list should produce error")
}

func TestFlowControlPolicyResource_ValidateConfig_KeyMultipleFields(t *testing.T) {
	ctx := context.Background()
	r := &FlowControlPolicyResource{}
	keyObjType := fcKeyObjType()
	stepObjType := fcStepObjType()

	config := buildConfig(ctx, t, r, map[string]tftypes.Value{
		"key": tftypes.NewValue(
			tftypes.List{ElementType: keyObjType},
			[]tftypes.Value{
				tftypes.NewValue(keyObjType, map[string]tftypes.Value{
					"attrs": tftypes.NewValue(tftypes.String, "ip"), "args": tftypes.NewValue(tftypes.String, "q"),
					"plugins": tftypes.NewValue(tftypes.String, nil), "cookies": tftypes.NewValue(tftypes.String, nil),
					"headers": tftypes.NewValue(tftypes.String, nil),
				}),
			},
		),
		"steps": tftypes.NewValue(
			tftypes.List{ElementType: stepObjType},
			[]tftypes.Value{
				tftypes.NewValue(stepObjType, map[string]tftypes.Value{
					"method": tftypes.NewValue(tftypes.String, "GET"), "uri": tftypes.NewValue(tftypes.String, "/test"),
					"headers": tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
					"cookies": tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
					"args":    tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
					"plugins": tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
				}),
			},
		),
	})
	req := resource.ValidateConfigRequest{Config: config}
	resp := &resource.ValidateConfigResponse{}
	r.ValidateConfig(ctx, req, resp)
	assert.True(t, resp.Diagnostics.HasError(), "key with multiple fields should produce error")
}

func TestFlowControlPolicyResource_ValidateConfig_KeyNoFields(t *testing.T) {
	ctx := context.Background()
	r := &FlowControlPolicyResource{}
	keyObjType := fcKeyObjType()
	stepObjType := fcStepObjType()

	config := buildConfig(ctx, t, r, map[string]tftypes.Value{
		"key": tftypes.NewValue(
			tftypes.List{ElementType: keyObjType},
			[]tftypes.Value{
				tftypes.NewValue(keyObjType, map[string]tftypes.Value{
					"attrs": tftypes.NewValue(tftypes.String, nil), "args": tftypes.NewValue(tftypes.String, nil),
					"plugins": tftypes.NewValue(tftypes.String, nil), "cookies": tftypes.NewValue(tftypes.String, nil),
					"headers": tftypes.NewValue(tftypes.String, nil),
				}),
			},
		),
		"steps": tftypes.NewValue(
			tftypes.List{ElementType: stepObjType},
			[]tftypes.Value{
				tftypes.NewValue(stepObjType, map[string]tftypes.Value{
					"method": tftypes.NewValue(tftypes.String, "GET"), "uri": tftypes.NewValue(tftypes.String, "/test"),
					"headers": tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
					"cookies": tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
					"args":    tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
					"plugins": tftypes.NewValue(tftypes.Map{ElementType: tftypes.String}, nil),
				}),
			},
		),
	})
	req := resource.ValidateConfigRequest{Config: config}
	resp := &resource.ValidateConfigResponse{}
	r.ValidateConfig(ctx, req, resp)
	assert.True(t, resp.Diagnostics.HasError(), "key with no fields should produce error")
}

func TestFlowControlPolicyResource_ValidateConfig_NoSteps(t *testing.T) {
	ctx := context.Background()
	r := &FlowControlPolicyResource{}
	keyObjType := fcKeyObjType()
	stepObjType := fcStepObjType()

	config := buildConfig(ctx, t, r, map[string]tftypes.Value{
		"key": tftypes.NewValue(
			tftypes.List{ElementType: keyObjType},
			[]tftypes.Value{
				tftypes.NewValue(keyObjType, map[string]tftypes.Value{
					"attrs": tftypes.NewValue(tftypes.String, "ip"), "args": tftypes.NewValue(tftypes.String, nil),
					"plugins": tftypes.NewValue(tftypes.String, nil), "cookies": tftypes.NewValue(tftypes.String, nil),
					"headers": tftypes.NewValue(tftypes.String, nil),
				}),
			},
		),
		"steps": tftypes.NewValue(tftypes.List{ElementType: stepObjType}, []tftypes.Value{}),
	})
	req := resource.ValidateConfigRequest{Config: config}
	resp := &resource.ValidateConfigResponse{}
	r.ValidateConfig(ctx, req, resp)
	assert.True(t, resp.Diagnostics.HasError(), "empty steps list should produce error")
}
