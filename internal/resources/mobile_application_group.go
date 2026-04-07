package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
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
	_ resource.Resource                = &MobileApplicationGroupResource{}
	_ resource.ResourceWithImportState = &MobileApplicationGroupResource{}
)

// MobileApplicationGroupResource manages a Mobile Application Group in Link11 WAAP.
type MobileApplicationGroupResource struct {
	client *client.Client
}

// MobileApplicationGroupResourceModel maps the Terraform schema data to
// a GO struct for easier handling in CRUD operations.
type MobileApplicationGroupResourceModel struct {
	ConfigID     types.String `tfsdk:"config_id"`
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	UIDHeader    types.String `tfsdk:"uid_header"`
	Grace        types.String `tfsdk:"grace"`
	ActiveConfig types.Set    `tfsdk:"active_config"`
	Signatures   types.Set    `tfsdk:"signatures"`
}

// ActiveConfigModel and SignatureModel are used to represent the nested blocks in the schema.
type ActiveConfigModel struct {
	Active types.Bool   `tfsdk:"active"`
	JSON   types.String `tfsdk:"json"`
	Name   types.String `tfsdk:"name"`
}

// SignatureModel represents the structure of each signature in the signatures block.
type SignatureModel struct {
	Active types.Bool   `tfsdk:"active"`
	Hash   types.String `tfsdk:"hash"`
	Name   types.String `tfsdk:"name"`
}

// NewMobileApplicationGroupResource is a helper function to simplify the creation of the resource.
func NewMobileApplicationGroupResource() resource.Resource {
	return &MobileApplicationGroupResource{}
}

// Metadata returns the resource type name.
func (r *MobileApplicationGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mobile_application_group"
}

// Schema defines the schema for the Mobile Application Group resource.
func (r *MobileApplicationGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Mobile Application Group in Link11 WAAP.",
		Attributes: map[string]schema.Attribute{
			"config_id": schema.StringAttribute{
				Description: "The configuration ID.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.StringAttribute{
				Description: "The unique identifier for the mobile application group.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the mobile application group.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the mobile application group.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"uid_header": schema.StringAttribute{
				Description: "The UID header name for device identification.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"grace": schema.StringAttribute{
				Description: "Grace period value.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
		},
		Blocks: map[string]schema.Block{
			"active_config": schema.SetNestedBlock{
				Description: "Set of active configuration entries.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"active": schema.BoolAttribute{
							Description: "Whether this configuration entry is active.",
							Required:    true,
						},
						"json": schema.StringAttribute{
							Description: "JSON configuration string.",
							Required:    true,
						},
						"name": schema.StringAttribute{
							Description: "Name of the configuration entry.",
							Required:    true,
						},
					},
				},
			},
			"signatures": schema.SetNestedBlock{
				Description: "Set of application signatures.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"active": schema.BoolAttribute{
							Description: "Whether this signature is active.",
							Required:    true,
						},
						"hash": schema.StringAttribute{
							Description: "Hash value of the signature.",
							Required:    true,
						},
						"name": schema.StringAttribute{
							Description: "Name of the signature.",
							Required:    true,
						},
					},
				},
			},
		},
	}
}

// Configure is called by the framework to set up the resource with the provider's configured client.
func (r *MobileApplicationGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = providerutil.ConfigureClient(req.ProviderData, &resp.Diagnostics)
}

// Create handles the creation of the Mobile Application Group resource.
func (r *MobileApplicationGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan MobileApplicationGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = types.StringValue(generateID())

	mag := buildMobileApplicationGroupAPIModel(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.CreateMobileApplicationGroup(ctx, plan.ConfigID.ValueString(), plan.ID.ValueString(), mag)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Mobile Application Group",
			"Could not create mobile application group: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read retrieves the current state of the Mobile Application Group resource
func (r *MobileApplicationGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state MobileApplicationGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mag, err := r.client.GetMobileApplicationGroup(ctx, state.ConfigID.ValueString(), state.ID.ValueString())
	if err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Mobile Application Group",
			"Could not read mobile application group: "+err.Error(),
		)
		return
	}

	state.Name = types.StringValue(mag.Name)
	state.Description = types.StringValue(mag.Description)
	state.UIDHeader = types.StringValue(mag.UIDHeader)
	state.Grace = types.StringValue(mag.Grace)

	// ActiveConfig
	if mag.ActiveConfig != nil {
		acModels := make([]ActiveConfigModel, len(mag.ActiveConfig))
		for i, ac := range mag.ActiveConfig {
			acModels[i] = ActiveConfigModel{
				Active: types.BoolValue(ac.Active),
				JSON:   types.StringValue(ac.JSON),
				Name:   types.StringValue(ac.Name),
			}
		}
		acSet, diags := types.SetValueFrom(ctx, types.ObjectType{
			AttrTypes: activeConfigAttrTypes(),
		}, acModels)
		resp.Diagnostics.Append(diags...)
		state.ActiveConfig = acSet
	} else {
		state.ActiveConfig = types.SetNull(types.ObjectType{AttrTypes: activeConfigAttrTypes()})
	}

	// Signatures
	if mag.Signatures != nil {
		sigModels := make([]SignatureModel, len(mag.Signatures))
		for i, sig := range mag.Signatures {
			sigModels[i] = SignatureModel{
				Active: types.BoolValue(sig.Active),
				Hash:   types.StringValue(sig.Hash),
				Name:   types.StringValue(sig.Name),
			}
		}
		sigSet, diags := types.SetValueFrom(ctx, types.ObjectType{
			AttrTypes: signatureAttrTypes(),
		}, sigModels)
		resp.Diagnostics.Append(diags...)
		state.Signatures = sigSet
	} else {
		state.Signatures = types.SetNull(types.ObjectType{AttrTypes: signatureAttrTypes()})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update handles updates to the Mobile Application Group resource.
func (r *MobileApplicationGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan MobileApplicationGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	mag := buildMobileApplicationGroupAPIModel(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.UpdateMobileApplicationGroup(ctx, plan.ConfigID.ValueString(), plan.ID.ValueString(), mag)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Mobile Application Group",
			"Could not update mobile application group: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete handles the deletion of the Mobile Application Group resource.
func (r *MobileApplicationGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state MobileApplicationGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteMobileApplicationGroup(ctx, state.ConfigID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Mobile Application Group",
			"Could not delete mobile application group: "+err.Error(),
		)
		return
	}
}

// ImportState allows importing an existing Mobile Application Group resource
// into Terraform state using the format 'config_id/mobile_application_group_id'.
func (r *MobileApplicationGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in format 'config_id/mobile_application_group_id', got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("config_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

func buildMobileApplicationGroupAPIModel(ctx context.Context, plan *MobileApplicationGroupResourceModel, diags *diag.Diagnostics) *client.MobileApplicationGroup {
	mag := &client.MobileApplicationGroup{
		ID:          plan.ID.ValueString(),
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		UIDHeader:   plan.UIDHeader.ValueString(),
		Grace:       plan.Grace.ValueString(),
	}

	// ActiveConfig
	if !plan.ActiveConfig.IsNull() && !plan.ActiveConfig.IsUnknown() {
		var acModels []ActiveConfigModel
		diags.Append(plan.ActiveConfig.ElementsAs(ctx, &acModels, false)...)

		mag.ActiveConfig = make([]client.ActiveConfig, len(acModels))
		for i, m := range acModels {
			mag.ActiveConfig[i] = client.ActiveConfig{
				Active: m.Active.ValueBool(),
				JSON:   m.JSON.ValueString(),
				Name:   m.Name.ValueString(),
			}
		}
	}

	// Signatures
	if !plan.Signatures.IsNull() && !plan.Signatures.IsUnknown() {
		var sigModels []SignatureModel
		diags.Append(plan.Signatures.ElementsAs(ctx, &sigModels, false)...)

		mag.Signatures = make([]client.Signature, len(sigModels))
		for i, m := range sigModels {
			mag.Signatures[i] = client.Signature{
				Active: m.Active.ValueBool(),
				Hash:   m.Hash.ValueString(),
				Name:   m.Name.ValueString(),
			}
		}
	}

	return mag
}

func activeConfigAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"active": types.BoolType,
		"json":   types.StringType,
		"name":   types.StringType,
	}
}

func signatureAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"active": types.BoolType,
		"hash":   types.StringType,
		"name":   types.StringType,
	}
}
