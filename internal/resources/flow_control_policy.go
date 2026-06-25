package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/link11/terraform-provider-link11waap/internal/client"
	"github.com/link11/terraform-provider-link11waap/internal/providerutil"
)

var (
	_ resource.Resource                   = &FlowControlPolicyResource{}
	_ resource.ResourceWithImportState    = &FlowControlPolicyResource{}
	_ resource.ResourceWithValidateConfig = &FlowControlPolicyResource{}
)

// flowControlAttrsEnum lists the allowed values for the key.attrs field.
var flowControlAttrsEnum = []string{
	"asnFlowSef", "authority", "company", "country", "ip", "method", "network",
	"path", "query", "region", "secpolentryid", "securitypolicyentryid",
	"securitypolicyentry", "secpolid", "securitypolicyid", "securitypolicy",
	"secpolname", "securitypolicyname", "secpolentryname", "securitypolicyentryname",
	"session", "subregion", "tags", "uri",
}

// flowControlMethodEnum lists the allowed HTTP methods for a flow step.
var flowControlMethodEnum = []string{
	"GET", "HEAD", "POST", "PUT", "DELETE", "CONNECT", "TRACE", "OPTIONS", "PATCH",
}

// FlowControlPolicyResource implements the resource for managing a flow control policy
type FlowControlPolicyResource struct {
	client *client.Client
}

// FlowControlPolicyResourceModel describes the resource model for a flow control policy
type FlowControlPolicyResourceModel struct {
	ConfigID    types.String                        `tfsdk:"config_id"`
	ID          types.String                        `tfsdk:"id"`
	Name        types.String                        `tfsdk:"name"`
	Description types.String                        `tfsdk:"description"`
	Active      types.Bool                          `tfsdk:"active"`
	Timeframe   types.Int64                         `tfsdk:"timeframe"`
	Tags        types.List                          `tfsdk:"tags"`
	Include     types.List                          `tfsdk:"include"`
	Exclude     types.List                          `tfsdk:"exclude"`
	Key         []providerutil.FlowControlKeyModel  `tfsdk:"key"`
	Steps       []providerutil.FlowControlStepModel `tfsdk:"steps"`
}

// NewFlowControlPolicyResource returns a new instance of the flow control policy resource
func NewFlowControlPolicyResource() resource.Resource {
	return &FlowControlPolicyResource{}
}

// Metadata returns the resource type name
func (r *FlowControlPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_flow_control_policy"
}

// Schema defines the schema for the flow control policy resource
func (r *FlowControlPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Flow Control Policy in Link11 WAAP.",
		Attributes: map[string]schema.Attribute{
			"config_id": schema.StringAttribute{
				Description: "The configuration ID.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.StringAttribute{
				Description: "The unique identifier for the flow control policy.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the flow control policy.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"description": schema.StringAttribute{
				Description: "Description of the flow control policy.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"active": schema.BoolAttribute{
				Description: "Whether the flow control policy is active.",
				Required:    true,
			},
			"timeframe": schema.Int64Attribute{
				Description: "Time window in seconds within which the flow steps must be completed.",
				Required:    true,
			},
			"tags": schema.ListAttribute{
				Description: "List of tags to apply to matching requests.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"include": schema.ListAttribute{
				Description: "Tags describing requests to include in the flow control policy.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
			},
			"exclude": schema.ListAttribute{
				Description: "Tags describing requests to exclude from the flow control policy.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
			},
		},
		Blocks: map[string]schema.Block{
			"key": schema.ListNestedBlock{
				Description: "Flow control key configuration. At least one block is required. Exactly one of attrs, args, plugins, cookies, or headers must be set per block.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"attrs": schema.StringAttribute{
							Description: "Flow control by request attribute.",
							Optional:    true,
							Validators: []validator.String{
								stringvalidator.OneOf(flowControlAttrsEnum...),
							},
						},
						"args": schema.StringAttribute{
							Description: "Flow control by query argument name.",
							Optional:    true,
						},
						"plugins": schema.StringAttribute{
							Description: "Flow control by plugin data.",
							Optional:    true,
						},
						"cookies": schema.StringAttribute{
							Description: "Flow control by cookie name.",
							Optional:    true,
						},
						"headers": schema.StringAttribute{
							Description: "Flow control by header name.",
							Optional:    true,
						},
					},
				},
			},
			"steps": schema.ListNestedBlock{
				Description: "Ordered steps describing the restricted request flow. At least one step is required.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"method": schema.StringAttribute{
							Description: "HTTP method for this step.",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOf(flowControlMethodEnum...),
							},
						},
						"uri": schema.StringAttribute{
							Description: "URI path for this step.",
							Required:    true,
						},
						"headers": schema.MapAttribute{
							Description: "Header name/value pairs required at this step.",
							Optional:    true,
							ElementType: types.StringType,
						},
						"cookies": schema.MapAttribute{
							Description: "Cookie name/value pairs.",
							Optional:    true,
							ElementType: types.StringType,
						},
						"args": schema.MapAttribute{
							Description: "Query argument name/value pairs.",
							Optional:    true,
							ElementType: types.StringType,
						},
						"plugins": schema.MapAttribute{
							Description: "Plugin key/value pairs.",
							Optional:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
		},
	}
}

// Configure sets the client for the resource
func (r *FlowControlPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = providerutil.ConfigureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new flow control policy using the API client and sets the resource state
func (r *FlowControlPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan FlowControlPolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = types.StringValue(generateIDNoDash())

	policy, diags := buildFlowControlPolicyAPIModel(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.CreateFlowControlPolicy(ctx, plan.ConfigID.ValueString(), plan.ID.ValueString(), policy)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Flow Control Policy",
			"Could not create flow control policy: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read retrieves the flow control policy from the API and updates the resource state. If the policy is not found, it removes the resource from state.
func (r *FlowControlPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state FlowControlPolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policy, err := r.client.GetFlowControlPolicy(ctx, state.ConfigID.ValueString(), state.ID.ValueString())
	if err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Flow Control Policy",
			"Could not read flow control policy: "+err.Error(),
		)
		return
	}

	state.Name = types.StringValue(policy.Name)
	state.Description = types.StringValue(policy.Description)
	state.Active = types.BoolValue(policy.Active)
	state.Timeframe = types.Int64Value(int64(policy.Timeframe))

	// Tags
	if len(policy.Tags) > 0 {
		tagsList, diags := types.ListValueFrom(ctx, types.StringType, policy.Tags)
		resp.Diagnostics.Append(diags...)
		state.Tags = tagsList
	} else {
		state.Tags = types.ListNull(types.StringType)
	}

	// Include
	include := policy.Include
	if include == nil {
		include = []string{}
	}
	includeList, diags := types.ListValueFrom(ctx, types.StringType, include)
	resp.Diagnostics.Append(diags...)
	state.Include = includeList

	// Exclude
	exclude := policy.Exclude
	if exclude == nil {
		exclude = []string{}
	}
	excludeList, diags := types.ListValueFrom(ctx, types.StringType, exclude)
	resp.Diagnostics.Append(diags...)
	state.Exclude = excludeList

	// Key
	state.Key = providerutil.ParseFlowControlKeys(policy.Key)

	// Steps
	state.Steps = providerutil.ParseFlowControlSteps(ctx, policy.Steps, &resp.Diagnostics)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the flow control policy using the API client and sets the resource state
func (r *FlowControlPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan FlowControlPolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	policy, diags := buildFlowControlPolicyAPIModel(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.UpdateFlowControlPolicy(ctx, plan.ConfigID.ValueString(), plan.ID.ValueString(), policy)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Flow Control Policy",
			"Could not update flow control policy: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete deletes the flow control policy using the API client
func (r *FlowControlPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state FlowControlPolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteFlowControlPolicy(ctx, state.ConfigID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Flow Control Policy",
			"Could not delete flow control policy: "+err.Error(),
		)
		return
	}
}

// ImportState imports the flow control policy using the format "config_id/flow_control_policy_id"
func (r *FlowControlPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in format 'config_id/flow_control_policy_id', got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("config_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

// ValidateConfig validates the key and steps blocks.
func (r *FlowControlPolicyResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config FlowControlPolicyResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if len(config.Key) == 0 {
		resp.Diagnostics.AddAttributeError(
			path.Root("key"),
			"Missing Required key Blocks",
			"At least one 'key' block must be specified.",
		)
	}
	for i, k := range config.Key {
		setCount := 0
		if !k.Attrs.IsNull() && !k.Attrs.IsUnknown() {
			setCount++
		}
		if !k.Args.IsNull() && !k.Args.IsUnknown() {
			setCount++
		}
		if !k.Plugins.IsNull() && !k.Plugins.IsUnknown() {
			setCount++
		}
		if !k.Cookies.IsNull() && !k.Cookies.IsUnknown() {
			setCount++
		}
		if !k.Headers.IsNull() && !k.Headers.IsUnknown() {
			setCount++
		}
		if setCount != 1 {
			resp.Diagnostics.AddAttributeError(
				path.Root("key").AtListIndex(i),
				"Invalid key block",
				"Exactly one of attrs, args, plugins, cookies, or headers must be set in each 'key' block.",
			)
		}
	}

	if len(config.Steps) == 0 {
		resp.Diagnostics.AddAttributeError(
			path.Root("steps"),
			"Missing Required steps Blocks",
			"At least one 'steps' block must be specified.",
		)
	}
}

func buildFlowControlPolicyAPIModel(ctx context.Context, plan *FlowControlPolicyResourceModel) (*client.FlowControl, diag.Diagnostics) {
	var diags diag.Diagnostics

	policy := &client.FlowControl{
		ID:          plan.ID.ValueString(),
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		Active:      plan.Active.ValueBool(),
		Timeframe:   int(plan.Timeframe.ValueInt64()),
	}

	// Tags
	if !plan.Tags.IsNull() && !plan.Tags.IsUnknown() {
		diags.Append(plan.Tags.ElementsAs(ctx, &policy.Tags, false)...)
	}

	// Include
	policy.Include = []string{}
	if !plan.Include.IsNull() && !plan.Include.IsUnknown() {
		diags.Append(plan.Include.ElementsAs(ctx, &policy.Include, false)...)
	}

	// Exclude
	policy.Exclude = []string{}
	if !plan.Exclude.IsNull() && !plan.Exclude.IsUnknown() {
		diags.Append(plan.Exclude.ElementsAs(ctx, &policy.Exclude, false)...)
	}

	// Key
	policy.Key = buildFlowControlKeys(plan.Key)

	// Steps
	policy.Steps = buildFlowControlSteps(ctx, plan.Steps, &diags)

	return policy, diags
}

// buildFlowControlKeys converts []providerutil.FlowControlKeyModel to []client.FlowControlKeyEntry for the API.
// Exactly one field per entry is populated.
func buildFlowControlKeys(keys []providerutil.FlowControlKeyModel) []client.FlowControlKeyEntry {
	result := make([]client.FlowControlKeyEntry, 0, len(keys))
	for _, k := range keys {
		var entry client.FlowControlKeyEntry
		switch {
		case !k.Attrs.IsNull() && !k.Attrs.IsUnknown():
			v := k.Attrs.ValueString()
			entry.Attrs = &v
		case !k.Args.IsNull() && !k.Args.IsUnknown():
			v := k.Args.ValueString()
			entry.Args = &v
		case !k.Plugins.IsNull() && !k.Plugins.IsUnknown():
			v := k.Plugins.ValueString()
			entry.Plugins = &v
		case !k.Cookies.IsNull() && !k.Cookies.IsUnknown():
			v := k.Cookies.ValueString()
			entry.Cookies = &v
		case !k.Headers.IsNull() && !k.Headers.IsUnknown():
			v := k.Headers.ValueString()
			entry.Headers = &v
		}
		result = append(result, entry)
	}
	return result
}

// buildFlowControlSteps converts []providerutil.FlowControlStepModel to []client.FlowStepItem for the API.
func buildFlowControlSteps(ctx context.Context, steps []providerutil.FlowControlStepModel, diags *diag.Diagnostics) []client.FlowStepItem {
	result := make([]client.FlowStepItem, 0, len(steps))
	for _, s := range steps {
		// These 3 maps must be pre-initialized to empty maps, otherwise
		// UI will not display them correctly when they are empty.
		item := client.FlowStepItem{
			Method:  s.Method.ValueString(),
			URI:     s.URI.ValueString(),
			Headers: map[string]string{},
			Cookies: map[string]string{},
			Args:    map[string]string{},
		}
		if !s.Headers.IsNull() && !s.Headers.IsUnknown() {
			m := map[string]string{}
			diags.Append(s.Headers.ElementsAs(ctx, &m, false)...)
			item.Headers = m
		}
		if !s.Cookies.IsNull() && !s.Cookies.IsUnknown() {
			m := map[string]string{}
			diags.Append(s.Cookies.ElementsAs(ctx, &m, false)...)
			item.Cookies = m
		}
		if !s.Args.IsNull() && !s.Args.IsUnknown() {
			m := map[string]string{}
			diags.Append(s.Args.ElementsAs(ctx, &m, false)...)
			item.Args = m
		}
		if !s.Plugins.IsNull() && !s.Plugins.IsUnknown() {
			m := map[string]string{}
			diags.Append(s.Plugins.ElementsAs(ctx, &m, false)...)
			item.Plugins = m
		}
		result = append(result, item)
	}
	return result
}
