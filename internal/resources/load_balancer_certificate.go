package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/link11/terraform-provider-link11waap/internal/client"
	"github.com/link11/terraform-provider-link11waap/internal/providerutil"
)

var (
	_ resource.Resource                = &LoadBalancerCertificateResource{}
	_ resource.ResourceWithImportState = &LoadBalancerCertificateResource{}
)

// LoadBalancerCertificateResource implements the load balancer certificate resource.
type LoadBalancerCertificateResource struct {
	client *client.Client
}

// LoadBalancerCertificateResourceModel describes the load balancer certificate resource data model.
type LoadBalancerCertificateResourceModel struct {
	ID               types.String `tfsdk:"id"`
	ConfigID         types.String `tfsdk:"config_id"`
	LoadBalancerName types.String `tfsdk:"load_balancer_name"`
	CertificateID    types.String `tfsdk:"certificate_id"`
	ProviderType     types.String `tfsdk:"provider_type"`
	Region           types.String `tfsdk:"region"`
	Listener         types.String `tfsdk:"listener"`
	ListenerPort     types.Int64  `tfsdk:"listener_port"`
	IsDefault        types.Bool   `tfsdk:"is_default"`
	ELBv2            types.Bool   `tfsdk:"elbv2"`
}

// NewLoadBalancerCertificateResource creates a new load balancer certificate resource instance.
func NewLoadBalancerCertificateResource() resource.Resource {
	return &LoadBalancerCertificateResource{}
}

// Metadata returns the resource type name.
func (r *LoadBalancerCertificateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_load_balancer_certificate"
}

// Schema defines the schema for the load balancer certificate resource.
func (r *LoadBalancerCertificateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Attaches a certificate to a load balancer in Link11 WAAP.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The composite identifier for this resource (config_id/load_balancer_name/certificate_id).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"config_id": schema.StringAttribute{
				Description: "The configuration ID.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"load_balancer_name": schema.StringAttribute{
				Description: "The name of the load balancer.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"certificate_id": schema.StringAttribute{
				Description: "The ID of the certificate to attach.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"provider_type": schema.StringAttribute{
				Description: "The cloud provider. Valid values: aws, gcp, link11.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("aws", "gcp", "link11"),
				},
			},
			"region": schema.StringAttribute{
				Description: "The cloud region.",
				Required:    true,
			},
			"listener": schema.StringAttribute{
				Description: "The listener identifier (ARN for AWS, name for others).",
				Required:    true,
			},
			"listener_port": schema.Int64Attribute{
				Description: "The listener port.",
				Required:    true,
			},
			"is_default": schema.BoolAttribute{
				Description: "Whether this is the default certificate for the load balancer.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"elbv2": schema.BoolAttribute{
				Description: "Use ELB v2 (Application Load Balancer) for AWS.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
		},
	}
}

// Configure configures the resource with the provider client.
func (r *LoadBalancerCertificateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = providerutil.ConfigureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new load balancer certificate resource.
func (r *LoadBalancerCertificateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan LoadBalancerCertificateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	opts := client.AttachCertificateOptions{
		Provider:     plan.ProviderType.ValueString(),
		Region:       plan.Region.ValueString(),
		Listener:     plan.Listener.ValueString(),
		ListenerPort: int(plan.ListenerPort.ValueInt64()),
		IsDefault:    plan.IsDefault.ValueBool(),
		ELBv2:        plan.ELBv2.ValueBool(),
	}

	err := r.client.AttachCertificateToLoadBalancer(
		ctx,
		plan.ConfigID.ValueString(),
		plan.LoadBalancerName.ValueString(),
		plan.CertificateID.ValueString(),
		opts,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Attaching Certificate",
			"Could not attach certificate to load balancer: "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s/%s/%s",
		plan.ConfigID.ValueString(),
		plan.LoadBalancerName.ValueString(),
		plan.CertificateID.ValueString(),
	))

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read reads the load balancer certificate resource.
func (r *LoadBalancerCertificateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state LoadBalancerCertificateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Fetch the certificate to obtain its provider link URLs.
	// The load balancer API returns full URLs in lb.Certificates, not plain IDs,
	// so we need the certificate's links to do a correct comparison.
	cert, err := r.client.GetCertificate(ctx, state.ConfigID.ValueString(), state.CertificateID.ValueString())
	if err != nil {
		if client.IsNotFoundError(err) {
			// Certificate itself was deleted; the attachment is gone.
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Certificate",
			"Could not read certificate: "+err.Error(),
		)
		return
	}

	// Build a set of known link URLs for this certificate so we can match
	// against whatever format the LB API returns (URL or plain ID).
	certLinkURLs := make(map[string]bool)
	for _, link := range cert.Links {
		if link.Link != "" {
			certLinkURLs[link.Link] = true
		}
	}
	// ProviderLinks carries the same provider URLs under a different JSON field;
	// include both to handle whichever the API populates.
	for _, link := range cert.ProviderLinks {
		if link.Link != "" {
			certLinkURLs[link.Link] = true
		}
	}

	// Verify the load balancer still has this certificate attached.
	lbs, err := r.client.ListLoadBalancers(ctx, state.ConfigID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Load Balancers",
			"Could not read load balancers: "+err.Error(),
		)
		return
	}

	certID := state.CertificateID.ValueString()

	// certURLMatch returns true when the given LB certificate string matches our cert.
	// Three checks (in order of precision):
	//  1. Exact URL match via the provider-link map (most reliable).
	//  2. Plain ID equality (for API responses that return IDs instead of URLs).
	//  3. URL path-suffix match: the API embeds the cert ID at the end of a URL
	//     path such as "https://.../sometestplanet-uploaded-test-cert" where the
	//     last segment either equals <certID> or ends with "-<certID>".
	certURLMatch := func(lbCert string) bool {
		if certLinkURLs[lbCert] || lbCert == certID {
			return true
		}
		if lastSlash := strings.LastIndex(lbCert, "/"); lastSlash >= 0 {
			lastSegment := lbCert[lastSlash+1:]
			if lastSegment == certID || strings.HasSuffix(lastSegment, "-"+certID) {
				return true
			}
		}
		return false
	}

	found := false
	for _, lb := range lbs {
		if lb.Name == state.LoadBalancerName.ValueString() {
			for _, lbCert := range lb.Certificates {
				if certURLMatch(lbCert) {
					found = true
					break
				}
			}
			// Check default certificate field too.
			if !found && certURLMatch(lb.DefaultCertificate) {
				found = true
			}
			break
		}
	}

	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the load balancer certificate resource.
func (r *LoadBalancerCertificateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Most attributes require replace, so updates are limited
	var plan LoadBalancerCertificateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete deletes the load balancer certificate resource.
func (r *LoadBalancerCertificateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state LoadBalancerCertificateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	certID := state.CertificateID.ValueString()

	// The detach API expects the full provider URL (e.g. "https://.../someslug-uploaded-test-cert")
	// as the certificate-id query param, not the plain cert ID stored in state.
	// Resolve the actual URL from the LB's certificate list using the same URL-suffix
	// matching used in Read.
	resolvedCertID := certID
	lbs, lbErr := r.client.ListLoadBalancers(ctx, state.ConfigID.ValueString())
	if lbErr == nil {
		certURLMatch := func(lbCert string) bool {
			if lbCert == certID {
				return true
			}
			if lastSlash := strings.LastIndex(lbCert, "/"); lastSlash >= 0 {
				lastSegment := lbCert[lastSlash+1:]
				return lastSegment == certID || strings.HasSuffix(lastSegment, "-"+certID)
			}
			return false
		}
	outer:
		for _, lb := range lbs {
			if lb.Name == state.LoadBalancerName.ValueString() {
				for _, lbCert := range lb.Certificates {
					if certURLMatch(lbCert) {
						resolvedCertID = lbCert
						break outer
					}
				}
				if certURLMatch(lb.DefaultCertificate) {
					resolvedCertID = lb.DefaultCertificate
				}
				break
			}
		}
	}

	opts := client.DetachCertificateOptions{
		Provider:      state.ProviderType.ValueString(),
		Region:        state.Region.ValueString(),
		CertificateID: resolvedCertID,
		Listener:      state.Listener.ValueString(),
		ListenerPort:  fmt.Sprintf("%d", state.ListenerPort.ValueInt64()),
		ELBv2:         state.ELBv2.ValueBool(),
	}

	err := r.client.DetachCertificateFromLoadBalancer(
		ctx,
		state.ConfigID.ValueString(),
		state.LoadBalancerName.ValueString(),
		opts,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Detaching Certificate",
			"Could not detach certificate from load balancer: "+err.Error(),
		)
		return
	}
}

// ImportState imports an existing load balancer certificate resource.
func (r *LoadBalancerCertificateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import ID format: config_id/load_balancer_name/certificate_id
	parts := strings.Split(req.ID, "/")
	if len(parts) != 3 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in format 'config_id/load_balancer_name/certificate_id', got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("config_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("load_balancer_name"), parts[1])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("certificate_id"), parts[2])...)
}
