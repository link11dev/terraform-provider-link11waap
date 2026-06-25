package providerutil

import (
	"context"

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

// ParseFlowControlKeys converts []client.FlowControlKeyEntry to []FlowControlKeyModel.
func ParseFlowControlKeys(keys []client.FlowControlKeyEntry) []FlowControlKeyModel {
	result := make([]FlowControlKeyModel, 0, len(keys))
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
		result = append(result, km)
	}
	return result
}

// ParseFlowControlSteps converts []client.FlowStepItem to []FlowControlStepModel.
func ParseFlowControlSteps(ctx context.Context, steps []client.FlowStepItem, diags *diag.Diagnostics) []FlowControlStepModel {
	result := make([]FlowControlStepModel, 0, len(steps))
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
		result = append(result, sm)
	}
	return result
}
