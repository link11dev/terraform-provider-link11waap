package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
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
	_ resource.Resource                = &ServerGroupResource{}
	_ resource.ResourceWithImportState = &ServerGroupResource{}
)

// ServerGroupResource implements the server group resource.
type ServerGroupResource struct {
	client *client.Client
}

// ServerGroupResourceModel describes the server group resource data model.
type ServerGroupResourceModel struct {
	ConfigID               types.String `tfsdk:"config_id"`
	ID                     types.String `tfsdk:"id"`
	Name                   types.String `tfsdk:"name"`
	Description            types.String `tfsdk:"description"`
	ServerNames            types.List   `tfsdk:"server_names"`
	SecurityPolicy         types.String `tfsdk:"security_policy"`
	RoutingProfile         types.String `tfsdk:"routing_profile"`
	ProxyTemplate          types.String `tfsdk:"proxy_template"`
	ChallengeCookieDomain  types.String `tfsdk:"challenge_cookie_domain"`
	SSLCertificate         types.String `tfsdk:"ssl_certificate"`
	ClientCertificate      types.String `tfsdk:"client_certificate"`
	ClientCertificateMode  types.String `tfsdk:"client_certificate_mode"`
	MobileApplicationGroup types.String `tfsdk:"mobile_application_group"`
}

// NewServerGroupResource creates a new server group resource instance.
func NewServerGroupResource() resource.Resource {
	return &ServerGroupResource{}
}

// Metadata returns the resource type name.
func (r *ServerGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server_group"
}

// Schema defines the schema for the server group resource.
func (r *ServerGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Server Group (site/domain) in Link11 WAAP.",
		Attributes: map[string]schema.Attribute{
			"config_id": schema.StringAttribute{
				Description: "The configuration ID.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.StringAttribute{
				Description: "The unique identifier for the server group.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The site name.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 250),
				},
			},
			"description": schema.StringAttribute{
				Description: "Description of the server group.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"server_names": schema.ListAttribute{
				Description: "Host names corresponding to the site.",
				Required:    true,
				ElementType: types.StringType,
			},
			"security_policy": schema.StringAttribute{
				Description: "ID of security policy applied on site.",
				Required:    true,
			},
			"routing_profile": schema.StringAttribute{
				Description: "ID of routing profile used for site.",
				Computed:    true,
				Optional:    true,
				Default:     stringdefault.StaticString("__default__"),
			},
			"proxy_template": schema.StringAttribute{
				Description: "ID of proxy template used for site.",
				Required:    true,
			},
			"challenge_cookie_domain": schema.StringAttribute{
				Description: "The domain for a challenge's cookie.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"ssl_certificate": schema.StringAttribute{
				Description: "ID of SSL certificate attached to site.",
				Optional:    true,
			},
			"client_certificate": schema.StringAttribute{
				Description: "ID of SSL client CA certificate attached to site.",
				Optional:    true,
			},
			"client_certificate_mode": schema.StringAttribute{
				Description: "Controls how client certificate is checked when mTLS is enabled. Valid values: on, off, optional.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("off"),
				Validators: []validator.String{
					stringvalidator.OneOf("on", "off", "optional", ""),
				},
			},
			"mobile_application_group": schema.StringAttribute{
				Description: "ID of Mobile Application Group used for site.",
				Optional:    true,
			},
		},
	}
}

// Configure configures the resource with the provider client.
func (r *ServerGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = providerutil.ConfigureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new server group resource.
func (r *ServerGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ServerGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = types.StringValue(generateID())

	// Convert server_names to []string
	var serverNames []string
	resp.Diagnostics.Append(plan.ServerNames.ElementsAs(ctx, &serverNames, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sgReq := &client.ServerGroupCreateRequest{
		ID:                     plan.ID.ValueString(),
		Name:                   plan.Name.ValueString(),
		Description:            plan.Description.ValueString(),
		ServerNames:            serverNames,
		SecurityPolicy:         plan.SecurityPolicy.ValueString(),
		RoutingProfile:         plan.RoutingProfile.ValueString(),
		ProxyTemplate:          plan.ProxyTemplate.ValueString(),
		ChallengeCookieDomain:  plan.ChallengeCookieDomain.ValueString(),
		SSLCertificate:         plan.SSLCertificate.ValueString(),
		ClientCertificate:      plan.ClientCertificate.ValueString(),
		ClientCertificateMode:  plan.ClientCertificateMode.ValueString(),
		MobileApplicationGroup: plan.MobileApplicationGroup.ValueString(),
	}

	err := r.client.CreateServerGroup(ctx, plan.ConfigID.ValueString(), plan.ID.ValueString(), sgReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Server Group",
			"Could not create server group: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read reads the server group resource.
func (r *ServerGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ServerGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sg, err := r.client.GetServerGroup(ctx, state.ConfigID.ValueString(), state.ID.ValueString())
	if err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Server Group",
			"Could not read server group: "+err.Error(),
		)
		return
	}

	// Update state from API response
	state.Name = types.StringValue(sg.Name)
	state.Description = types.StringValue(sg.Description)
	state.SecurityPolicy = types.StringValue(sg.SecurityPolicy)
	state.RoutingProfile = types.StringValue(sg.RoutingProfile)
	state.ProxyTemplate = types.StringValue(sg.ProxyTemplate)
	state.ChallengeCookieDomain = types.StringValue(sg.ChallengeCookieDomain)

	serverNames, diags := types.ListValueFrom(ctx, types.StringType, sg.ServerNames)
	resp.Diagnostics.Append(diags...)
	state.ServerNames = serverNames

	if sg.SSLCertificate != "" {
		state.SSLCertificate = types.StringValue(sg.SSLCertificate)
	} else {
		state.SSLCertificate = types.StringNull()
	}
	if sg.ClientCertificate != "" {
		state.ClientCertificate = types.StringValue(sg.ClientCertificate)
	} else {
		state.ClientCertificate = types.StringNull()
	}
	if sg.ClientCertificateMode != "" {
		state.ClientCertificateMode = types.StringValue(sg.ClientCertificateMode)
	}
	if sg.MobileApplicationGroup != "" {
		state.MobileApplicationGroup = types.StringValue(sg.MobileApplicationGroup)
	} else {
		state.MobileApplicationGroup = types.StringNull()
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the server group resource.
func (r *ServerGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ServerGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var serverNames []string
	resp.Diagnostics.Append(plan.ServerNames.ElementsAs(ctx, &serverNames, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	sgReq := &client.ServerGroupCreateRequest{
		ID:                     plan.ID.ValueString(),
		Name:                   plan.Name.ValueString(),
		Description:            plan.Description.ValueString(),
		ServerNames:            serverNames,
		SecurityPolicy:         plan.SecurityPolicy.ValueString(),
		RoutingProfile:         plan.RoutingProfile.ValueString(),
		ProxyTemplate:          plan.ProxyTemplate.ValueString(),
		ChallengeCookieDomain:  plan.ChallengeCookieDomain.ValueString(),
		SSLCertificate:         plan.SSLCertificate.ValueString(),
		ClientCertificate:      plan.ClientCertificate.ValueString(),
		ClientCertificateMode:  plan.ClientCertificateMode.ValueString(),
		MobileApplicationGroup: plan.MobileApplicationGroup.ValueString(),
	}

	err := r.client.UpdateServerGroup(ctx, plan.ConfigID.ValueString(), plan.ID.ValueString(), sgReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Server Group",
			"Could not update server group: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete deletes the server group resource.
func (r *ServerGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ServerGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteServerGroup(ctx, state.ConfigID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Server Group",
			"Could not delete server group: "+err.Error(),
		)
		return
	}
}

// ImportState imports an existing server group resource.
func (r *ServerGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import ID format: config_id/server_group_id
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in format 'config_id/server_group_id', got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("config_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}
