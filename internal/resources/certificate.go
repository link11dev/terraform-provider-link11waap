package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
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
	_ resource.Resource                = &CertificateResource{}
	_ resource.ResourceWithImportState = &CertificateResource{}
)

// CertificateResource implements the certificate resource.
type CertificateResource struct {
	client *client.Client
}

// CertificateResourceModel describes the certificate resource data model.
type CertificateResourceModel struct {
	ConfigID      types.String `tfsdk:"config_id"`
	ID            types.String `tfsdk:"id"`
	CertBody      types.String `tfsdk:"cert_body"`
	PrivateKey    types.String `tfsdk:"private_key"`
	Domains       types.List   `tfsdk:"domains"`
	LEAutoRenew   types.Bool   `tfsdk:"le_auto_renew"`
	LEAutoReplace types.Bool   `tfsdk:"le_auto_replace"`
	Side          types.String `tfsdk:"side"`
	// Computed attributes
	Name     types.String `tfsdk:"name"`
	Subject  types.String `tfsdk:"subject"`
	Issuer   types.String `tfsdk:"issuer"`
	SAN      types.List   `tfsdk:"san"`
	Expires  types.String `tfsdk:"expires"`
	Uploaded types.String `tfsdk:"uploaded"`
	Revoked  types.Bool   `tfsdk:"revoked"`
	Links    types.List   `tfsdk:"links"`
}

// ProviderLinkModel describes the data model for a certificate provider link.
type ProviderLinkModel struct {
	Provider types.String `tfsdk:"provider"`
	Link     types.String `tfsdk:"link"`
	Region   types.String `tfsdk:"region"`
}

// NewCertificateResource creates a new certificate resource instance.
func NewCertificateResource() resource.Resource {
	return &CertificateResource{}
}

// Metadata returns the resource type name.
func (r *CertificateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_certificate"
}

// Schema defines the schema for the certificate resource.
func (r *CertificateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an SSL certificate in Link11 WAAP.",
		Attributes: map[string]schema.Attribute{
			"config_id": schema.StringAttribute{
				Description: "The configuration ID.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.StringAttribute{
				Description: "The unique identifier for the certificate.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"cert_body": schema.StringAttribute{
				Description: "The certificate PEM content. Required for uploaded certificates.",
				Optional:    true,
				Sensitive:   false,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"private_key": schema.StringAttribute{
				Description: "The private key PEM content. Required for uploaded certificates.",
				Optional:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"domains": schema.ListAttribute{
				Description: "Domains for Let's Encrypt certificate generation.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"le_auto_renew": schema.BoolAttribute{
				Description: "Enable automatic renewal for Let's Encrypt certificates.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"le_auto_replace": schema.BoolAttribute{
				Description: "Enable automatic replacement for Let's Encrypt certificates.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"side": schema.StringAttribute{
				Description: "Certificate side. Valid values: clientCA, server, serverToBackendMTLS, backendCA.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("server"),
				Validators: []validator.String{
					stringvalidator.OneOf("clientCA", "server", "serverToBackendMTLS", "backendCA"),
				},
			},
			// Computed attributes
			"name": schema.StringAttribute{
				Description: "Certificate name.",
				Computed:    true,
			},
			"subject": schema.StringAttribute{
				Description: "Certificate subject.",
				Computed:    true,
			},
			"issuer": schema.StringAttribute{
				Description: "Certificate issuer.",
				Computed:    true,
			},
			"san": schema.ListAttribute{
				Description: "Subject Alternative Names.",
				Computed:    true,
				ElementType: types.StringType,
			},
			"expires": schema.StringAttribute{
				Description: "Certificate expiration date.",
				Computed:    true,
			},
			"uploaded": schema.StringAttribute{
				Description: "Certificate upload timestamp.",
				Computed:    true,
			},
			"revoked": schema.BoolAttribute{
				Description: "Whether the certificate is revoked.",
				Computed:    true,
			},
			"links": schema.ListNestedAttribute{
				Description: "Provider links (cloud provider associations).",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"provider": schema.StringAttribute{
							Computed: true,
						},
						"link": schema.StringAttribute{
							Computed: true,
						},
						"region": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

// Configure configures the resource with the provider client.
func (r *CertificateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = providerutil.ConfigureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new certificate resource.
func (r *CertificateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan CertificateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	plan.ID = types.StringValue(generateCertificateID())

	certReq := &client.CertificateCreateRequest{
		ID:            plan.ID.ValueString(),
		CertBody:      plan.CertBody.ValueString(),
		PrivateKey:    plan.PrivateKey.ValueString(),
		LEAutoRenew:   plan.LEAutoRenew.ValueBool(),
		LEAutoReplace: plan.LEAutoReplace.ValueBool(),
		Side:          plan.Side.ValueString(),
	}

	var domains []string
	if !plan.Domains.IsNull() {
		resp.Diagnostics.Append(plan.Domains.ElementsAs(ctx, &domains, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	err := r.client.CreateCertificate(ctx, plan.ConfigID.ValueString(), plan.ID.ValueString(), certReq, domains)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Certificate",
			"Could not create certificate: "+err.Error(),
		)
		return
	}

	// Read back to get computed values
	cert, err := r.client.GetCertificate(ctx, plan.ConfigID.ValueString(), plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Certificate",
			"Could not read certificate after creation: "+err.Error(),
		)
		return
	}

	r.updateModelFromAPI(ctx, &plan, cert, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read reads the certificate resource.
func (r *CertificateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CertificateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cert, err := r.client.GetCertificate(ctx, state.ConfigID.ValueString(), state.ID.ValueString())
	if err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Certificate",
			"Could not read certificate: "+err.Error(),
		)
		return
	}

	r.updateModelFromAPI(ctx, &state, cert, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the certificate resource.
func (r *CertificateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan CertificateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.UpdateCertificate(
		ctx,
		plan.ConfigID.ValueString(),
		plan.ID.ValueString(),
		plan.LEAutoRenew.ValueBool(),
		plan.LEAutoReplace.ValueBool(),
		"",
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Certificate",
			"Could not update certificate: "+err.Error(),
		)
		return
	}

	// Read back to get updated values
	cert, err := r.client.GetCertificate(ctx, plan.ConfigID.ValueString(), plan.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Certificate",
			"Could not read certificate after update: "+err.Error(),
		)
		return
	}

	r.updateModelFromAPI(ctx, &plan, cert, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete deletes the certificate resource.
func (r *CertificateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CertificateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteCertificate(ctx, state.ConfigID.ValueString(), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Certificate",
			"Could not delete certificate: "+err.Error(),
		)
		return
	}
}

// ImportState imports an existing certificate resource.
func (r *CertificateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, "/")
	if len(parts) != 2 {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Expected import ID in format 'config_id/certificate_id', got: %s", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("config_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}

func (r *CertificateResource) updateModelFromAPI(ctx context.Context, model *CertificateResourceModel, cert *client.Certificate, diags *diag.Diagnostics) {
	model.Name = types.StringValue(cert.Name)
	model.Subject = types.StringValue(cert.Subject)
	model.Issuer = types.StringValue(cert.Issuer)
	model.Expires = types.StringValue(cert.Expires)
	model.Uploaded = types.StringValue(cert.Uploaded)
	model.Revoked = types.BoolValue(cert.Revoked)
	model.LEAutoRenew = types.BoolValue(cert.LEAutoRenew)
	model.LEAutoReplace = types.BoolValue(cert.LEAutoReplace)

	if cert.Side != "" {
		model.Side = types.StringValue(cert.Side)
	}

	// Handle SAN
	if cert.SAN != nil {
		san, d := types.ListValueFrom(ctx, types.StringType, cert.SAN)
		diags.Append(d...)
		model.SAN = san
	} else {
		model.SAN = types.ListNull(types.StringType)
	}

	// Handle provider links
	links := cert.Links
	if len(links) == 0 {
		links = cert.ProviderLinks
	}

	if len(links) > 0 {
		linkAttrTypes := map[string]attr.Type{
			"provider": types.StringType,
			"link":     types.StringType,
			"region":   types.StringType,
		}

		linkValues := make([]attr.Value, len(links))
		for i, link := range links {
			linkObj, d := types.ObjectValue(linkAttrTypes, map[string]attr.Value{
				"provider": types.StringValue(link.Provider),
				"link":     types.StringValue(link.Link),
				"region":   types.StringValue(link.Region),
			})
			diags.Append(d...)
			linkValues[i] = linkObj
		}

		linksList, d := types.ListValue(types.ObjectType{AttrTypes: linkAttrTypes}, linkValues)
		diags.Append(d...)
		model.Links = linksList
	} else {
		linkAttrTypes := map[string]attr.Type{
			"provider": types.StringType,
			"link":     types.StringType,
			"region":   types.StringType,
		}
		model.Links = types.ListNull(types.ObjectType{AttrTypes: linkAttrTypes})
	}
}
