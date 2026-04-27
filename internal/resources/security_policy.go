package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/link11/terraform-provider-link11waap/internal/client"
	"github.com/link11/terraform-provider-link11waap/internal/providerutil"
)

var (
	_ resource.Resource                   = &SecurityPolicyResource{}
	_ resource.ResourceWithImportState    = &SecurityPolicyResource{}
	_ resource.ResourceWithValidateConfig = &SecurityPolicyResource{}
	_ resource.ResourceWithModifyPlan     = &SecurityPolicyResource{}
)

// SecurityPolicyResource implements the security policy resource.
type SecurityPolicyResource struct {
	client *client.Client
}

// SecurityPolicyResourceModel describes the security policy resource data model.
type SecurityPolicyResourceModel struct {
	ConfigID    types.String         `tfsdk:"config_id"`
	ID          types.String         `tfsdk:"id"`
	Name        types.String         `tfsdk:"name"`
	Description types.String         `tfsdk:"description"`
	Tags        types.List           `tfsdk:"tags"`
	Map         []SecProfileMapModel `tfsdk:"map"`
	Session     []SessionKeyModel    `tfsdk:"session"`
	SessionIDs  []SessionKeyModel    `tfsdk:"session_ids"`
}

// SessionKeyModel describes the data model for a session key entry.
type SessionKeyModel struct {
	Attrs   types.String `tfsdk:"attrs"`
	Args    types.String `tfsdk:"args"`
	Plugins types.String `tfsdk:"plugins"`
	Cookies types.String `tfsdk:"cookies"`
	Headers types.String `tfsdk:"headers"`
}

// SecProfileMapModel describes the data model for a security profile map entry.
type SecProfileMapModel struct {
	ID                         types.String `tfsdk:"id"`
	Name                       types.String `tfsdk:"name"`
	Match                      types.String `tfsdk:"match"`
	ACLProfile                 types.String `tfsdk:"acl_profile"`
	ACLProfileActive           types.Bool   `tfsdk:"acl_profile_active"`
	ContentFilterProfile       types.String `tfsdk:"content_filter_profile"`
	ContentFilterProfileActive types.Bool   `tfsdk:"content_filter_profile_active"`
	BackendService             types.String `tfsdk:"backend_service"`
	Description                types.String `tfsdk:"description"`
	RateLimitRules             types.List   `tfsdk:"rate_limit_rules"`
	EdgeFunctions              types.List   `tfsdk:"edge_functions"`
}

// NewSecurityPolicyResource creates a new security policy resource instance.
func NewSecurityPolicyResource() resource.Resource {
	return &SecurityPolicyResource{}
}

// Metadata returns the resource type name.
func (r *SecurityPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security_policy"
}

// Schema defines the schema for the security policy resource.
func (r *SecurityPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Security Policy in Link11 WAAP.",
		Attributes: map[string]schema.Attribute{
			"config_id": schema.StringAttribute{
				Description: "The configuration ID.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.StringAttribute{
				Description: "The unique identifier for the security policy.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the security policy.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the security policy.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"tags": schema.ListAttribute{
				Description: "List of tags associated with the security policy.",
				Optional:    true,
				ElementType: types.StringType,
			},
		},
		Blocks: map[string]schema.Block{
			"map": schema.SetNestedBlock{
				Description: "Security profile map entries.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Map entry ID.",
							Required:    true,
						},
						"name": schema.StringAttribute{
							Description: "Map entry name.",
							Required:    true,
						},
						"match": schema.StringAttribute{
							Description: "Match expression (path/header matching).",
							Required:    true,
						},
						"acl_profile": schema.StringAttribute{
							Description: "ID of the ACL profile to apply.",
							Required:    true,
						},
						"acl_profile_active": schema.BoolAttribute{
							Description: "Whether the ACL profile is active.",
							Required:    true,
						},
						"content_filter_profile": schema.StringAttribute{
							Description: "ID of the content filter profile to apply.",
							Required:    true,
						},
						"content_filter_profile_active": schema.BoolAttribute{
							Description: "Whether the content filter profile is active.",
							Required:    true,
						},
						"backend_service": schema.StringAttribute{
							Description: "ID of the backend service.",
							Required:    true,
						},
						"description": schema.StringAttribute{
							Description: "Description of the map entry.",
							Optional:    true,
							Computed:    true,
							Default:     stringdefault.StaticString(""),
						},
						"rate_limit_rules": schema.ListAttribute{
							Description: "List of rate limit rule IDs to apply.",
							Optional:    true,
							ElementType: types.StringType,
						},
						"edge_functions": schema.ListAttribute{
							Description: "List of edge function IDs to apply.",
							Optional:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
			"session": schema.ListNestedBlock{
				Description: "Session key configuration. Exactly one block is required.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"attrs":   schema.StringAttribute{Description: "Session by request attribute (e.g., 'ip', 'authority').", Optional: true},
						"args":    schema.StringAttribute{Description: "Session by query argument name.", Optional: true},
						"plugins": schema.StringAttribute{Description: "Session by plugin data.", Optional: true},
						"cookies": schema.StringAttribute{Description: "Session by cookie name.", Optional: true},
						"headers": schema.StringAttribute{Description: "Session by header name.", Optional: true},
					},
				},
			},
			"session_ids": schema.ListNestedBlock{
				Description: "Session IDs key configuration. Optional, null by default.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"attrs":   schema.StringAttribute{Description: "Session ID by request attribute.", Optional: true},
						"args":    schema.StringAttribute{Description: "Session ID by query argument name.", Optional: true},
						"plugins": schema.StringAttribute{Description: "Session ID by plugin data.", Optional: true},
						"cookies": schema.StringAttribute{Description: "Session ID by cookie name.", Optional: true},
						"headers": schema.StringAttribute{Description: "Session ID by header name.", Optional: true},
					},
				},
			},
		},
	}
}

// Configure configures the resource with the provider client.
func (r *SecurityPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = providerutil.ConfigureClient(req.ProviderData, &resp.Diagnostics)
}

// ValidateConfig validates the security policy configuration.
func (r *SecurityPolicyResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config SecurityPolicyResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// session must have exactly one block
	if len(config.Session) != 1 {
		resp.Diagnostics.AddAttributeError(
			path.Root("session"),
			"Invalid session Configuration",
			"Exactly one 'session' block must be specified.",
		)
		return
	}

	// Validate each session entry has exactly one field set
	for i, s := range config.Session {
		setCount := countSessionKeyFields(s)
		if setCount != 1 {
			resp.Diagnostics.AddAttributeError(
				path.Root("session").AtListIndex(i),
				"Invalid session block",
				"Exactly one of attrs, args, plugins, cookies, or headers must be set in each 'session' block.",
			)
		}
	}

	// Validate each session_ids entry has exactly one field set
	for i, s := range config.SessionIDs {
		setCount := countSessionKeyFields(s)
		if setCount != 1 {
			resp.Diagnostics.AddAttributeError(
				path.Root("session_ids").AtListIndex(i),
				"Invalid session_ids block",
				"Exactly one of attrs, args, plugins, cookies, or headers must be set in each 'session_ids' block.",
			)
		}
	}
}

// ModifyPlan ensures that server-managed default maps (like __site_level__) are preserved
// in the plan when they exist in the prior state, preventing spurious diffs.
func (r *SecurityPolicyResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Only handle updates: both state and plan must exist.
	if req.Plan.Raw.IsNull() || req.State.Raw.IsNull() {
		return
	}

	var state SecurityPolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var plan SecurityPolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	merged := mergeDefaultMaps(state.Map, plan.Map)
	if merged != nil {
		plan.Map = merged
		resp.Diagnostics.Append(resp.Plan.Set(ctx, plan)...)
	}
}

// countSessionKeyFields counts the number of non-null, non-unknown fields in a SessionKeyModel.
func countSessionKeyFields(k SessionKeyModel) int {
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
	return setCount
}

// Create creates a new security policy resource.
func (r *SecurityPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SecurityPolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = types.StringValue(generateID())

	sp := buildSecurityPolicyAPIModel(ctx, &plan)

	err := r.client.CreateSecurityPolicy(ctx, plan.ConfigID.ValueString(), plan.ID.ValueString(), sp)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Security Policy",
			"Could not create security policy: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read reads the security policy resource.
func (r *SecurityPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SecurityPolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sp, err := r.client.GetSecurityPolicy(ctx, state.ConfigID.ValueString(), state.ID.ValueString())
	if err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Security Policy",
			"Could not read security policy: "+err.Error(),
		)
		return
	}

	state.Name = types.StringValue(sp.Name)
	state.Description = types.StringValue(sp.Description)

	// Tags
	if len(sp.Tags) > 0 {
		tagsList, diags := types.ListValueFrom(ctx, types.StringType, sp.Tags)
		resp.Diagnostics.Append(diags...)
		state.Tags = tagsList
	} else {
		state.Tags = types.ListNull(types.StringType)
	}

	// Session
	state.Session = parseSessionKeys(sp.Session)

	// SessionIDs
	if sp.SessionIDs != nil {
		state.SessionIDs = parseSessionKeys(sp.SessionIDs)
	} else {
		state.SessionIDs = []SessionKeyModel{}
	}

	// Map
	state.Map = make([]SecProfileMapModel, len(sp.Map))
	for i, m := range sp.Map {
		mapModel := SecProfileMapModel{
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

		// Normalize nil slices to empty slices to avoid diffs between nil and []
		// It's applied only for rate_limit_rules and edge_functions
		if len(m.RateLimitRules) > 0 {
			rateLimitsList, diags := types.ListValueFrom(ctx, types.StringType, m.RateLimitRules)
			resp.Diagnostics.Append(diags...)
			mapModel.RateLimitRules = rateLimitsList
		} else {
			mapModel.RateLimitRules = types.ListNull(types.StringType)
		}

		if len(m.EdgeFunctions) > 0 {
			edgeFunctionsList, diags := types.ListValueFrom(ctx, types.StringType, m.EdgeFunctions)
			resp.Diagnostics.Append(diags...)
			mapModel.EdgeFunctions = edgeFunctionsList
		} else {
			mapModel.EdgeFunctions = types.ListNull(types.StringType)
		}
		state.Map[i] = mapModel
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the security policy resource.
func (r *SecurityPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SecurityPolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sp := buildSecurityPolicyAPIModel(ctx, &plan)

	err := r.client.UpdateSecurityPolicy(ctx, plan.ConfigID.ValueString(), plan.ID.ValueString(), sp)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Security Policy",
			"Could not update security policy: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete deletes the security policy resource.
func (r *SecurityPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SecurityPolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteSecurityPolicy(ctx, state.ConfigID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Security Policy",
			"Could not delete security policy: "+err.Error(),
		)
		return
	}
}

// ImportState imports an existing security policy resource.
func (r *SecurityPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in format 'config_id/security_policy_id', got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("config_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

// buildSecurityPolicyAPIModel converts the Terraform resource model into the API client struct.
func buildSecurityPolicyAPIModel(ctx context.Context, plan *SecurityPolicyResourceModel) *client.SecurityPolicy {
	sp := &client.SecurityPolicy{
		ID:          plan.ID.ValueString(),
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
	}

	// Tags
	if !plan.Tags.IsNull() && !plan.Tags.IsUnknown() {
		var tags []string
		plan.Tags.ElementsAs(ctx, &tags, false)
		sp.Tags = tags
	}

	// Session
	sp.Session = buildSessionKeys(plan.Session)

	// SessionIDs
	if len(plan.SessionIDs) == 0 {
		sp.SessionIDs = []map[string]string{}
	} else {
		sp.SessionIDs = buildSessionKeys(plan.SessionIDs)
	}

	// Map
	if len(plan.Map) > 0 {
		sp.Map = make([]client.SecProfileMap, len(plan.Map))
		for i, m := range plan.Map {
			sp.Map[i] = buildSecProfileMapEntry(ctx, m)
		}
	}

	return sp
}

// buildSecProfileMapEntry converts a single SecProfileMapModel into the API client struct.
func buildSecProfileMapEntry(ctx context.Context, m SecProfileMapModel) client.SecProfileMap {
	entry := client.SecProfileMap{
		ID:                         m.ID.ValueString(),
		Name:                       m.Name.ValueString(),
		Match:                      m.Match.ValueString(),
		ACLProfile:                 m.ACLProfile.ValueString(),
		ACLProfileActive:           m.ACLProfileActive.ValueBool(),
		ContentFilterProfile:       m.ContentFilterProfile.ValueString(),
		ContentFilterProfileActive: m.ContentFilterProfileActive.ValueBool(),
		BackendService:             m.BackendService.ValueString(),
		Description:                m.Description.ValueString(),
	}

	if !m.RateLimitRules.IsNull() && !m.RateLimitRules.IsUnknown() {
		m.RateLimitRules.ElementsAs(ctx, &entry.RateLimitRules, false)
	}
	if !m.EdgeFunctions.IsNull() && !m.EdgeFunctions.IsUnknown() {
		m.EdgeFunctions.ElementsAs(ctx, &entry.EdgeFunctions, false)
	}

	return entry
}

// mergeDefaultMaps returns a new map slice that includes server-managed default maps
// from the prior state that are absent from the planned maps, preventing spurious diffs.
// Returns nil if no changes are needed.
func mergeDefaultMaps(stateMaps, planMaps []SecProfileMapModel) []SecProfileMapModel {
	const defaultMapID = "__site_level__"

	// Find the default map in state.
	var defaultMap *SecProfileMapModel
	for i := range stateMaps {
		if stateMaps[i].ID.ValueString() == defaultMapID {
			m := stateMaps[i]
			defaultMap = &m
			break
		}
	}
	if defaultMap == nil {
		return nil
	}

	// Check if default map is already in the plan.
	for _, m := range planMaps {
		if m.ID.ValueString() == defaultMapID {
			return nil
		}
	}

	// Append the default map to the plan.
	merged := make([]SecProfileMapModel, len(planMaps)+1)
	copy(merged, planMaps)
	merged[len(planMaps)] = *defaultMap
	return merged
}

// parseSessionKeys converts the API interface{} to []SessionKeyModel.
func parseSessionKeys(raw interface{}) []SessionKeyModel {
	if raw == nil {
		return []SessionKeyModel{}
	}
	rawKeys, ok := raw.([]interface{})
	if !ok {
		return []SessionKeyModel{}
	}
	keys := make([]SessionKeyModel, 0, len(rawKeys))
	for _, rk := range rawKeys {
		m, ok := rk.(map[string]interface{})
		if !ok {
			continue
		}
		km := SessionKeyModel{
			Attrs:   types.StringNull(),
			Args:    types.StringNull(),
			Plugins: types.StringNull(),
			Cookies: types.StringNull(),
			Headers: types.StringNull(),
		}
		if v, ok := m["attrs"].(string); ok {
			km.Attrs = types.StringValue(v)
		} else if v, ok := m["args"].(string); ok {
			km.Args = types.StringValue(v)
		} else if v, ok := m["plugins"].(string); ok {
			km.Plugins = types.StringValue(v)
		} else if v, ok := m["cookies"].(string); ok {
			km.Cookies = types.StringValue(v)
		} else if v, ok := m["headers"].(string); ok {
			km.Headers = types.StringValue(v)
		}
		keys = append(keys, km)
	}
	return keys
}

// buildSessionKeys converts []SessionKeyModel to []map[string]string for the API.
func buildSessionKeys(keys []SessionKeyModel) []map[string]string {
	result := make([]map[string]string, 0, len(keys))
	for _, k := range keys {
		if !k.Attrs.IsNull() && !k.Attrs.IsUnknown() {
			result = append(result, map[string]string{"attrs": k.Attrs.ValueString()})
		} else if !k.Args.IsNull() && !k.Args.IsUnknown() {
			result = append(result, map[string]string{"args": k.Args.ValueString()})
		} else if !k.Plugins.IsNull() && !k.Plugins.IsUnknown() {
			result = append(result, map[string]string{"plugins": k.Plugins.ValueString()})
		} else if !k.Cookies.IsNull() && !k.Cookies.IsUnknown() {
			result = append(result, map[string]string{"cookies": k.Cookies.ValueString()})
		} else if !k.Headers.IsNull() && !k.Headers.IsUnknown() {
			result = append(result, map[string]string{"headers": k.Headers.ValueString()})
		}
	}
	return result
}
