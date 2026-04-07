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

var _ datasource.DataSource = &BackendServicesDataSource{}

// BackendServicesDataSource defines the data source for listing backend services.
type BackendServicesDataSource struct {
	client *client.Client
}

// BackendServicesDataSourceModel describes the data model for the backend services data source.
type BackendServicesDataSourceModel struct {
	ConfigID        types.String              `tfsdk:"config_id"`
	ID              types.String              `tfsdk:"id"`
	Name            types.String              `tfsdk:"name"`
	BackendServices []BackendServiceDataModel `tfsdk:"backend_services"`
}

// BackendServiceDataModel represents a single backend service in the data source.
type BackendServiceDataModel struct {
	ID                     types.String `tfsdk:"id"`
	Name                   types.String `tfsdk:"name"`
	Description            types.String `tfsdk:"description"`
	HTTP11                 types.Bool   `tfsdk:"http11"`
	TransportMode          types.String `tfsdk:"transport_mode"`
	Sticky                 types.String `tfsdk:"sticky"`
	StickyCookieName       types.String `tfsdk:"sticky_cookie_name"`
	LeastConn              types.Bool   `tfsdk:"least_conn"`
	MtlsCertificate        types.String `tfsdk:"mtls_certificate"`
	MtlsTrustedCertificate types.String `tfsdk:"mtls_trusted_certificate"`
	BackHosts              types.Set    `tfsdk:"back_hosts"`
}

// BackendHostDataModel represents a backend host entry in the data source.
type BackendHostDataModel struct {
	Host         types.String `tfsdk:"host"`
	HTTPPorts    types.List   `tfsdk:"http_ports"`
	HTTPSPorts   types.List   `tfsdk:"https_ports"`
	Weight       types.Int64  `tfsdk:"weight"`
	MaxFails     types.Int64  `tfsdk:"max_fails"`
	FailTimeout  types.Int64  `tfsdk:"fail_timeout"`
	Down         types.Bool   `tfsdk:"down"`
	MonitorState types.String `tfsdk:"monitor_state"`
	Backup       types.Bool   `tfsdk:"backup"`
}

// NewBackendServicesDataSource creates a new backend services data source instance.
func NewBackendServicesDataSource() datasource.DataSource {
	return &BackendServicesDataSource{}
}

// Metadata returns the data source type name.
func (d *BackendServicesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_backend_services"
}

// Schema defines the schema for the backend services data source.
func (d *BackendServicesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all backend services in a configuration.",
		Attributes: map[string]schema.Attribute{
			"config_id": schema.StringAttribute{
				Description: "Configuration ID.",
				Required:    true,
			},
			"id": schema.StringAttribute{
				Description: "Backend Service ID. If specified, only the backend service with this ID will be returned.",
				Optional:    true,
			},
			"name": schema.StringAttribute{
				Description: "Backend Service Name. If specified, only the backend service with this name will be returned.",
				Optional:    true,
			},
			"backend_services": schema.ListNestedAttribute{
				Description: "List of backend services.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":                       schema.StringAttribute{Computed: true},
						"name":                     schema.StringAttribute{Computed: true},
						"description":              schema.StringAttribute{Computed: true},
						"http11":                   schema.BoolAttribute{Computed: true},
						"transport_mode":           schema.StringAttribute{Computed: true},
						"sticky":                   schema.StringAttribute{Computed: true},
						"sticky_cookie_name":       schema.StringAttribute{Computed: true},
						"least_conn":               schema.BoolAttribute{Computed: true},
						"mtls_certificate":         schema.StringAttribute{Computed: true},
						"mtls_trusted_certificate": schema.StringAttribute{Computed: true},
						"back_hosts": schema.SetNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"host":          schema.StringAttribute{Computed: true},
									"http_ports":    schema.ListAttribute{Computed: true, ElementType: types.Int64Type},
									"https_ports":   schema.ListAttribute{Computed: true, ElementType: types.Int64Type},
									"weight":        schema.Int64Attribute{Computed: true},
									"max_fails":     schema.Int64Attribute{Computed: true},
									"fail_timeout":  schema.Int64Attribute{Computed: true},
									"down":          schema.BoolAttribute{Computed: true},
									"monitor_state": schema.StringAttribute{Computed: true},
									"backup":        schema.BoolAttribute{Computed: true},
								},
							},
						},
					},
				},
			},
		},
	}
}

// Configure configures the data source with the provider client.
func (d *BackendServicesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = providerutil.ConfigureClient(req.ProviderData, &resp.Diagnostics)
}

// Read reads the backend services data source.
func (d *BackendServicesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data BackendServicesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var services []client.BackendService

	if !data.ID.IsNull() && !data.ID.IsUnknown() {
		bs, err := d.client.GetBackendService(ctx, data.ConfigID.ValueString(), data.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading Backend Service",
				"Could not read backend service with ID "+data.ID.ValueString()+": "+err.Error(),
			)
			return
		}
		services = []client.BackendService{*bs}
	} else if !data.Name.IsNull() && !data.Name.IsUnknown() {
		allServices, err := d.client.ListBackendServices(ctx, data.ConfigID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading Backend Services",
				"Could not read backend services: "+err.Error(),
			)
			return
		}
		for _, bs := range allServices {
			if bs.Name == data.Name.ValueString() {
				services = []client.BackendService{bs}
				break
			}
		}
		if len(services) == 0 {
			resp.Diagnostics.AddError(
				"Backend Service Not Found",
				"No backend service found with name: "+data.Name.ValueString(),
			)
			return
		}
	} else {
		var err error
		services, err = d.client.ListBackendServices(ctx, data.ConfigID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading Backend Services",
				"Could not read backend services: "+err.Error(),
			)
			return
		}
	}

	data.BackendServices = make([]BackendServiceDataModel, len(services))
	for i, bs := range services {
		// Build back_hosts nested set (API returns a list, convert to set)
		var backHostsSet types.Set
		if bs.BackHosts != nil {
			hostModels := make([]BackendHostDataModel, len(bs.BackHosts))
			for j, bh := range bs.BackHosts {
				httpPorts, diags := types.ListValueFrom(ctx, types.Int64Type, dsIntSliceToInt64(bh.HTTPPorts))
				resp.Diagnostics.Append(diags...)
				httpsPorts, diags := types.ListValueFrom(ctx, types.Int64Type, dsIntSliceToInt64(bh.HTTPSPorts))
				resp.Diagnostics.Append(diags...)

				hostModels[j] = BackendHostDataModel{
					Host:         types.StringValue(bh.Host),
					HTTPPorts:    httpPorts,
					HTTPSPorts:   httpsPorts,
					Weight:       types.Int64Value(int64(bh.Weight)),
					MaxFails:     types.Int64Value(int64(bh.MaxFails)),
					FailTimeout:  types.Int64Value(int64(bh.FailTimeout)),
					Down:         types.BoolValue(bh.Down),
					MonitorState: types.StringValue(bh.MonitorState),
					Backup:       types.BoolValue(bh.Backup),
				}
			}

			hs, diags := types.SetValueFrom(ctx, types.ObjectType{
				AttrTypes: dsBackendHostAttrTypes(),
			}, hostModels)
			resp.Diagnostics.Append(diags...)
			backHostsSet = hs
		} else {
			backHostsSet = types.SetNull(types.ObjectType{AttrTypes: dsBackendHostAttrTypes()})
		}

		data.BackendServices[i] = BackendServiceDataModel{
			ID:                     types.StringValue(bs.ID),
			Name:                   types.StringValue(bs.Name),
			Description:            types.StringValue(bs.Description),
			HTTP11:                 types.BoolValue(bs.HTTP11),
			TransportMode:          types.StringValue(bs.TransportMode),
			Sticky:                 types.StringValue(bs.Sticky),
			StickyCookieName:       types.StringValue(bs.StickyCookieName),
			LeastConn:              types.BoolValue(bs.LeastConn),
			MtlsCertificate:        types.StringValue(bs.MtlsCertificate),
			MtlsTrustedCertificate: types.StringValue(bs.MtlsTrustedCertificate),
			BackHosts:              backHostsSet,
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// dsBackendHostAttrTypes returns the attribute type map for BackendHostDataModel.
func dsBackendHostAttrTypes() map[string]attr.Type {
	return map[string]attr.Type{
		"host":          types.StringType,
		"http_ports":    types.ListType{ElemType: types.Int64Type},
		"https_ports":   types.ListType{ElemType: types.Int64Type},
		"weight":        types.Int64Type,
		"max_fails":     types.Int64Type,
		"fail_timeout":  types.Int64Type,
		"down":          types.BoolType,
		"monitor_state": types.StringType,
		"backup":        types.BoolType,
	}
}

// dsIntSliceToInt64 converts []int to []int64 for Terraform types.
func dsIntSliceToInt64(in []int) []int64 {
	out := make([]int64, len(in))
	for i, v := range in {
		out[i] = int64(v)
	}
	return out
}
