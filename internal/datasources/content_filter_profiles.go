package datasources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/link11/terraform-provider-link11waap/internal/client"
	"github.com/link11/terraform-provider-link11waap/internal/providerutil"
)

var _ datasource.DataSource = &ContentFilterProfilesDataSource{}

// ContentFilterProfilesDataSource implements the data source for listing content filter profiles.
type ContentFilterProfilesDataSource struct {
	client *client.Client
}

// ContentFilterProfilesDataSourceModel describes the data source model for content filter profiles.
type ContentFilterProfilesDataSourceModel struct {
	ConfigID              types.String                    `tfsdk:"config_id"`
	ID                    types.String                    `tfsdk:"id"`
	Name                  types.String                    `tfsdk:"name"`
	ContentFilterProfiles []ContentFilterProfileDataModel `tfsdk:"content_filter_profiles"`
}

// ContentFilterProfileDataModel describes the data model for a single content filter profile.
type ContentFilterProfileDataModel struct {
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

// dsCfEntryMatchAttrTypes returns the attribute type map for a parameter-style (Type A) matcher entry object.
func dsCfEntryMatchAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"parameter":           types.StringType,
		"value":               types.StringType,
		"restrict":            types.BoolType,
		"mask":                types.BoolType,
		"ignore_cf_rule_tags": types.ListType{ElemType: types.StringType},
		"case_insensitive":    types.BoolType,
		"active":              types.BoolType,
	}
}

// dsCfEntryMatchURLPathAttrTypes returns the attribute type map for a url/path-style (Type B) matcher entry object.
func dsCfEntryMatchURLPathAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"restrict":            types.BoolType,
		"mask":                types.BoolType,
		"ignore_cf_rule_tags": types.ListType{ElemType: types.StringType},
		"domain":              types.StringType,
		"path":                types.StringType,
		"case_insensitive":    types.BoolType,
		"active":              types.BoolType,
	}
}

// dsCfSectionAttrTypes returns the attribute type map for a section object.
func dsCfSectionAttrTypes() map[string]attr.Type {
	entryList := types.ListType{ElemType: types.ObjectType{AttrTypes: dsCfEntryMatchAttrTypes()}}
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

// dsCfURLSectionAttrTypes returns the attribute type map for the url section (no names).
func dsCfURLSectionAttrTypes() map[string]attr.Type {
	entryList := types.ListType{ElemType: types.ObjectType{AttrTypes: dsCfEntryMatchURLPathAttrTypes()}}
	return map[string]attr.Type{
		"max_count":         types.Int64Type,
		"max_length":        types.Int64Type,
		"enable_max_count":  types.BoolType,
		"enable_max_length": types.BoolType,
		"regex":             entryList,
		"text":              entryList,
	}
}

// dsCfPathSectionAttrTypes returns the attribute type map for the path section (url/path-style entries).
func dsCfPathSectionAttrTypes() map[string]attr.Type {
	entryList := types.ListType{ElemType: types.ObjectType{AttrTypes: dsCfEntryMatchURLPathAttrTypes()}}
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

// dsCfDecodingAttrTypes returns the attribute type map for the decoding object.
func dsCfDecodingAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"base64":  types.BoolType,
		"dual":    types.BoolType,
		"html":    types.BoolType,
		"unicode": types.BoolType,
	}
}

// NewContentFilterProfilesDataSource returns a new instance of the content filter profiles data source.
func NewContentFilterProfilesDataSource() datasource.DataSource {
	return &ContentFilterProfilesDataSource{}
}

// Metadata returns the data source type name.
func (d *ContentFilterProfilesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_content_filter_profiles"
}

// dsCfSectionSchema returns the computed SingleNestedAttribute for a section.
func dsCfSectionSchema(description string) schema.SingleNestedAttribute {
	entryList := schema.ListNestedAttribute{
		Computed: true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: map[string]schema.Attribute{
				"parameter":           schema.StringAttribute{Computed: true, Description: "Exact name to match."},
				"value":               schema.StringAttribute{Computed: true, Description: "Regular expression to match."},
				"restrict":            schema.BoolAttribute{Computed: true, Description: "Whether the matched entry is restricted."},
				"mask":                schema.BoolAttribute{Computed: true, Description: "Whether to mask the matched value."},
				"ignore_cf_rule_tags": schema.ListAttribute{Computed: true, ElementType: types.StringType, Description: "Content filter rule tags to exclude."},
				"case_insensitive":    schema.BoolAttribute{Computed: true, Description: "Whether matching is case insensitive."},
				"active":              schema.BoolAttribute{Computed: true, Description: "Whether the entry is active."},
			},
		},
	}
	return schema.SingleNestedAttribute{
		Computed:    true,
		Description: description,
		Attributes: map[string]schema.Attribute{
			"max_count":         schema.Int64Attribute{Computed: true, Description: "Maximum number of items of this section type allowed."},
			"max_length":        schema.Int64Attribute{Computed: true, Description: "Maximum number of characters per item."},
			"enable_max_count":  schema.BoolAttribute{Computed: true, Description: "Enable max-count enforcement."},
			"enable_max_length": schema.BoolAttribute{Computed: true, Description: "Enable max-length enforcement."},
			"names":             entryList,
			"regex":             entryList,
			"text":              entryList,
		},
	}
}

// dsCfURLPathEntrySchemaAttrs returns the computed schema attributes for a url/path-style (Type B) matcher entry.
func dsCfURLPathEntrySchemaAttrs() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"restrict":            schema.BoolAttribute{Computed: true, Description: "Whether the matched entry is restricted."},
		"mask":                schema.BoolAttribute{Computed: true, Description: "Whether to mask the matched value."},
		"ignore_cf_rule_tags": schema.ListAttribute{Computed: true, ElementType: types.StringType, Description: "Content filter rule tags to exclude."},
		"domain":              schema.StringAttribute{Computed: true, Description: "Domain the entry applies to."},
		"path":                schema.StringAttribute{Computed: true, Description: "Path the entry applies to."},
		"case_insensitive":    schema.BoolAttribute{Computed: true, Description: "Whether matching is case insensitive."},
		"active":              schema.BoolAttribute{Computed: true, Description: "Whether the entry is active."},
	}
}

// dsCfURLSectionSchema returns the computed SingleNestedAttribute for the url section (no names).
func dsCfURLSectionSchema(description string) schema.SingleNestedAttribute {
	entryList := schema.ListNestedAttribute{
		Computed: true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: dsCfURLPathEntrySchemaAttrs(),
		},
	}
	return schema.SingleNestedAttribute{
		Computed:    true,
		Description: description,
		Attributes: map[string]schema.Attribute{
			"max_count":         schema.Int64Attribute{Computed: true, Description: "Maximum number of items of this section type allowed."},
			"max_length":        schema.Int64Attribute{Computed: true, Description: "Maximum number of characters per item."},
			"enable_max_count":  schema.BoolAttribute{Computed: true, Description: "Enable max-count enforcement."},
			"enable_max_length": schema.BoolAttribute{Computed: true, Description: "Enable max-length enforcement."},
			"regex":             entryList,
			"text":              entryList,
		},
	}
}

// dsCfPathSectionSchema returns the computed SingleNestedAttribute for the path section (url/path-style entries).
func dsCfPathSectionSchema(description string) schema.SingleNestedAttribute {
	entryList := schema.ListNestedAttribute{
		Computed: true,
		NestedObject: schema.NestedAttributeObject{
			Attributes: dsCfURLPathEntrySchemaAttrs(),
		},
	}
	return schema.SingleNestedAttribute{
		Computed:    true,
		Description: description,
		Attributes: map[string]schema.Attribute{
			"max_count":         schema.Int64Attribute{Computed: true, Description: "Maximum number of items of this section type allowed."},
			"max_length":        schema.Int64Attribute{Computed: true, Description: "Maximum number of characters per item."},
			"enable_max_count":  schema.BoolAttribute{Computed: true, Description: "Enable max-count enforcement."},
			"enable_max_length": schema.BoolAttribute{Computed: true, Description: "Enable max-length enforcement."},
			"names":             entryList,
			"regex":             entryList,
			"text":              entryList,
		},
	}
}

// Schema defines the schema for the content filter profiles data source.
func (d *ContentFilterProfilesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists content filter profiles in a configuration.",
		Attributes: map[string]schema.Attribute{
			"config_id": schema.StringAttribute{
				Description: "Configuration ID.",
				Required:    true,
			},
			"id": schema.StringAttribute{
				Description: "Profile ID. If specified, only the profile with this ID will be returned.",
				Optional:    true,
			},
			"name": schema.StringAttribute{
				Description: "Profile name. If specified, only the profile with this name will be returned.",
				Optional:    true,
			},
			"content_filter_profiles": schema.ListNestedAttribute{
				Description: "List of content filter profiles.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":              schema.StringAttribute{Computed: true, Description: "Unique identifier."},
						"name":            schema.StringAttribute{Computed: true, Description: "Name of the profile."},
						"description":     schema.StringAttribute{Computed: true, Description: "Description."},
						"ignore_alphanum": schema.BoolAttribute{Computed: true, Description: "Whether alphanumeric-only entries are ignored."},
						"masking_seed":    schema.StringAttribute{Computed: true, Description: "Seed used when masking values."},
						"content_type":    schema.ListAttribute{Computed: true, ElementType: types.StringType, Description: "List of content types."},
						"graphql_path":    schema.StringAttribute{Computed: true, Description: "JSONPath of the GraphQL property."},
						"ignore_body":     schema.BoolAttribute{Computed: true, Description: "Whether to ignore the request body."},
						"active":          schema.ListAttribute{Computed: true, ElementType: types.StringType, Description: "List of active tags."},
						"report":          schema.ListAttribute{Computed: true, ElementType: types.StringType, Description: "List of report tags."},
						"ignore":          schema.ListAttribute{Computed: true, ElementType: types.StringType, Description: "List of ignore tags."},
						"tags":            schema.ListAttribute{Computed: true, ElementType: types.StringType, Description: "List of tags."},
						"action":          schema.StringAttribute{Computed: true, Description: "Action id or name applied by the profile."},
						"args":            dsCfSectionSchema("Arguments section."),
						"headers":         dsCfSectionSchema("Headers section."),
						"cookies":         dsCfSectionSchema("Cookies section."),
						"path":            dsCfPathSectionSchema("Path section."),
						"url":             dsCfURLSectionSchema("URL section."),
						"allsections":     dsCfSectionSchema("All sections section."),
						"decoding": schema.SingleNestedAttribute{
							Computed:    true,
							Description: "Decoding flags.",
							Attributes: map[string]schema.Attribute{
								"base64":  schema.BoolAttribute{Computed: true, Description: "Enable base64 decoding."},
								"dual":    schema.BoolAttribute{Computed: true, Description: "Enable dual decoding."},
								"html":    schema.BoolAttribute{Computed: true, Description: "Enable HTML entity decoding."},
								"unicode": schema.BoolAttribute{Computed: true, Description: "Enable unicode decoding."},
							},
						},
					},
				},
			},
		},
	}
}

// Configure initializes the data source with the provider client.
func (d *ContentFilterProfilesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = providerutil.ConfigureClient(req.ProviderData, &resp.Diagnostics)
}

// Read retrieves the content filter profiles for the specified configuration.
func (d *ContentFilterProfilesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ContentFilterProfilesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var profiles []client.ContentFilterProfile

	if !data.ID.IsNull() && !data.ID.IsUnknown() {
		p, err := d.client.GetContentFilterProfile(ctx, data.ConfigID.ValueString(), data.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading Content Filter Profile",
				"Could not read content filter profile with ID "+data.ID.ValueString()+": "+err.Error(),
			)
			return
		}
		profiles = []client.ContentFilterProfile{*p}
	} else if !data.Name.IsNull() && !data.Name.IsUnknown() {
		allProfiles, err := d.client.ListContentFilterProfiles(ctx, data.ConfigID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading Content Filter Profiles",
				"Could not read content filter profiles: "+err.Error(),
			)
			return
		}
		for _, p := range allProfiles {
			if p.Name == data.Name.ValueString() {
				profiles = []client.ContentFilterProfile{p}
				break
			}
		}
		if len(profiles) == 0 {
			resp.Diagnostics.AddError(
				"Content Filter Profile Not Found",
				"No content filter profile found with name: "+data.Name.ValueString(),
			)
			return
		}
	} else {
		var err error
		profiles, err = d.client.ListContentFilterProfiles(ctx, data.ConfigID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading Content Filter Profiles",
				"Could not read content filter profiles: "+err.Error(),
			)
			return
		}
	}

	data.ContentFilterProfiles = make([]ContentFilterProfileDataModel, len(profiles))
	for i := range profiles {
		p := profiles[i]

		args, d1 := dsFlattenCfSection(ctx, p.Args)
		resp.Diagnostics.Append(d1...)
		headers, d2 := dsFlattenCfSection(ctx, p.Headers)
		resp.Diagnostics.Append(d2...)
		cookies, d3 := dsFlattenCfSection(ctx, p.Cookies)
		resp.Diagnostics.Append(d3...)
		pathSec, d4 := dsFlattenCfPathSection(ctx, p.Path)
		resp.Diagnostics.Append(d4...)
		urlSec, d5 := dsFlattenCfURLSection(ctx, p.URL)
		resp.Diagnostics.Append(d5...)
		allSec, d6 := dsFlattenCfSection(ctx, p.AllSections)
		resp.Diagnostics.Append(d6...)

		data.ContentFilterProfiles[i] = ContentFilterProfileDataModel{
			ID:             types.StringValue(p.ID),
			Name:           types.StringValue(p.Name),
			Description:    types.StringValue(p.Description),
			IgnoreAlphanum: types.BoolValue(p.IgnoreAlphanum),
			MaskingSeed:    types.StringValue(p.MaskingSeed),
			ContentType:    dsFlattenStringList(ctx, p.ContentType, &resp.Diagnostics),
			GraphqlPath:    types.StringValue(p.GraphqlPath),
			IgnoreBody:     types.BoolValue(p.IgnoreBody),
			Active:         dsFlattenStringList(ctx, p.Active, &resp.Diagnostics),
			Report:         dsFlattenStringList(ctx, p.Report, &resp.Diagnostics),
			Ignore:         dsFlattenStringList(ctx, p.Ignore, &resp.Diagnostics),
			Tags:           dsFlattenStringList(ctx, p.Tags, &resp.Diagnostics),
			Action:         types.StringValue(p.Action),
			Args:           args,
			Headers:        headers,
			Cookies:        cookies,
			Path:           pathSec,
			URL:            urlSec,
			AllSections:    allSec,
			Decoding:       dsFlattenCfDecoding(p.Decoding),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// dsFlattenStringList builds a Terraform list value, using null for empty slices.
func dsFlattenStringList(ctx context.Context, in []string, diags *diag.Diagnostics) types.List {
	if len(in) == 0 {
		return types.ListNull(types.StringType)
	}
	lv, d := types.ListValueFrom(ctx, types.StringType, in)
	diags.Append(d...)
	return lv
}

// dsFlattenCfSection converts a client section into a Terraform object value.
func dsFlattenCfSection(ctx context.Context, section client.ContentFilterProfileSection) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	names, dn := dsFlattenCfEntryMatchesTypeA(ctx, section.Names)
	diags.Append(dn...)
	regex, dr := dsFlattenCfEntryMatchesTypeA(ctx, section.Regex)
	diags.Append(dr...)
	text, dt := dsFlattenCfEntryMatchesTypeA(ctx, section.Text)
	diags.Append(dt...)

	obj, d := types.ObjectValue(dsCfSectionAttrTypes(), map[string]attr.Value{
		"max_count":         types.Int64Value(int64(section.MaxCount)),
		"max_length":        types.Int64Value(int64(section.MaxLength)),
		"enable_max_count":  types.BoolValue(section.EnableMaxCount),
		"enable_max_length": types.BoolValue(section.EnableMaxLength),
		"names":             names,
		"regex":             regex,
		"text":              text,
	})
	diags.Append(d...)
	return obj, diags
}

// dsFlattenCfPathSection converts the path section (url/path-style entries) into a Terraform object value.
func dsFlattenCfPathSection(ctx context.Context, section client.ContentFilterProfileSection) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	names, dn := dsFlattenCfEntryMatchesTypeB(ctx, section.Names)
	diags.Append(dn...)
	regex, dr := dsFlattenCfEntryMatchesTypeB(ctx, section.Regex)
	diags.Append(dr...)
	text, dt := dsFlattenCfEntryMatchesTypeB(ctx, section.Text)
	diags.Append(dt...)

	obj, d := types.ObjectValue(dsCfPathSectionAttrTypes(), map[string]attr.Value{
		"max_count":         types.Int64Value(int64(section.MaxCount)),
		"max_length":        types.Int64Value(int64(section.MaxLength)),
		"enable_max_count":  types.BoolValue(section.EnableMaxCount),
		"enable_max_length": types.BoolValue(section.EnableMaxLength),
		"names":             names,
		"regex":             regex,
		"text":              text,
	})
	diags.Append(d...)
	return obj, diags
}

// dsFlattenCfURLSection converts the url section (no names) into a Terraform object value.
func dsFlattenCfURLSection(ctx context.Context, section client.ContentFilterURLSection) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	regex, dr := dsFlattenCfEntryMatchesTypeB(ctx, section.Regex)
	diags.Append(dr...)
	text, dt := dsFlattenCfEntryMatchesTypeB(ctx, section.Text)
	diags.Append(dt...)

	obj, d := types.ObjectValue(dsCfURLSectionAttrTypes(), map[string]attr.Value{
		"max_count":         types.Int64Value(int64(section.MaxCount)),
		"max_length":        types.Int64Value(int64(section.MaxLength)),
		"enable_max_count":  types.BoolValue(section.EnableMaxCount),
		"enable_max_length": types.BoolValue(section.EnableMaxLength),
		"regex":             regex,
		"text":              text,
	})
	diags.Append(d...)
	return obj, diags
}

// dsFlattenCfEntryMatchesTypeA converts parameter-style (Type A) matcher entries into a Terraform list value.
func dsFlattenCfEntryMatchesTypeA(ctx context.Context, in []client.ContentFilterEntryMatch) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	objType := types.ObjectType{AttrTypes: dsCfEntryMatchAttrTypes()}

	if len(in) == 0 {
		return types.ListNull(objType), diags
	}

	values := make([]attr.Value, len(in))
	for i, m := range in {
		obj, d := types.ObjectValue(dsCfEntryMatchAttrTypes(), map[string]attr.Value{
			"parameter":           types.StringValue(m.Key),
			"value":               types.StringValue(m.Reg),
			"restrict":            types.BoolValue(m.Restrict),
			"mask":                types.BoolValue(m.Mask),
			"ignore_cf_rule_tags": dsFlattenStringList(ctx, m.IgnoreCFRuleTags, &diags),
			"case_insensitive":    types.BoolValue(m.CaseInsensitive),
			"active":              types.BoolValue(m.Active),
		})
		diags.Append(d...)
		values[i] = obj
	}

	lv, d := types.ListValue(objType, values)
	diags.Append(d...)
	return lv, diags
}

// dsFlattenCfEntryMatchesTypeB converts url/path-style (Type B) matcher entries into a Terraform list value.
func dsFlattenCfEntryMatchesTypeB(ctx context.Context, in []client.ContentFilterEntryMatch) (types.List, diag.Diagnostics) {
	var diags diag.Diagnostics
	objType := types.ObjectType{AttrTypes: dsCfEntryMatchURLPathAttrTypes()}

	if len(in) == 0 {
		return types.ListNull(objType), diags
	}

	values := make([]attr.Value, len(in))
	for i, m := range in {
		obj, d := types.ObjectValue(dsCfEntryMatchURLPathAttrTypes(), map[string]attr.Value{
			"restrict":            types.BoolValue(m.Restrict),
			"mask":                types.BoolValue(m.Mask),
			"ignore_cf_rule_tags": dsFlattenStringList(ctx, m.IgnoreCFRuleTags, &diags),
			"domain":              types.StringValue(m.Domain),
			"path":                types.StringValue(m.Path),
			"case_insensitive":    types.BoolValue(m.CaseInsensitive),
			"active":              types.BoolValue(m.Active),
		})
		diags.Append(d...)
		values[i] = obj
	}

	lv, d := types.ListValue(objType, values)
	diags.Append(d...)
	return lv, diags
}

// dsFlattenCfDecoding converts decoding flags into a Terraform object value.
func dsFlattenCfDecoding(dec client.ContentFilterDecoding) types.Object {
	return types.ObjectValueMust(dsCfDecodingAttrTypes(), map[string]attr.Value{
		"base64":  types.BoolValue(dec.Base64),
		"dual":    types.BoolValue(dec.Dual),
		"html":    types.BoolValue(dec.HTML),
		"unicode": types.BoolValue(dec.Unicode),
	})
}
