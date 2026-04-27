// Package resources contains the implementation of Terraform resources for Link11 WAAP.
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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/link11/terraform-provider-link11waap/internal/client"
	"github.com/link11/terraform-provider-link11waap/internal/providerutil"
)

var (
	_ resource.Resource                = &GlobalFilterResource{}
	_ resource.ResourceWithImportState = &GlobalFilterResource{}
)

// GlobalFilterResource implements the global filter resource.
type GlobalFilterResource struct {
	client *client.Client
}

// GlobalFilterResourceModel describes the global filter resource data model.
type GlobalFilterResourceModel struct {
	ConfigID    types.String `tfsdk:"config_id"`
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Source      types.String `tfsdk:"source"`
	Mdate       types.String `tfsdk:"mdate"`
	Active      types.Bool   `tfsdk:"active"`
	Tags        types.List   `tfsdk:"tags"`
	Action      types.String `tfsdk:"action"`
	Rule        *RuleModel   `tfsdk:"rule"`
}

// RuleModel describes the rule block for a global filter.
type RuleModel struct {
	Relation types.String `tfsdk:"relation"`
	Entries  []EntryModel `tfsdk:"entry"`
	Groups   []GroupModel `tfsdk:"group"`
}

// EntryModel describes a leaf condition entry in a rule.
type EntryModel struct {
	Type    types.String `tfsdk:"type"`
	Name    types.String `tfsdk:"name"`
	Value   types.String `tfsdk:"value"`
	Comment types.String `tfsdk:"comment"`
}

// GroupModel describes a nested rule group.
type GroupModel struct {
	Relation types.String `tfsdk:"relation"`
	Entries  []EntryModel `tfsdk:"entry"`
}

// NewGlobalFilterResource creates a new global filter resource instance.
func NewGlobalFilterResource() resource.Resource {
	return &GlobalFilterResource{}
}

// Metadata returns the resource type name.
func (r *GlobalFilterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_global_filter"
}

// Schema defines the schema for the global filter resource.
func (r *GlobalFilterResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Global Filter in Link11 WAAP.",
		Attributes: map[string]schema.Attribute{
			"config_id": schema.StringAttribute{
				Description: "The configuration ID.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.StringAttribute{
				Description: "The unique identifier for the global filter.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the global filter.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the global filter.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"source": schema.StringAttribute{
				Description: "The source of the global filter (always self-managed for user-created filters).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"mdate": schema.StringAttribute{
				Description: "The last modification date (server-managed).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"active": schema.BoolAttribute{
				Description: "Whether the global filter is active.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"tags": schema.ListAttribute{
				Description: "List of tags associated with the global filter.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"action": schema.StringAttribute{
				Description: "Action to take when requests match this filter. Defaults to action-monitor. Allowed values: action-challenge, action-monitor, action-skip, action-global-filter-block, action-rate-limit-block, action-acl-block, action-contentfilter-block, action-dynamic-rule-block, action-waap-feed-block, action-https-redirect.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("action-monitor"),
				Validators: []validator.String{
					stringvalidator.OneOf(
						"action-challenge",
						"action-monitor",
						"action-skip",
						"action-global-filter-block",
						"action-rate-limit-block",
						"action-acl-block",
						"action-contentfilter-block",
						"action-dynamic-rule-block",
						"action-waap-feed-block",
						"action-https-redirect",
					),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"rule": schema.SingleNestedBlock{
				Description: "The rule definition for matching requests.",
				Attributes: map[string]schema.Attribute{
					"relation": schema.StringAttribute{
						Description: "Logical relation for combining entries. Must be OR or AND.",
						Required:    true,
						Validators: []validator.String{
							stringvalidator.OneOf("OR", "AND"),
						},
					},
				},
				Blocks: map[string]schema.Block{
					"entry": schema.ListNestedBlock{
						Description: "A leaf condition entry. Use 'name' for types that match against a named field (headers, cookies, args).",
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"type": schema.StringAttribute{
									Required:    true,
									Description: "Entry category. Valid values from AttributesEnum: asn, authority, company, cookies, country, headers, ip, method, network, organization, path, query, region, secpolentryid, secpolid, secpolentryname, secpolname, securitypolicy, securitypolicyentry, securitypolicyentryid, securitypolicyentryname, securitypolicyid, securitypolicyname, session, subregion, tags, uri. Use 'name' + 'value' for types that match against a named field (e.g. headers, cookies).",
									Validators: []validator.String{
										stringvalidator.OneOf(
											"asn", "authority", "company", "cookies", "country", "headers",
											"ip", "method", "network", "organization", "path", "query",
											"region", "secpolentryid", "secpolid", "secpolentryname",
											"secpolname", "securitypolicy", "securitypolicyentry",
											"securitypolicyentryid", "securitypolicyentryname",
											"securitypolicyid", "securitypolicyname", "session",
											"subregion", "tags", "uri",
										),
									},
								},
								"name":  schema.StringAttribute{Optional: true, Description: "For entry types that use a key-value pair (e.g. headers, cookies): the field name (e.g. header name or cookie name). When set, the entry is sent as [type, [name, value], comment] to the API."},
								"value": schema.StringAttribute{Required: true, Description: "Value or regex to match."},
								"comment": schema.StringAttribute{
									Optional:    true,
									Computed:    true,
									Default:     stringdefault.StaticString(""),
									Description: "Human-readable comment for this entry.",
								},
							},
						},
					},
					"group": schema.ListNestedBlock{
						Description: "A nested rule group, combining entries with its own relation.",
						NestedObject: schema.NestedBlockObject{
							Attributes: map[string]schema.Attribute{
								"relation": schema.StringAttribute{
									Required:    true,
									Description: "Logical relation for combining entries in this group. Must be OR or AND.",
									Validators: []validator.String{
										stringvalidator.OneOf("OR", "AND"),
									},
								},
							},
							Blocks: map[string]schema.Block{
								"entry": schema.ListNestedBlock{
									Description: "A leaf condition entry within this group.",
									NestedObject: schema.NestedBlockObject{
										Attributes: map[string]schema.Attribute{
											"type": schema.StringAttribute{
												Required:    true,
												Description: "Entry category. Valid values from AttributesEnum: asn, authority, company, cookies, country, headers, ip, method, network, organization, path, query, region, secpolentryid, secpolid, secpolentryname, secpolname, securitypolicy, securitypolicyentry, securitypolicyentryid, securitypolicyentryname, securitypolicyid, securitypolicyname, session, subregion, tags, uri. Use 'name' + 'value' for types that match against a named field (e.g. headers, cookies).",
												Validators: []validator.String{
													stringvalidator.OneOf(
														"asn", "authority", "company", "cookies", "country", "headers",
														"ip", "method", "network", "organization", "path", "query",
														"region", "secpolentryid", "secpolid", "secpolentryname",
														"secpolname", "securitypolicy", "securitypolicyentry",
														"securitypolicyentryid", "securitypolicyentryname",
														"securitypolicyid", "securitypolicyname", "session",
														"subregion", "tags", "uri",
													),
												},
											},
											"name":  schema.StringAttribute{Optional: true, Description: "For entry types that use a key-value pair (e.g. headers, cookies): the field name (e.g. header name or cookie name). When set, the entry is sent as [type, [name, value], comment] to the API."},
											"value": schema.StringAttribute{Required: true},
											"comment": schema.StringAttribute{
												Optional:    true,
												Computed:    true,
												Default:     stringdefault.StaticString(""),
												Description: "Human-readable comment.",
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

// Configure configures the resource with the provider client.
func (r *GlobalFilterResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = providerutil.ConfigureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new global filter resource.
func (r *GlobalFilterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan GlobalFilterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = types.StringValue(generateID())

	filter, diags := buildGlobalFilterAPIModel(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.CreateGlobalFilter(ctx, plan.ConfigID.ValueString(), plan.ID.ValueString(), filter)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Global Filter",
			"Could not create global filter: "+err.Error(),
		)
		return
	}

	// mdate is server-managed, set empty after create
	plan.Mdate = types.StringValue("")
	plan.Source = types.StringValue("self-managed")

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read reads the global filter resource.
func (r *GlobalFilterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state GlobalFilterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	filter, err := r.client.GetGlobalFilter(ctx, state.ConfigID.ValueString(), state.ID.ValueString())
	if err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Global Filter",
			"Could not read global filter: "+err.Error(),
		)
		return
	}

	state.Name = types.StringValue(filter.Name)
	state.Description = types.StringValue(filter.Description)
	state.Source = types.StringValue(filter.Source)
	state.Mdate = types.StringValue(filter.Mdate)
	state.Active = types.BoolValue(filter.Active)

	// Tags
	if len(filter.Tags) > 0 {
		tagsList, diags := types.ListValueFrom(ctx, types.StringType, filter.Tags)
		resp.Diagnostics.Append(diags...)
		state.Tags = tagsList
	} else {
		state.Tags = types.ListNull(types.StringType)
	}

	action := filter.Action
	if action == "" {
		action = "action-monitor"
	}
	state.Action = types.StringValue(action)

	// Rule: interface{} -> RuleModel
	ruleModel, parseErr := apiRuleToModel(filter.Rule)
	if parseErr != nil {
		resp.Diagnostics.AddError("Error Parsing Rule", parseErr.Error())
		return
	}
	state.Rule = ruleModel

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the global filter resource.
func (r *GlobalFilterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan GlobalFilterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	filter, diags := buildGlobalFilterAPIModel(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.UpdateGlobalFilter(ctx, plan.ConfigID.ValueString(), plan.ID.ValueString(), filter)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Global Filter",
			"Could not update global filter: "+err.Error(),
		)
		return
	}

	plan.Source = types.StringValue("self-managed")

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete deletes the global filter resource.
func (r *GlobalFilterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state GlobalFilterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteGlobalFilter(ctx, state.ConfigID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Global Filter",
			"Could not delete global filter: "+err.Error(),
		)
		return
	}
}

// ImportState imports an existing global filter resource.
func (r *GlobalFilterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in format 'config_id/global_filter_id', got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("config_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

// buildGlobalFilterAPIModel converts the Terraform resource model to the API model.
func buildGlobalFilterAPIModel(ctx context.Context, plan *GlobalFilterResourceModel) (*client.GlobalFilter, diag.Diagnostics) {
	var diags diag.Diagnostics

	filter := &client.GlobalFilter{
		ID:          plan.ID.ValueString(),
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		Source:      "self-managed",
		Active:      plan.Active.ValueBool(),
	}

	// Tags
	if !plan.Tags.IsNull() && !plan.Tags.IsUnknown() {
		diags.Append(plan.Tags.ElementsAs(ctx, &filter.Tags, false)...)
	}

	// Action: plain string
	filter.Action = plan.Action.ValueString()

	// Rule: RuleModel -> API wire format
	filter.Rule = ruleModelToAPI(plan.Rule)

	return filter, diags
}

// ruleModelToAPI converts a RuleModel to the API wire format (interface{} ready for JSON marshal).
// Entries appear before groups in the entries array.
func ruleModelToAPI(rule *RuleModel) interface{} {
	if rule == nil {
		return nil
	}
	entries := make([]interface{}, 0)
	for _, e := range rule.Entries {
		entries = append(entries, entryModelToAPI(e))
	}
	for _, g := range rule.Groups {
		entries = append(entries, groupModelToAPI(g))
	}
	return map[string]interface{}{
		"relation": rule.Relation.ValueString(),
		"entries":  entries,
	}
}

func entryModelToAPI(e EntryModel) interface{} {
	comment := ""
	if !e.Comment.IsNull() && !e.Comment.IsUnknown() {
		comment = e.Comment.ValueString()
	}
	name := ""
	if !e.Name.IsNull() && !e.Name.IsUnknown() {
		name = e.Name.ValueString()
	}
	if name != "" {
		return []interface{}{e.Type.ValueString(), []interface{}{name, e.Value.ValueString()}, comment}
	}
	return []interface{}{e.Type.ValueString(), e.Value.ValueString(), comment}
}

func groupModelToAPI(g GroupModel) interface{} {
	entries := make([]interface{}, 0)
	for _, e := range g.Entries {
		entries = append(entries, entryModelToAPI(e))
	}
	return map[string]interface{}{
		"relation": g.Relation.ValueString(),
		"entries":  entries,
	}
}

// apiRuleToModel parses the API response into a RuleModel.
// Returns nil, nil when rawRule is nil.
func apiRuleToModel(rawRule interface{}) (*RuleModel, error) {
	if rawRule == nil {
		return nil, nil
	}
	ruleMap, ok := rawRule.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("rule is not an object, got %T", rawRule)
	}
	model := &RuleModel{}
	if rel, ok := ruleMap["relation"].(string); ok {
		model.Relation = types.StringValue(rel)
	}
	rawEntries, _ := ruleMap["entries"].([]interface{})
	for _, rawEntry := range rawEntries {
		switch e := rawEntry.(type) {
		case []interface{}:
			em, err := apiEntryToModel(e)
			if err != nil {
				return nil, err
			}
			model.Entries = append(model.Entries, em)
		case map[string]interface{}:
			gm, err := apiGroupToModel(e)
			if err != nil {
				return nil, err
			}
			model.Groups = append(model.Groups, *gm)
		}
	}
	return model, nil
}

func apiEntryToModel(raw []interface{}) (EntryModel, error) {
	if len(raw) < 2 {
		return EntryModel{}, fmt.Errorf("entry must have at least 2 elements, got %d", len(raw))
	}
	em := EntryModel{Name: types.StringNull()}
	if t, ok := raw[0].(string); ok {
		em.Type = types.StringValue(t)
	}
	comment := ""
	if len(raw) >= 3 {
		if c, ok := raw[2].(string); ok {
			comment = c
		}
	}
	em.Comment = types.StringValue(comment)
	switch v := raw[1].(type) {
	case string:
		em.Value = types.StringValue(v)
	case []interface{}:
		if len(v) >= 2 {
			if name, ok := v[0].(string); ok {
				em.Name = types.StringValue(name)
			}
			if val, ok := v[1].(string); ok {
				em.Value = types.StringValue(val)
			}
		}
	default:
		em.Value = types.StringValue(fmt.Sprintf("%v", v))
	}
	return em, nil
}

func apiGroupToModel(raw map[string]interface{}) (*GroupModel, error) {
	gm := &GroupModel{}
	if rel, ok := raw["relation"].(string); ok {
		gm.Relation = types.StringValue(rel)
	}
	rawEntries, _ := raw["entries"].([]interface{})
	for _, rawEntry := range rawEntries {
		if entryArr, ok := rawEntry.([]interface{}); ok {
			em, err := apiEntryToModel(entryArr)
			if err != nil {
				return nil, err
			}
			gm.Entries = append(gm.Entries, em)
		}
	}
	return gm, nil
}
