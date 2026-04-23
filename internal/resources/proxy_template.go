package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/link11/terraform-provider-link11waap/internal/client"
	"github.com/link11/terraform-provider-link11waap/internal/providerutil"
)

var (
	_ resource.Resource                = &ProxyTemplateResource{}
	_ resource.ResourceWithImportState = &ProxyTemplateResource{}
)

// ProxyTemplateResource manages a Link11 WAAP Proxy Template
type ProxyTemplateResource struct {
	client *client.Client
}

// ProxyTemplateResourceModel maps the resource schema data
type ProxyTemplateResourceModel struct {
	ConfigID                      types.String `tfsdk:"config_id"`
	ID                            types.String `tfsdk:"id"`
	Name                          types.String `tfsdk:"name"`
	Description                   types.String `tfsdk:"description"`
	ACAOHeader                    types.Bool   `tfsdk:"acao_header"`
	XFFHeaderName                 types.List   `tfsdk:"xff_header_name"`
	XRealIPHeaderName             types.String `tfsdk:"xrealip_header_name"`
	ProxyConnectTimeout           types.String `tfsdk:"proxy_connect_timeout"`
	ProxyReadTimeout              types.String `tfsdk:"proxy_read_timeout"`
	ProxySendTimeout              types.String `tfsdk:"proxy_send_timeout"`
	UpstreamHost                  types.String `tfsdk:"upstream_host"`
	ClientBodyTimeout             types.String `tfsdk:"client_body_timeout"`
	ClientBodyBufferSize          types.String `tfsdk:"client_body_buffer_size"`
	ClientHeaderTimeout           types.String `tfsdk:"client_header_timeout"`
	ClientHeaderBufferSize        types.String `tfsdk:"client_header_buffer_size"`
	ClientMaxBodySize             types.String `tfsdk:"client_max_body_size"`
	KeepaliveTimeout              types.String `tfsdk:"keepalive_timeout"`
	SendTimeout                   types.String `tfsdk:"send_timeout"`
	LimitReqRate                  types.String `tfsdk:"limit_req_rate"`
	LimitReqBurst                 types.String `tfsdk:"limit_req_burst"`
	MaskHeaders                   types.String `tfsdk:"mask_headers"`
	CustomListener                types.Bool   `tfsdk:"custom_listener"`
	LargeClientHeaderBuffersCount types.String `tfsdk:"large_client_header_buffers_count"`
	LargeClientHeaderBuffersSize  types.String `tfsdk:"large_client_header_buffers_size"`
	ConfSpecific                  types.String `tfsdk:"conf_specific"`
	SSLConfSpecific               types.String `tfsdk:"ssl_conf_specific"`
	SSLCiphers                    types.String `tfsdk:"ssl_ciphers"`
	SSLProtocols                  types.List   `tfsdk:"ssl_protocols"`
	AdvancedConfiguration         types.List   `tfsdk:"advanced_configuration"`
}

// AdvancedConfigModel represents an individual advanced configuration block
// in the resource schema
type AdvancedConfigModel struct {
	Name          types.String `tfsdk:"name"`
	Protocol      types.List   `tfsdk:"protocol"`
	Configuration types.String `tfsdk:"configuration"`
	Description   types.String `tfsdk:"description"`
}

// NewProxyTemplateResource creates a new instance of ProxyTemplateResource
func NewProxyTemplateResource() resource.Resource {
	return &ProxyTemplateResource{}
}

// Metadata returns the proxy template resource type name
func (r *ProxyTemplateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_proxy_template"
}

// Schema defines the schema for the proxy template resource
func (r *ProxyTemplateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Proxy Template in Link11 WAAP.",
		Attributes: map[string]schema.Attribute{
			"config_id": schema.StringAttribute{
				Description: "The configuration ID.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.StringAttribute{
				Description: "The unique identifier for the proxy template.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the proxy template.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "Description of the proxy template.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"acao_header": schema.BoolAttribute{
				Description: "Whether to add Access-Control-Allow-Origin header.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"xff_header_name": schema.ListAttribute{
				Description: "X-Forwarded-For header names.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Default: listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue("X-Forwarded-For"),
				})),
				Validators: []validator.List{
					listvalidator.SizeBetween(1, 5),
				},
			},
			"xrealip_header_name": schema.StringAttribute{
				Description: "X-Real-IP header name.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("X-Real-IP"),
			},
			"proxy_connect_timeout": schema.StringAttribute{
				Description: "Proxy connect timeout.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("5"),
			},
			"proxy_read_timeout": schema.StringAttribute{
				Description: "Proxy read timeout.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("60"),
			},
			"proxy_send_timeout": schema.StringAttribute{
				Description: "Proxy send timeout.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("30"),
			},
			"upstream_host": schema.StringAttribute{
				Description: "Upstream host value.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("$host"),
			},
			"client_body_timeout": schema.StringAttribute{
				Description: "Client body timeout.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("5"),
			},
			"client_body_buffer_size": schema.StringAttribute{
				Description: "Client body buffer size.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("4"),
			},
			"client_header_timeout": schema.StringAttribute{
				Description: "Client header timeout.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("5"),
			},
			"client_header_buffer_size": schema.StringAttribute{
				Description: "Client header buffer size.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("32"),
			},
			"client_max_body_size": schema.StringAttribute{
				Description: "Maximum allowed client body size.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("150"),
			},
			"keepalive_timeout": schema.StringAttribute{
				Description: "Keepalive timeout.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("660"),
			},
			"send_timeout": schema.StringAttribute{
				Description: "Send timeout.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("5"),
			},
			"limit_req_rate": schema.StringAttribute{
				Description: "Request rate limit.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("1200"),
			},
			"limit_req_burst": schema.StringAttribute{
				Description: "Request rate limit burst.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("400"),
			},
			"mask_headers": schema.StringAttribute{
				Description: "Headers to mask in responses.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"custom_listener": schema.BoolAttribute{
				Description: "Whether to use a custom listener.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"large_client_header_buffers_count": schema.StringAttribute{
				Description: "Number of large client header buffers.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("2"),
			},
			"large_client_header_buffers_size": schema.StringAttribute{
				Description: "Size of large client header buffers.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("32"),
			},
			"conf_specific": schema.StringAttribute{
				Description: "Configuration-specific nginx directives.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"ssl_conf_specific": schema.StringAttribute{
				Description: "SSL-specific nginx directives.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"ssl_ciphers": schema.StringAttribute{
				Description: "SSL cipher suite string.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:DHE-RSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-SHA384:ECDHE-ECDSA-AES256-SHA384:ECDHE-ECDSA-AES256-SHA:ECDHE-RSA-AES256-SHA:DHE-RSA-AES256-SHA256:DHE-RSA-AES256-SHA:!SHA1:!SHA256:!SHA384:!DSS:!aNULL"),
			},
			"ssl_protocols": schema.ListAttribute{
				Description: "List of SSL/TLS protocols to enable. Valid values: TLSv1.1, TLSv1.2, TLSv1.3, SSLv2, SSLv3, TLSv1.",
				Optional:    true,
				Computed:    true,
				ElementType: types.StringType,
				Default: listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{
					types.StringValue("TLSv1.2"),
					types.StringValue("TLSv1.3"),
				})),
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
					listvalidator.ValueStringsAre(
						stringvalidator.OneOf("TLSv1.1", "TLSv1.2", "TLSv1.3", "SSLv2", "SSLv3", "TLSv1"),
					),
				},
			},
			"advanced_configuration": schema.ListNestedAttribute{
				Description: "Advanced nginx configuration blocks.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "Name of the advanced configuration block.",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
							},
						},
						"protocol": schema.ListAttribute{
							Description: "Protocols for which config applies. Valid values: http, https.",
							Required:    true,
							ElementType: types.StringType,
							Validators: []validator.List{
								listvalidator.SizeAtLeast(1),
								listvalidator.ValueStringsAre(
									stringvalidator.OneOf("http", "https"),
								),
							},
						},
						"configuration": schema.StringAttribute{
							Description: "Custom nginx configuration lines.",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.LengthAtLeast(1),
							},
						},
						"description": schema.StringAttribute{
							Description: "Description of the configuration block.",
							Optional:    true,
							Computed:    true,
							Default:     stringdefault.StaticString(""),
						},
					},
				},
			},
		},
	}
}

// Configure sets up the API client for the proxy template resource using provider data
func (r *ProxyTemplateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = providerutil.ConfigureClient(req.ProviderData, &resp.Diagnostics)
}

// Create handles the creation of a new proxy template resource
func (r *ProxyTemplateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ProxyTemplateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = types.StringValue(generateID())

	pt := buildProxyTemplateAPIModel(ctx, &plan)

	err := r.client.CreateProxyTemplate(ctx, plan.ConfigID.ValueString(), plan.ID.ValueString(), pt)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Proxy Template",
			"Could not create proxy template: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read reads the proxy template resource
func (r *ProxyTemplateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ProxyTemplateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pt, err := r.client.GetProxyTemplate(ctx, state.ConfigID.ValueString(), state.ID.ValueString())
	if err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Proxy Template",
			"Could not read proxy template: "+err.Error(),
		)
		return
	}

	state.Name = types.StringValue(pt.Name)
	state.Description = types.StringValue(pt.Description)
	state.ACAOHeader = types.BoolValue(pt.ACAOHeader)
	if pt.XFFHeaderName != nil {
		xffList, diags := types.ListValueFrom(ctx, types.StringType, pt.XFFHeaderName)
		resp.Diagnostics.Append(diags...)
		state.XFFHeaderName = xffList
	} else {
		state.XFFHeaderName = types.ListNull(types.StringType)
	}
	state.XRealIPHeaderName = types.StringValue(pt.XRealIPHeaderName)
	state.ProxyConnectTimeout = types.StringValue(pt.ProxyConnectTimeout)
	state.ProxyReadTimeout = types.StringValue(pt.ProxyReadTimeout)
	state.ProxySendTimeout = types.StringValue(pt.ProxySendTimeout)
	state.UpstreamHost = types.StringValue(pt.UpstreamHost)
	state.ClientBodyTimeout = types.StringValue(pt.ClientBodyTimeout)
	state.ClientBodyBufferSize = types.StringValue(pt.ClientBodyBufferSize)
	state.ClientHeaderTimeout = types.StringValue(pt.ClientHeaderTimeout)
	state.ClientHeaderBufferSize = types.StringValue(pt.ClientHeaderBufferSize)
	state.ClientMaxBodySize = types.StringValue(pt.ClientMaxBodySize)
	state.KeepaliveTimeout = types.StringValue(pt.KeepaliveTimeout)
	state.SendTimeout = types.StringValue(pt.SendTimeout)
	state.LimitReqRate = types.StringValue(pt.LimitReqRate)
	state.LimitReqBurst = types.StringValue(pt.LimitReqBurst)
	state.MaskHeaders = types.StringValue(pt.MaskHeaders)
	state.CustomListener = types.BoolValue(pt.CustomListener)
	state.LargeClientHeaderBuffersCount = types.StringValue(pt.LargeClientHeaderBuffersCount)
	state.LargeClientHeaderBuffersSize = types.StringValue(pt.LargeClientHeaderBuffersSize)
	state.ConfSpecific = types.StringValue(pt.ConfSpecific)
	state.SSLConfSpecific = types.StringValue(pt.SSLConfSpecific)
	state.SSLCiphers = types.StringValue(pt.SSLCiphers)

	// SSL Protocols
	if pt.SSLProtocols != nil {
		sslList, diags := types.ListValueFrom(ctx, types.StringType, pt.SSLProtocols)
		resp.Diagnostics.Append(diags...)
		state.SSLProtocols = sslList
	} else {
		state.SSLProtocols = types.ListNull(types.StringType)
	}

	// Advanced Configuration
	if len(pt.AdvancedConfiguration) > 0 {
		advModels := make([]AdvancedConfigModel, len(pt.AdvancedConfiguration))
		for i, ac := range pt.AdvancedConfiguration {
			model := AdvancedConfigModel{
				Name:          types.StringValue(ac.Name),
				Configuration: types.StringValue(ac.Configuration),
				Description:   types.StringValue(ac.Description),
			}

			if ac.Protocol != nil {
				protoList, diags := types.ListValueFrom(ctx, types.StringType, ac.Protocol)
				resp.Diagnostics.Append(diags...)
				model.Protocol = protoList
			} else {
				model.Protocol = types.ListNull(types.StringType)
			}

			advModels[i] = model
		}

		advList, diags := types.ListValueFrom(ctx, types.ObjectType{
			AttrTypes: advancedConfigAttrTypes(),
		}, advModels)
		resp.Diagnostics.Append(diags...)
		state.AdvancedConfiguration = advList
	} else {
		state.AdvancedConfiguration = types.ListNull(types.ObjectType{AttrTypes: advancedConfigAttrTypes()})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the proxy template resource
func (r *ProxyTemplateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ProxyTemplateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pt := buildProxyTemplateAPIModel(ctx, &plan)

	err := r.client.UpdateProxyTemplate(ctx, plan.ConfigID.ValueString(), plan.ID.ValueString(), pt)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Proxy Template",
			"Could not update proxy template: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete deletes the proxy template resource
func (r *ProxyTemplateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ProxyTemplateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteProxyTemplate(ctx, state.ConfigID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Proxy Template",
			"Could not delete proxy template: "+err.Error(),
		)
		return
	}
}

// ImportState imports the proxy template resource state using the provided ID
func (r *ProxyTemplateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in format 'config_id/proxy_template_id', got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("config_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

func buildProxyTemplateAPIModel(ctx context.Context, plan *ProxyTemplateResourceModel) *client.ProxyTemplate {
	pt := &client.ProxyTemplate{
		ID:                            plan.ID.ValueString(),
		Name:                          plan.Name.ValueString(),
		Description:                   plan.Description.ValueString(),
		ACAOHeader:                    plan.ACAOHeader.ValueBool(),
		XRealIPHeaderName:             plan.XRealIPHeaderName.ValueString(),
		ProxyConnectTimeout:           plan.ProxyConnectTimeout.ValueString(),
		ProxyReadTimeout:              plan.ProxyReadTimeout.ValueString(),
		ProxySendTimeout:              plan.ProxySendTimeout.ValueString(),
		UpstreamHost:                  plan.UpstreamHost.ValueString(),
		ClientBodyTimeout:             plan.ClientBodyTimeout.ValueString(),
		ClientBodyBufferSize:          plan.ClientBodyBufferSize.ValueString(),
		ClientHeaderTimeout:           plan.ClientHeaderTimeout.ValueString(),
		ClientHeaderBufferSize:        plan.ClientHeaderBufferSize.ValueString(),
		ClientMaxBodySize:             plan.ClientMaxBodySize.ValueString(),
		KeepaliveTimeout:              plan.KeepaliveTimeout.ValueString(),
		SendTimeout:                   plan.SendTimeout.ValueString(),
		LimitReqRate:                  plan.LimitReqRate.ValueString(),
		LimitReqBurst:                 plan.LimitReqBurst.ValueString(),
		MaskHeaders:                   plan.MaskHeaders.ValueString(),
		CustomListener:                plan.CustomListener.ValueBool(),
		LargeClientHeaderBuffersCount: plan.LargeClientHeaderBuffersCount.ValueString(),
		LargeClientHeaderBuffersSize:  plan.LargeClientHeaderBuffersSize.ValueString(),
		ConfSpecific:                  plan.ConfSpecific.ValueString(),
		SSLConfSpecific:               plan.SSLConfSpecific.ValueString(),
		SSLCiphers:                    plan.SSLCiphers.ValueString(),
	}

	// XFF Header Names
	if !plan.XFFHeaderName.IsNull() && !plan.XFFHeaderName.IsUnknown() {
		plan.XFFHeaderName.ElementsAs(ctx, &pt.XFFHeaderName, false)
	}

	// SSL Protocols
	if !plan.SSLProtocols.IsNull() && !plan.SSLProtocols.IsUnknown() {
		plan.SSLProtocols.ElementsAs(ctx, &pt.SSLProtocols, false)
	}

	// Advanced Configuration
	pt.AdvancedConfiguration = []client.ProxyTemplateAdvancedConfig{}
	if !plan.AdvancedConfiguration.IsNull() && !plan.AdvancedConfiguration.IsUnknown() {
		var advModels []AdvancedConfigModel
		plan.AdvancedConfiguration.ElementsAs(ctx, &advModels, false)

		pt.AdvancedConfiguration = make([]client.ProxyTemplateAdvancedConfig, len(advModels))
		for i, m := range advModels {
			ac := client.ProxyTemplateAdvancedConfig{
				Name:          m.Name.ValueString(),
				Configuration: m.Configuration.ValueString(),
				Description:   m.Description.ValueString(),
			}
			if !m.Protocol.IsNull() && !m.Protocol.IsUnknown() {
				m.Protocol.ElementsAs(ctx, &ac.Protocol, false)
			}
			pt.AdvancedConfiguration[i] = ac
		}
	}

	return pt
}

func advancedConfigAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"name":          types.StringType,
		"protocol":      types.ListType{ElemType: types.StringType},
		"configuration": types.StringType,
		"description":   types.StringType,
	}
}
