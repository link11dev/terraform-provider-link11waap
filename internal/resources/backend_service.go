package resources

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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

// lookupHostFunc is the function used to resolve hostnames.
// It is a pkg-level variable so that tests can replace it with a stub
var lookupHostFunc = net.LookupHost

var (
	_ resource.Resource                = &BackendServiceResource{}
	_ resource.ResourceWithImportState = &BackendServiceResource{}
)

// BackendServiceResource implements the backend service resource.
type BackendServiceResource struct {
	client *client.Client
}

// BackendServiceResourceModel describes the backend service resource data model.
type BackendServiceResourceModel struct {
	ConfigID               types.String `tfsdk:"config_id"`
	ID                     types.String `tfsdk:"id"`
	Name                   types.String `tfsdk:"name"`
	Description            types.String `tfsdk:"description"`
	HTTP11                 types.Bool   `tfsdk:"http11"`
	TransportMode          types.String `tfsdk:"transport_mode"`
	Sticky                 types.String `tfsdk:"sticky"`
	StickyCookieName       types.String `tfsdk:"sticky_cookie_name"`
	LeastConn              types.Bool   `tfsdk:"least_conn"`
	BackHosts              types.Set    `tfsdk:"back_hosts"`
	MtlsCertificate        types.String `tfsdk:"mtls_certificate"`
	MtlsTrustedCertificate types.String `tfsdk:"mtls_trusted_certificate"`
}

// BackendHostModel describes the data model for a backend host.
type BackendHostModel struct {
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

// NewBackendServiceResource creates a new backend service resource instance.
func NewBackendServiceResource() resource.Resource {
	return &BackendServiceResource{}
}

// Metadata returns the resource type name.
func (r *BackendServiceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_backend_service"
}

// Schema defines the schema for the backend service resource.
func (r *BackendServiceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Backend Service in Link11 WAAP.",
		Attributes: map[string]schema.Attribute{
			"config_id": schema.StringAttribute{
				Description: "The configuration ID.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.StringAttribute{
				Description: "The unique identifier for the backend service.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the backend service.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
			},
			"description": schema.StringAttribute{
				Description: "Description of the backend service.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"http11": schema.BoolAttribute{
				Description: "Whether to use HTTP/1.1 for upstream connections.",
				Required:    true,
			},
			"transport_mode": schema.StringAttribute{
				Description: "Transport protocol. Valid values: default, http, https, port_bridge.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("default", "http", "https", "port_bridge"),
				},
			},
			"sticky": schema.StringAttribute{
				Description: "Load balancing stickiness model. Valid values: none, autocookie, customcookie, iphash, least_conn.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("none", "autocookie", "customcookie", "iphash", "least_conn"),
				},
			},
			"sticky_cookie_name": schema.StringAttribute{
				Description: "Custom cookie name for sticky sessions (used when sticky is 'customcookie').",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
			},
			"least_conn": schema.BoolAttribute{
				Description: "Whether to use least-connections load balancing.",
				Required:    true,
			},
			"mtls_certificate": schema.StringAttribute{
				Description: "ID of mTLS certificate attached to backend service.",
				Optional:    true,
			},
			"mtls_trusted_certificate": schema.StringAttribute{
				Description: "ID of CA certificate attached to backend service.",
				Optional:    true,
			},
		},
		Blocks: map[string]schema.Block{
			"back_hosts": schema.SetNestedBlock{
				Description: "Set of backend hosts.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"host": schema.StringAttribute{
							Description: "URI or IP address of the backend host.",
							Required:    true,
						},
						"http_ports": schema.ListAttribute{
							Description: "HTTP port numbers (e.g. [80, 8080]).",
							Required:    true,
							ElementType: types.Int64Type,
						},
						"https_ports": schema.ListAttribute{
							Description: "HTTPS port numbers (e.g. [443, 8443]).",
							Required:    true,
							ElementType: types.Int64Type,
						},
						"weight": schema.Int64Attribute{
							Description: "Weight for load balancing.",
							Required:    true,
						},
						"max_fails": schema.Int64Attribute{
							Description: "Maximum number of failed attempts before marking host as down.",
							Required:    true,
						},
						"fail_timeout": schema.Int64Attribute{
							Description: "Timeout in seconds after which failed attempts are reset.",
							Required:    true,
						},
						"down": schema.BoolAttribute{
							Description: "Whether the host is marked as down.",
							Required:    true,
						},
						"monitor_state": schema.StringAttribute{
							Description: "Monitor state of the host.",
							Required:    true,
						},
						"backup": schema.BoolAttribute{
							Description: "Whether this is a backup host.",
							Required:    true,
						},
					},
				},
			},
		},
	}
}

// Configure configures the resource with the provider client.
func (r *BackendServiceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = providerutil.ConfigureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new backend service resource.
func (r *BackendServiceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan BackendServiceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = types.StringValue(generateID())

	bs := buildBackendServiceAPIModel(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.CreateBackendService(ctx, plan.ConfigID.ValueString(), plan.ID.ValueString(), bs)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Backend Service",
			"Could not create backend service: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read reads the backend service resource.
func (r *BackendServiceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state BackendServiceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bs, err := r.client.GetBackendService(ctx, state.ConfigID.ValueString(), state.ID.ValueString())
	if err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Backend Service",
			"Could not read backend service: "+err.Error(),
		)
		return
	}

	state.Name = types.StringValue(bs.Name)
	state.Description = types.StringValue(bs.Description)
	state.HTTP11 = types.BoolValue(bs.HTTP11)
	state.TransportMode = types.StringValue(bs.TransportMode)
	state.Sticky = types.StringValue(bs.Sticky)
	state.StickyCookieName = types.StringValue(bs.StickyCookieName)
	state.LeastConn = types.BoolValue(bs.LeastConn)

	// Optional string fields
	if bs.MtlsCertificate != "" {
		state.MtlsCertificate = types.StringValue(bs.MtlsCertificate)
	} else {
		state.MtlsCertificate = types.StringNull()
	}
	if bs.MtlsTrustedCertificate != "" {
		state.MtlsTrustedCertificate = types.StringValue(bs.MtlsTrustedCertificate)
	} else {
		state.MtlsTrustedCertificate = types.StringNull()
	}

	// BackHosts -- convert []client.BackendHost -> types.Set of BackendHostModel
	if bs.BackHosts != nil {
		hostModels := make([]BackendHostModel, len(bs.BackHosts))
		for i, bh := range bs.BackHosts {
			httpPorts, diags := types.ListValueFrom(ctx, types.Int64Type, intSliceToInt64(bh.HTTPPorts))
			resp.Diagnostics.Append(diags...)
			httpsPorts, diags := types.ListValueFrom(ctx, types.Int64Type, intSliceToInt64(bh.HTTPSPorts))
			resp.Diagnostics.Append(diags...)

			hostModels[i] = BackendHostModel{
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

		hostSet, diags := types.SetValueFrom(ctx, types.ObjectType{
			AttrTypes: backendHostAttrTypes(),
		}, hostModels)
		resp.Diagnostics.Append(diags...)
		state.BackHosts = hostSet
	} else {
		state.BackHosts = types.SetNull(types.ObjectType{AttrTypes: backendHostAttrTypes()})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the backend service resource.
func (r *BackendServiceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan BackendServiceResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bs := buildBackendServiceAPIModel(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.UpdateBackendService(ctx, plan.ConfigID.ValueString(), plan.ID.ValueString(), bs)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Backend Service",
			"Could not update backend service: "+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete deletes the backend service resource.
func (r *BackendServiceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state BackendServiceResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteBackendService(ctx, state.ConfigID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Backend Service",
			"Could not delete backend service: "+err.Error(),
		)
		return
	}
}

// ImportState imports an existing backend service resource.
func (r *BackendServiceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in format 'config_id/backend_service_id', got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("config_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

func buildBackendServiceAPIModel(ctx context.Context, plan *BackendServiceResourceModel, diags *diag.Diagnostics) *client.BackendService {
	bs := &client.BackendService{
		ID:               plan.ID.ValueString(),
		Name:             plan.Name.ValueString(),
		Description:      plan.Description.ValueString(),
		HTTP11:           plan.HTTP11.ValueBool(),
		TransportMode:    plan.TransportMode.ValueString(),
		Sticky:           plan.Sticky.ValueString(),
		StickyCookieName: plan.StickyCookieName.ValueString(),
		LeastConn:        plan.LeastConn.ValueBool(),
	}

	if !plan.MtlsCertificate.IsNull() && !plan.MtlsCertificate.IsUnknown() {
		bs.MtlsCertificate = plan.MtlsCertificate.ValueString()
	}
	if !plan.MtlsTrustedCertificate.IsNull() && !plan.MtlsTrustedCertificate.IsUnknown() {
		bs.MtlsTrustedCertificate = plan.MtlsTrustedCertificate.ValueString()
	}

	// BackHosts
	if !plan.BackHosts.IsNull() && !plan.BackHosts.IsUnknown() {
		var hostModels []BackendHostModel
		diags.Append(plan.BackHosts.ElementsAs(ctx, &hostModels, false)...)

		bs.BackHosts = make([]client.BackendHost, len(hostModels))
		for i, m := range hostModels {
			hostValue := m.Host.ValueString()

			// Validate hostname resolution: if the value is not an IP address,
			// treat it as a hostname and verify that it resolves via DNS.
			// Because if hostname is not resolvable, we get a non-obvious error
			//  when try to call `Publish` action
			if err := validateHostResolution(hostValue); err != nil {
				diags.AddError(
					"Backend Host Resolution Failed",
					err.Error(),
				)
				return nil
			}

			var httpPorts, httpsPorts []int64
			diags.Append(m.HTTPPorts.ElementsAs(ctx, &httpPorts, false)...)
			diags.Append(m.HTTPSPorts.ElementsAs(ctx, &httpsPorts, false)...)

			bs.BackHosts[i] = client.BackendHost{
				Host:         hostValue,
				HTTPPorts:    int64SliceToInt(httpPorts),
				HTTPSPorts:   int64SliceToInt(httpsPorts),
				Weight:       int(m.Weight.ValueInt64()),
				MaxFails:     int(m.MaxFails.ValueInt64()),
				FailTimeout:  int(m.FailTimeout.ValueInt64()),
				Down:         m.Down.ValueBool(),
				MonitorState: m.MonitorState.ValueString(),
				Backup:       m.Backup.ValueBool(),
			}
		}
	}

	return bs
}

// backendHostAttrTypes returns the attribute type map for BackendHostModel.
// This is required when constructing types.SetValueFrom for nested objects.
func backendHostAttrTypes() map[string]attr.Type {
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

// intSliceToInt64 converts []int to []int64 for Terraform types.
func intSliceToInt64(in []int) []int64 {
	out := make([]int64, len(in))
	for i, v := range in {
		out[i] = int64(v)
	}
	return out
}

// int64SliceToInt converts []int64 to []int for the API client.
func int64SliceToInt(in []int64) []int {
	out := make([]int, len(in))
	for i, v := range in {
		out[i] = int(v)
	}
	return out
}

// validateHostResolution checks whether the given host value is resolvable.
// If the value is an IP address (IPv4 or IPv6), no DNS lookup is performed.
// If it is a hostname, it must resolve via DNS; otherwise an error is returned.
func validateHostResolution(host string) error {
	if net.ParseIP(host) != nil {
		return nil
	}

	_, err := lookupHostFunc(host)
	if err != nil {
		return fmt.Errorf("hostname %q does not resolve: %s", host, err.Error())
	}

	return nil
}
