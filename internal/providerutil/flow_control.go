package providerutil

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/link11/terraform-provider-link11waap/internal/client"
)

// FlowControlKeyModel describes the data model for a flow control key entry.
type FlowControlKeyModel struct {
	Attrs   types.String `tfsdk:"attrs"`
	Args    types.String `tfsdk:"args"`
	Plugins types.String `tfsdk:"plugins"`
	Cookies types.String `tfsdk:"cookies"`
	Headers types.String `tfsdk:"headers"`
}

// FlowControlStepModel describes the data model for a single flow step.
type FlowControlStepModel struct {
	Method  types.String `tfsdk:"method"`
	URI     types.String `tfsdk:"uri"`
	Headers types.Map    `tfsdk:"headers"`
	Cookies types.Map    `tfsdk:"cookies"`
	Args    types.Map    `tfsdk:"args"`
	Plugins types.Map    `tfsdk:"plugins"`
}

// FlowControlKeyModelType is the object type matching FlowControlKeyModel, used to
// convert between types.List and []FlowControlKeyModel. Exported because the key
// and steps blocks are represented as types.List (not native Go slices) on
// FlowControlPolicyResourceModel: the framework's reflection-based decoding cannot
// represent an unknown value in a plain slice, and Terraform produces unknown
// collections for blocks generated via `dynamic`.
func FlowControlKeyModelType() types.ObjectType {
	return types.ObjectType{AttrTypes: map[string]attr.Type{
		"attrs":   types.StringType,
		"args":    types.StringType,
		"plugins": types.StringType,
		"cookies": types.StringType,
		"headers": types.StringType,
	}}
}

// FlowControlStepModelType is the object type matching FlowControlStepModel, used
// to convert between types.List and []FlowControlStepModel.
func FlowControlStepModelType() types.ObjectType {
	return types.ObjectType{AttrTypes: map[string]attr.Type{
		"method":  types.StringType,
		"uri":     types.StringType,
		"headers": types.MapType{ElemType: types.StringType},
		"cookies": types.MapType{ElemType: types.StringType},
		"args":    types.MapType{ElemType: types.StringType},
		"plugins": types.MapType{ElemType: types.StringType},
	}}
}

// ParseFlowControlKeys converts []client.FlowControlKeyEntry to a non-null types.List.
func ParseFlowControlKeys(ctx context.Context, keys []client.FlowControlKeyEntry) (types.List, diag.Diagnostics) {
	models := make([]FlowControlKeyModel, 0, len(keys))
	for _, k := range keys {
		km := FlowControlKeyModel{
			Attrs:   types.StringNull(),
			Args:    types.StringNull(),
			Plugins: types.StringNull(),
			Cookies: types.StringNull(),
			Headers: types.StringNull(),
		}
		switch {
		case k.Attrs != nil:
			km.Attrs = types.StringValue(*k.Attrs)
		case k.Args != nil:
			km.Args = types.StringValue(*k.Args)
		case k.Plugins != nil:
			km.Plugins = types.StringValue(*k.Plugins)
		case k.Cookies != nil:
			km.Cookies = types.StringValue(*k.Cookies)
		case k.Headers != nil:
			km.Headers = types.StringValue(*k.Headers)
		}
		models = append(models, km)
	}
	return types.ListValueFrom(ctx, FlowControlKeyModelType(), models)
}

// ParseFlowControlSteps converts []client.FlowStepItem to a non-null types.List.
func ParseFlowControlSteps(ctx context.Context, steps []client.FlowStepItem) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	models := make([]FlowControlStepModel, 0, len(steps))
	for _, s := range steps {
		sm := FlowControlStepModel{
			Method:  types.StringValue(s.Method),
			URI:     types.StringValue(s.URI),
			Headers: types.MapNull(types.StringType),
			Cookies: types.MapNull(types.StringType),
			Args:    types.MapNull(types.StringType),
			Plugins: types.MapNull(types.StringType),
		}
		if len(s.Headers) > 0 {
			m, d := types.MapValueFrom(ctx, types.StringType, s.Headers)
			diags.Append(d...)
			sm.Headers = m
		}
		if len(s.Cookies) > 0 {
			m, d := types.MapValueFrom(ctx, types.StringType, s.Cookies)
			diags.Append(d...)
			sm.Cookies = m
		}
		if len(s.Args) > 0 {
			m, d := types.MapValueFrom(ctx, types.StringType, s.Args)
			diags.Append(d...)
			sm.Args = m
		}
		if len(s.Plugins) > 0 {
			m, d := types.MapValueFrom(ctx, types.StringType, s.Plugins)
			diags.Append(d...)
			sm.Plugins = m
		}
		models = append(models, sm)
	}
	list, d := types.ListValueFrom(ctx, FlowControlStepModelType(), models)
	diags.Append(d...)
	return list, diags
}
