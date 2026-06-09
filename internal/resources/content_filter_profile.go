package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/link11/terraform-provider-link11waap/internal/client"
	"github.com/link11/terraform-provider-link11waap/internal/providerutil"
)

var (
	_ resource.Resource                = &ContentFilterProfileResource{}
	_ resource.ResourceWithImportState = &ContentFilterProfileResource{}
)

// ContentFilterProfileResource implements the content filter profile resource.
type ContentFilterProfileResource struct {
	client *client.Client
}

// ContentFilterProfileResourceModel describes the content filter profile resource data model.
type ContentFilterProfileResourceModel struct {
	ConfigID       types.String `tfsdk:"config_id"`
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	IgnoreAlphanum types.Bool   `tfsdk:"ignore_alphanum"`
	MaskingSeed    types.String `tfsdk:"masking_seed"`
	ContentType    types.List   `tfsdk:"content_type"`
	GraphqlPath    types.String `tfsdk:"graphql_path"`
	IgnoreBody     types.Bool   `tfsdk:"ignore_body"`
	Active         types.List   `tfsdk:"active"`
	Report         types.List   `tfsdk:"report"`
	Ignore         types.List   `tfsdk:"ignore"`
	Tags           types.List   `tfsdk:"tags"`
	Action         types.String `tfsdk:"action"`
	Args           types.Object `tfsdk:"args"`
	Headers        types.Object `tfsdk:"headers"`
	Cookies        types.Object `tfsdk:"cookies"`
	Path           types.Object `tfsdk:"path"`
	URL            types.Object `tfsdk:"url"`
	AllSections    types.Object `tfsdk:"allsections"`
	Decoding       types.Object `tfsdk:"decoding"`
}

// cfEntryMatchModel mirrors the section matcher object schema.
type cfEntryMatchModel struct {
	ID               types.String `tfsdk:"id"`
	Key              types.String `tfsdk:"key"`
	Reg              types.String `tfsdk:"reg"`
	Restrict         types.Bool   `tfsdk:"restrict"`
	Mask             types.Bool   `tfsdk:"mask"`
	IgnoreCFRuleTags types.List   `tfsdk:"ignore_cf_rule_tags"`
	Domain           types.String `tfsdk:"domain"`
	Path             types.String `tfsdk:"path"`
	CaseInsensitive  types.Bool   `tfsdk:"case_insensitive"`
	Active           types.Bool   `tfsdk:"active"`
}

// cfSectionModel mirrors a section object schema.
type cfSectionModel struct {
	MaxCount        types.Int64 `tfsdk:"max_count"`
	MaxLength       types.Int64 `tfsdk:"max_length"`
	EnableMaxCount  types.Bool  `tfsdk:"enable_max_count"`
	EnableMaxLength types.Bool  `tfsdk:"enable_max_length"`
	Names           types.List  `tfsdk:"names"`
	Regex           types.List  `tfsdk:"regex"`
	Text            types.List  `tfsdk:"text"`
}

// cfURLSectionModel mirrors the url section (no names).
type cfURLSectionModel struct {
	MaxCount        types.Int64 `tfsdk:"max_count"`
	MaxLength       types.Int64 `tfsdk:"max_length"`
	EnableMaxCount  types.Bool  `tfsdk:"enable_max_count"`
	EnableMaxLength types.Bool  `tfsdk:"enable_max_length"`
	Regex           types.List  `tfsdk:"regex"`
	Text            types.List  `tfsdk:"text"`
}

// cfDecodingModel mirrors the decoding object schema.
type cfDecodingModel struct {
	Base64  types.Bool `tfsdk:"base64"`
	Dual    types.Bool `tfsdk:"dual"`
	HTML    types.Bool `tfsdk:"html"`
	Unicode types.Bool `tfsdk:"unicode"`
}

// cfEntryMatchAttrTypes is the attr.Type map for a matcher entry object.
func cfEntryMatchAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"id":                  types.StringType,
		"key":                 types.StringType,
		"reg":                 types.StringType,
		"restrict":            types.BoolType,
		"mask":                types.BoolType,
		"ignore_cf_rule_tags": types.ListType{ElemType: types.StringType},
		"domain":              types.StringType,
		"path":                types.StringType,
		"case_insensitive":    types.BoolType,
		"active":              types.BoolType,
	}
}

// cfSectionAttrTypes is the attr.Type map for a section object.
func cfSectionAttrTypes() map[string]attr.Type {
	entryList := types.ListType{ElemType: types.ObjectType{AttrTypes: cfEntryMatchAttrTypes()}}
	return map[string]attr.Type{
		"max_count":         types.Int64Type,
		"max_length":        types.Int64Type,
		"enable_max_count":  types.BoolType,
		"enable_max_length": types.BoolType,
		"names":             entryList,
		"regex":             entryList,
		"text":              entryList,
	}
}

// cfURLSectionAttrTypes is the attr.Type map for the url section (no names).
func cfURLSectionAttrTypes() map[string]attr.Type {
	entryList := types.ListType{ElemType: types.ObjectType{AttrTypes: cfEntryMatchAttrTypes()}}
	return map[string]attr.Type{
		"max_count":         types.Int64Type,
		"max_length":        types.Int64Type,
		"enable_max_count":  types.BoolType,
		"enable_max_length": types.BoolType,
		"regex":             entryList,
		"text":              entryList,
	}
}

// cfDecodingAttrTypes is the attr.Type map for the decoding object.
func cfDecodingAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"base64":  types.BoolType,
		"dual":    types.BoolType,
		"html":    types.BoolType,
		"unicode": types.BoolType,
	}
}

// cfEntryMatchSchemaAttrs returns the schema attributes for a matcher entry.
func cfEntryMatchSchemaAttrs() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "Unique identifier for the entry, generated by the provider.",
			Optional:    true,
			Computed:    true,
		},
		"key": schema.StringAttribute{
			Description: "Exact name to match.",
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(""),
		},
		"reg": schema.StringAttribute{
			Description: "Regular expression to match.",
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(""),
		},
		"restrict": schema.BoolAttribute{
			Description: "Whether the matched entry is restricted.",
			Optional:    true,
			Computed:    true,
			Default:     booldefault.StaticBool(false),
		},
		"mask": schema.BoolAttribute{
			Description: "Whether to mask the matched value.",
			Optional:    true,
			Computed:    true,
			Default:     booldefault.StaticBool(false),
		},
		"ignore_cf_rule_tags": schema.ListAttribute{
			Description: "Content filter rule tags to exclude.",
			Optional:    true,
			ElementType: types.StringType,
		},
		"domain": schema.StringAttribute{
			Description: "Domain the entry applies to.",
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(""),
		},
		"path": schema.StringAttribute{
			Description: "Path the entry applies to.",
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString(""),
		},
		"case_insensitive": schema.BoolAttribute{
			Description: "Whether matching is case insensitive.",
			Optional:    true,
			Computed:    true,
			Default:     booldefault.StaticBool(false),
		},
		"active": schema.BoolAttribute{
			Description: "Whether the entry is active.",
			Optional:    true,
			Computed:    true,
			Default:     booldefault.StaticBool(true),
		},
	}
}

// cfSectionSchema returns a SingleNestedAttribute describing a profile section.
func cfSectionSchema(description string, required bool) schema.SingleNestedAttribute {
	entryList := schema.ListNestedAttribute{
		Optional: true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: cfEntryMatchSchemaAttrs(),
		},
	}
	attrs := map[string]schema.Attribute{
		"max_count": schema.Int64Attribute{
			Description: "Maximum number of items of this section type allowed.",
			Optional:    true,
			Computed:    true,
			Default:     int64default.StaticInt64(1),
			Validators: []validator.Int64{
				int64validator.AtLeast(1),
			},
		},
		"max_length": schema.Int64Attribute{
			Description: "Maximum number of characters per item.",
			Optional:    true,
			Computed:    true,
			Default:     int64default.StaticInt64(1024),
			Validators: []validator.Int64{
				int64validator.AtLeast(1),
			},
		},
		"enable_max_count": schema.BoolAttribute{
			Description: "Enable max-count enforcement.",
			Optional:    true,
			Computed:    true,
			Default:     booldefault.StaticBool(false),
		},
		"enable_max_length": schema.BoolAttribute{
			Description: "Enable max-length enforcement.",
			Optional:    true,
			Computed:    true,
			Default:     booldefault.StaticBool(false),
		},
		"names": entryList,
		"regex": entryList,
		"text":  entryList,
	}

	out := schema.SingleNestedAttribute{
		Description: description,
		Attributes:  attrs,
	}
	if required {
		out.Required = true
	} else {
		out.Optional = true
		out.Computed = true
	}
	return out
}

// cfURLSectionSchema returns a SingleNestedAttribute for the url section (no names).
func cfURLSectionSchema(description string) schema.SingleNestedAttribute {
	entryList := schema.ListNestedAttribute{
		Optional: true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: cfEntryMatchSchemaAttrs(),
		},
	}
	return schema.SingleNestedAttribute{
		Optional:    true,
		Computed:    true,
		Description: description,
		Attributes: map[string]schema.Attribute{
			"max_count": schema.Int64Attribute{
				Description: "Maximum number of items of this section type allowed.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(1),
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"max_length": schema.Int64Attribute{
				Description: "Maximum number of characters per item.",
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(1024),
				Validators: []validator.Int64{
					int64validator.AtLeast(1),
				},
			},
			"enable_max_count": schema.BoolAttribute{
				Description: "Enable max-count enforcement.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"enable_max_length": schema.BoolAttribute{
				Description: "Enable max-length enforcement.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"regex": entryList,
			"text":  entryList,
		},
	}
}

// NewContentFilterProfileResource creates a new content filter profile resource instance.
func NewContentFilterProfileResource() resource.Resource {
	return &ContentFilterProfileResource{}
}

// Metadata returns the resource type name.
func (r *ContentFilterProfileResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_content_filter_profile"
}

// Schema defines the schema for the content filter profile resource.
func (r *ContentFilterProfileResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Content Filter Profile in Link11 WAAP.",
		Attributes: map[string]schema.Attribute{
			"config_id": schema.StringAttribute{
				Description: "The configuration ID.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.StringAttribute{
				Description: "The unique identifier for the content filter profile.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the content filter profile.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"description": schema.StringAttribute{
				Description: "Description of the content filter profile.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"ignore_alphanum": schema.BoolAttribute{
				Description: "When true, alphanumeric-only args/headers/cookies are ignored.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"masking_seed": schema.StringAttribute{
				Description: "Seed used when masking values.",
				Required:    true,
			},
			"content_type": schema.ListAttribute{
				Description: "List of content types.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"graphql_path": schema.StringAttribute{
				Description: "JSONPath of the GraphQL property.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"ignore_body": schema.BoolAttribute{
				Description: "Whether to ignore the request body.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"active": schema.ListAttribute{
				Description: "List of active tags.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"report": schema.ListAttribute{
				Description: "List of report tags.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"ignore": schema.ListAttribute{
				Description: "List of ignore tags.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"tags": schema.ListAttribute{
				Description: "List of tags to apply.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"action": schema.StringAttribute{
				Description: "Action id or name applied by the profile.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"args":        cfSectionSchema("Arguments section.", true),
			"headers":     cfSectionSchema("Headers section.", true),
			"cookies":     cfSectionSchema("Cookies section.", true),
			"path":        cfSectionSchema("Path section.", true),
			"url":         cfURLSectionSchema("URL section."),
			"allsections": cfSectionSchema("All sections section.", false),
			"decoding": schema.SingleNestedAttribute{
				Description: "Decoding flags.",
				Required:    true,
				Attributes: map[string]schema.Attribute{
					"base64": schema.BoolAttribute{
						Description: "Enable base64 decoding.",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(true),
					},
					"dual": schema.BoolAttribute{
						Description: "Enable dual decoding.",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
					},
					"html": schema.BoolAttribute{
						Description: "Enable HTML entity decoding.",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
					},
					"unicode": schema.BoolAttribute{
						Description: "Enable unicode decoding.",
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(false),
					},
				},
			},
		},
	}
}

// Configure configures the resource with the provider client.
func (r *ContentFilterProfileResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = providerutil.ConfigureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new content filter profile resource.
func (r *ContentFilterProfileResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ContentFilterProfileResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = types.StringValue(generateIDNoDash())

	p := r.buildProfile(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.CreateContentFilterProfile(ctx, plan.ConfigID.ValueString(), plan.ID.ValueString(), p)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Content Filter Profile",
			"Could not create content filter profile: "+err.Error(),
		)
		return
	}

	r.flattenProfile(ctx, p, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read reads the content filter profile resource.
func (r *ContentFilterProfileResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ContentFilterProfileResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	p, err := r.client.GetContentFilterProfile(ctx, state.ConfigID.ValueString(), state.ID.ValueString())
	if err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Content Filter Profile",
			"Could not read content filter profile: "+err.Error(),
		)
		return
	}

	r.flattenProfile(ctx, p, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the content filter profile resource.
func (r *ContentFilterProfileResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ContentFilterProfileResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	p := r.buildProfile(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.UpdateContentFilterProfile(ctx, plan.ConfigID.ValueString(), plan.ID.ValueString(), p)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Content Filter Profile",
			"Could not update content filter profile: "+err.Error(),
		)
		return
	}

	r.flattenProfile(ctx, p, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete deletes the content filter profile resource.
func (r *ContentFilterProfileResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ContentFilterProfileResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteContentFilterProfile(ctx, state.ConfigID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Content Filter Profile",
			"Could not delete content filter profile: "+err.Error(),
		)
		return
	}
}

// ImportState imports an existing content filter profile resource.
func (r *ContentFilterProfileResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in format 'config_id/content_filter_profile_id', got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("config_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

// buildProfile maps the plan model into the API client struct.
func (r *ContentFilterProfileResource) buildProfile(ctx context.Context, plan *ContentFilterProfileResourceModel, diags *diag.Diagnostics) *client.ContentFilterProfile {
	p := &client.ContentFilterProfile{
		ID:             plan.ID.ValueString(),
		Name:           plan.Name.ValueString(),
		Description:    plan.Description.ValueString(),
		IgnoreAlphanum: plan.IgnoreAlphanum.ValueBool(),
		MaskingSeed:    plan.MaskingSeed.ValueString(),
		GraphqlPath:    plan.GraphqlPath.ValueString(),
		IgnoreBody:     plan.IgnoreBody.ValueBool(),
		Action:         plan.Action.ValueString(),
	}

	buildStringList(ctx, plan.ContentType, &p.ContentType, diags)
	buildStringList(ctx, plan.Active, &p.Active, diags)
	buildStringList(ctx, plan.Report, &p.Report, diags)
	buildStringList(ctx, plan.Ignore, &p.Ignore, diags)
	buildStringList(ctx, plan.Tags, &p.Tags, diags)

	p.Args = buildSection(ctx, plan.Args, false, diags)
	p.Headers = buildSection(ctx, plan.Headers, false, diags)
	p.Cookies = buildSection(ctx, plan.Cookies, false, diags)
	p.Path = buildSection(ctx, plan.Path, false, diags)
	p.URL = buildURLSection(ctx, plan.URL, diags)
	p.AllSections = buildSection(ctx, plan.AllSections, false, diags)
	p.Decoding = buildDecoding(ctx, plan.Decoding, diags)

	return p
}

// buildStringList copies a types.List of strings into dst when set.
func buildStringList(ctx context.Context, list types.List, dst *[]string, diags *diag.Diagnostics) {
	if !list.IsNull() && !list.IsUnknown() {
		diags.Append(list.ElementsAs(ctx, dst, false)...)
	}
}

// buildSection converts a section object into the client struct.
// useText additionally populates the "text" entry list (url section); all sections always include "names".
func buildSection(ctx context.Context, obj types.Object, useText bool, diags *diag.Diagnostics) client.ContentFilterProfileSection {
	if obj.IsNull() || obj.IsUnknown() {
		section := client.ContentFilterProfileSection{
			MaxCount:  1,
			MaxLength: 1024,
			Names:     []client.ContentFilterEntryMatch{},
			Regex:     []client.ContentFilterEntryMatch{},
		}
		if useText {
			section.Text = []client.ContentFilterEntryMatch{}
		}
		return section
	}

	var sm cfSectionModel
	diags.Append(obj.As(ctx, &sm, basetypes.ObjectAsOptions{})...)

	section := client.ContentFilterProfileSection{
		MaxCount:        int(sm.MaxCount.ValueInt64()),
		MaxLength:       int(sm.MaxLength.ValueInt64()),
		EnableMaxCount:  sm.EnableMaxCount.ValueBool(),
		EnableMaxLength: sm.EnableMaxLength.ValueBool(),
	}

	if names := buildEntryMatches(ctx, sm.Names, "names", diags); names != nil {
		section.Names = names
	} else {
		section.Names = []client.ContentFilterEntryMatch{}
	}

	if useText {
		if text := buildEntryMatches(ctx, sm.Text, "text", diags); text != nil {
			section.Text = text
		} else {
			section.Text = []client.ContentFilterEntryMatch{}
		}
	}

	if regex := buildEntryMatches(ctx, sm.Regex, "regex", diags); regex != nil {
		section.Regex = regex
	} else {
		section.Regex = []client.ContentFilterEntryMatch{}
	}

	return section
}

// buildURLSection converts the url section object (no names) into the client struct.
func buildURLSection(ctx context.Context, obj types.Object, diags *diag.Diagnostics) client.ContentFilterURLSection {
	if obj.IsNull() || obj.IsUnknown() {
		return client.ContentFilterURLSection{
			MaxCount:  1,
			MaxLength: 1024,
			Regex:     []client.ContentFilterEntryMatch{},
			Text:      []client.ContentFilterEntryMatch{},
		}
	}

	var sm cfURLSectionModel
	diags.Append(obj.As(ctx, &sm, basetypes.ObjectAsOptions{})...)

	section := client.ContentFilterURLSection{
		MaxCount:        int(sm.MaxCount.ValueInt64()),
		MaxLength:       int(sm.MaxLength.ValueInt64()),
		EnableMaxCount:  sm.EnableMaxCount.ValueBool(),
		EnableMaxLength: sm.EnableMaxLength.ValueBool(),
	}

	if text := buildEntryMatches(ctx, sm.Text, "text", diags); text != nil {
		section.Text = text
	} else {
		section.Text = []client.ContentFilterEntryMatch{}
	}

	if regex := buildEntryMatches(ctx, sm.Regex, "regex", diags); regex != nil {
		section.Regex = regex
	} else {
		section.Regex = []client.ContentFilterEntryMatch{}
	}

	return section
}

// buildEntryMatches converts a list of matcher objects into the client slice.
func buildEntryMatches(ctx context.Context, list types.List, matchType string, diags *diag.Diagnostics) []client.ContentFilterEntryMatch {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}

	var models []cfEntryMatchModel
	diags.Append(list.ElementsAs(ctx, &models, false)...)
	if diags.HasError() {
		return nil
	}

	out := make([]client.ContentFilterEntryMatch, len(models))
	for i, m := range models {
		id := m.ID.ValueString()
		if id == "" {
			id = generateIDNoDash()
		}
		out[i] = client.ContentFilterEntryMatch{
			ID:              id,
			Type:            matchType,
			Key:             m.Key.ValueString(),
			Reg:             m.Reg.ValueString(),
			Restrict:        m.Restrict.ValueBool(),
			Mask:            m.Mask.ValueBool(),
			Domain:          m.Domain.ValueString(),
			Path:            m.Path.ValueString(),
			CaseInsensitive: m.CaseInsensitive.ValueBool(),
			Active:          m.Active.ValueBool(),
		}
		buildStringList(ctx, m.IgnoreCFRuleTags, &out[i].IgnoreCFRuleTags, diags)
	}

	return out
}

// buildDecoding converts the decoding object into the client struct.
func buildDecoding(ctx context.Context, obj types.Object, diags *diag.Diagnostics) client.ContentFilterDecoding {
	var dec client.ContentFilterDecoding
	if obj.IsNull() || obj.IsUnknown() {
		return dec
	}

	var dm cfDecodingModel
	diags.Append(obj.As(ctx, &dm, basetypes.ObjectAsOptions{})...)

	dec.Base64 = dm.Base64.ValueBool()
	dec.Dual = dm.Dual.ValueBool()
	dec.HTML = dm.HTML.ValueBool()
	dec.Unicode = dm.Unicode.ValueBool()

	return dec
}

// flattenProfile maps the API client struct back into the resource state model.
func (r *ContentFilterProfileResource) flattenProfile(ctx context.Context, p *client.ContentFilterProfile, state *ContentFilterProfileResourceModel, diags *diag.Diagnostics) {
	state.ID = types.StringValue(p.ID)
	state.Name = types.StringValue(p.Name)
	state.Description = types.StringValue(p.Description)
	state.IgnoreAlphanum = types.BoolValue(p.IgnoreAlphanum)
	state.MaskingSeed = types.StringValue(p.MaskingSeed)
	state.GraphqlPath = types.StringValue(p.GraphqlPath)
	state.IgnoreBody = types.BoolValue(p.IgnoreBody)
	state.Action = types.StringValue(p.Action)

	state.ContentType = flattenStringList(ctx, p.ContentType, diags)
	state.Active = flattenStringList(ctx, p.Active, diags)
	state.Report = flattenStringList(ctx, p.Report, diags)
	state.Ignore = flattenStringList(ctx, p.Ignore, diags)
	state.Tags = flattenStringList(ctx, p.Tags, diags)

	state.Args = flattenSection(ctx, p.Args, diags)
	state.Headers = flattenSection(ctx, p.Headers, diags)
	state.Cookies = flattenSection(ctx, p.Cookies, diags)
	state.Path = flattenSection(ctx, p.Path, diags)
	state.URL = flattenURLSection(ctx, p.URL, diags)
	state.AllSections = flattenSection(ctx, p.AllSections, diags)
	state.Decoding = flattenDecoding(p.Decoding)
}

// flattenStringList builds a Terraform list value, using null for empty slices.
func flattenStringList(ctx context.Context, in []string, diags *diag.Diagnostics) types.List {
	if len(in) == 0 {
		return types.ListNull(types.StringType)
	}
	lv, d := types.ListValueFrom(ctx, types.StringType, in)
	diags.Append(d...)
	return lv
}

// flattenSection builds the Terraform object value for a section.
func flattenSection(ctx context.Context, section client.ContentFilterProfileSection, diags *diag.Diagnostics) types.Object {
	obj, d := types.ObjectValue(cfSectionAttrTypes(), map[string]attr.Value{
		"max_count":         types.Int64Value(int64(section.MaxCount)),
		"max_length":        types.Int64Value(int64(section.MaxLength)),
		"enable_max_count":  types.BoolValue(section.EnableMaxCount),
		"enable_max_length": types.BoolValue(section.EnableMaxLength),
		"names":             flattenEntryMatches(ctx, section.Names, diags),
		"regex":             flattenEntryMatches(ctx, section.Regex, diags),
		"text":              flattenEntryMatches(ctx, section.Text, diags),
	})
	diags.Append(d...)
	return obj
}

// flattenURLSection builds the Terraform object value for the url section (no names).
func flattenURLSection(ctx context.Context, section client.ContentFilterURLSection, diags *diag.Diagnostics) types.Object {
	obj, d := types.ObjectValue(cfURLSectionAttrTypes(), map[string]attr.Value{
		"max_count":         types.Int64Value(int64(section.MaxCount)),
		"max_length":        types.Int64Value(int64(section.MaxLength)),
		"enable_max_count":  types.BoolValue(section.EnableMaxCount),
		"enable_max_length": types.BoolValue(section.EnableMaxLength),
		"regex":             flattenEntryMatches(ctx, section.Regex, diags),
		"text":              flattenEntryMatches(ctx, section.Text, diags),
	})
	diags.Append(d...)
	return obj
}

// flattenEntryMatches builds the Terraform list value for matcher entries.
func flattenEntryMatches(ctx context.Context, in []client.ContentFilterEntryMatch, diags *diag.Diagnostics) types.List {
	objType := types.ObjectType{AttrTypes: cfEntryMatchAttrTypes()}
	if len(in) == 0 {
		return types.ListNull(objType)
	}

	models := make([]cfEntryMatchModel, len(in))
	for i, m := range in {
		models[i] = cfEntryMatchModel{
			ID:               types.StringValue(m.ID),
			Key:              types.StringValue(m.Key),
			Reg:              types.StringValue(m.Reg),
			Restrict:         types.BoolValue(m.Restrict),
			Mask:             types.BoolValue(m.Mask),
			IgnoreCFRuleTags: flattenStringList(ctx, m.IgnoreCFRuleTags, diags),
			Domain:           types.StringValue(m.Domain),
			Path:             types.StringValue(m.Path),
			CaseInsensitive:  types.BoolValue(m.CaseInsensitive),
			Active:           types.BoolValue(m.Active),
		}
	}

	lv, d := types.ListValueFrom(ctx, objType, models)
	diags.Append(d...)
	return lv
}

// flattenDecoding builds the Terraform object value for decoding flags.
func flattenDecoding(dec client.ContentFilterDecoding) types.Object {
	return types.ObjectValueMust(cfDecodingAttrTypes(), map[string]attr.Value{
		"base64":  types.BoolValue(dec.Base64),
		"dual":    types.BoolValue(dec.Dual),
		"html":    types.BoolValue(dec.HTML),
		"unicode": types.BoolValue(dec.Unicode),
	})
}
