package datasources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/link11/terraform-provider-link11waap/internal/client"
	"github.com/link11/terraform-provider-link11waap/internal/providerutil"
)

var _ datasource.DataSource = &ProxyTemplatesDataSource{}

type ProxyTemplatesDataSource struct {
	client *client.Client
}

type ProxyTemplatesDataSourceModel struct {
	ConfigID       types.String               `tfsdk:"config_id"`
	ProxyTemplates []ProxyTemplateDataModel   `tfsdk:"proxy_templates"`
}

type ProxyTemplateDataModel struct {
	Name                          types.String `tfsdk:"name"`
	Description                   types.String `tfsdk:"description"`
	ACAOHeader                    types.Bool   `tfsdk:"acao_header"`
	XFFHeaderName                 types.String `tfsdk:"xff_header_name"`
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

type AdvancedConfigDataModel struct {
	Name          types.String `tfsdk:"name"`
	Protocol      types.List   `tfsdk:"protocol"`
	Configuration types.String `tfsdk:"configuration"`
	Description   types.String `tfsdk:"description"`
}

func NewProxyTemplatesDataSource() datasource.DataSource {
	return &ProxyTemplatesDataSource{}
}

func (d *ProxyTemplatesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_proxy_templates"
}

func (d *ProxyTemplatesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all proxy templates in a configuration.",
		Attributes: map[string]schema.Attribute{
			"config_id": schema.StringAttribute{
				Description: "Configuration ID.",
				Required:    true,
			},
			"proxy_templates": schema.ListNestedAttribute{
				Description: "List of proxy templates.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name":                              schema.StringAttribute{Computed: true},
						"description":                       schema.StringAttribute{Computed: true},
						"acao_header":                        schema.BoolAttribute{Computed: true},
						"xff_header_name":                    schema.StringAttribute{Computed: true},
						"xrealip_header_name":                schema.StringAttribute{Computed: true},
						"proxy_connect_timeout":              schema.StringAttribute{Computed: true},
						"proxy_read_timeout":                 schema.StringAttribute{Computed: true},
						"proxy_send_timeout":                 schema.StringAttribute{Computed: true},
						"upstream_host":                      schema.StringAttribute{Computed: true},
						"client_body_timeout":                schema.StringAttribute{Computed: true},
						"client_body_buffer_size":            schema.StringAttribute{Computed: true},
						"client_header_timeout":              schema.StringAttribute{Computed: true},
						"client_header_buffer_size":          schema.StringAttribute{Computed: true},
						"client_max_body_size":               schema.StringAttribute{Computed: true},
						"keepalive_timeout":                  schema.StringAttribute{Computed: true},
						"send_timeout":                       schema.StringAttribute{Computed: true},
						"limit_req_rate":                     schema.StringAttribute{Computed: true},
						"limit_req_burst":                    schema.StringAttribute{Computed: true},
						"mask_headers":                       schema.StringAttribute{Computed: true},
						"custom_listener":                    schema.BoolAttribute{Computed: true},
						"large_client_header_buffers_count":  schema.StringAttribute{Computed: true},
						"large_client_header_buffers_size":   schema.StringAttribute{Computed: true},
						"conf_specific":                      schema.StringAttribute{Computed: true},
						"ssl_conf_specific":                  schema.StringAttribute{Computed: true},
						"ssl_ciphers":                        schema.StringAttribute{Computed: true},
						"ssl_protocols":                      schema.ListAttribute{Computed: true, ElementType: types.StringType},
						"advanced_configuration": schema.ListNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"name":          schema.StringAttribute{Computed: true},
									"protocol":      schema.ListAttribute{Computed: true, ElementType: types.StringType},
									"configuration": schema.StringAttribute{Computed: true},
									"description":   schema.StringAttribute{Computed: true},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *ProxyTemplatesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = providerutil.ConfigureClient(req.ProviderData, &resp.Diagnostics)
}

func (d *ProxyTemplatesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProxyTemplatesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	templates, err := d.client.ListProxyTemplates(ctx, data.ConfigID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Proxy Templates",
			"Could not read proxy templates: "+err.Error(),
		)
		return
	}

	data.ProxyTemplates = make([]ProxyTemplateDataModel, len(templates))
	for i, pt := range templates {
		var sslProtocols types.List
		if pt.SSLProtocols != nil {
			sl, diags := types.ListValueFrom(ctx, types.StringType, pt.SSLProtocols)
			resp.Diagnostics.Append(diags...)
			sslProtocols = sl
		} else {
			sslProtocols = types.ListNull(types.StringType)
		}

		var advancedConfiguration types.List
		if pt.AdvancedConfiguration != nil {
			advModels := make([]AdvancedConfigDataModel, len(pt.AdvancedConfiguration))
			for j, ac := range pt.AdvancedConfiguration {
				model := AdvancedConfigDataModel{
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
				advModels[j] = model
			}
			advList, diags := types.ListValueFrom(ctx, types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"name":          types.StringType,
					"protocol":      types.ListType{ElemType: types.StringType},
					"configuration": types.StringType,
					"description":   types.StringType,
				},
			}, advModels)
			resp.Diagnostics.Append(diags...)
			advancedConfiguration = advList
		} else {
			advancedConfiguration = types.ListNull(types.ObjectType{
				AttrTypes: map[string]attr.Type{
					"name":          types.StringType,
					"protocol":      types.ListType{ElemType: types.StringType},
					"configuration": types.StringType,
					"description":   types.StringType,
				},
			})
		}

		data.ProxyTemplates[i] = ProxyTemplateDataModel{
			Name:                          types.StringValue(pt.Name),
			Description:                   types.StringValue(pt.Description),
			ACAOHeader:                    types.BoolValue(pt.ACAOHeader),
			XFFHeaderName:                 types.StringValue(pt.XFFHeaderName),
			XRealIPHeaderName:             types.StringValue(pt.XRealIPHeaderName),
			ProxyConnectTimeout:           types.StringValue(pt.ProxyConnectTimeout),
			ProxyReadTimeout:              types.StringValue(pt.ProxyReadTimeout),
			ProxySendTimeout:              types.StringValue(pt.ProxySendTimeout),
			UpstreamHost:                  types.StringValue(pt.UpstreamHost),
			ClientBodyTimeout:             types.StringValue(pt.ClientBodyTimeout),
			ClientBodyBufferSize:          types.StringValue(pt.ClientBodyBufferSize),
			ClientHeaderTimeout:           types.StringValue(pt.ClientHeaderTimeout),
			ClientHeaderBufferSize:        types.StringValue(pt.ClientHeaderBufferSize),
			ClientMaxBodySize:             types.StringValue(pt.ClientMaxBodySize),
			KeepaliveTimeout:              types.StringValue(pt.KeepaliveTimeout),
			SendTimeout:                   types.StringValue(pt.SendTimeout),
			LimitReqRate:                  types.StringValue(pt.LimitReqRate),
			LimitReqBurst:                 types.StringValue(pt.LimitReqBurst),
			MaskHeaders:                   types.StringValue(pt.MaskHeaders),
			CustomListener:                types.BoolValue(pt.CustomListener),
			LargeClientHeaderBuffersCount: types.StringValue(pt.LargeClientHeaderBuffersCount),
			LargeClientHeaderBuffersSize:  types.StringValue(pt.LargeClientHeaderBuffersSize),
			ConfSpecific:                  types.StringValue(pt.ConfSpecific),
			SSLConfSpecific:               types.StringValue(pt.SSLConfSpecific),
			SSLCiphers:                    types.StringValue(pt.SSLCiphers),
			SSLProtocols:                  sslProtocols,
			AdvancedConfiguration:         advancedConfiguration,
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
