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

var _ datasource.DataSource = &CertificatesDataSource{}

// CertificatesDataSource defines the data source for listing certificates.
type CertificatesDataSource struct {
	client *client.Client
}

// CertificatesDataSourceModel describes the data model for the certificates data source.
type CertificatesDataSourceModel struct {
	ConfigID     types.String           `tfsdk:"config_id"`
	ID           types.String           `tfsdk:"id"`
	Name         types.String           `tfsdk:"name"`
	Certificates []CertificateDataModel `tfsdk:"certificates"`
}

// CertificateDataModel represents a single certificate in the data source.
type CertificateDataModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Subject       types.String `tfsdk:"subject"`
	Issuer        types.String `tfsdk:"issuer"`
	SAN           types.List   `tfsdk:"san"`
	Expires       types.String `tfsdk:"expires"`
	Uploaded      types.String `tfsdk:"uploaded"`
	LEAutoRenew   types.Bool   `tfsdk:"le_auto_renew"`
	LEAutoReplace types.Bool   `tfsdk:"le_auto_replace"`
	Revoked       types.Bool   `tfsdk:"revoked"`
	Side          types.String `tfsdk:"side"`
	Links         types.List   `tfsdk:"links"`
}

// NewCertificatesDataSource creates a new certificates data source instance.
func NewCertificatesDataSource() datasource.DataSource {
	return &CertificatesDataSource{}
}

// Metadata returns the data source type name.
func (d *CertificatesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_certificates"
}

// Schema defines the schema for the certificates data source.
func (d *CertificatesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all certificates in a configuration.",
		Attributes: map[string]schema.Attribute{
			"config_id": schema.StringAttribute{
				Description: "Configuration ID.",
				Required:    true,
			},
			"id": schema.StringAttribute{
				Description: "Certificate ID. If specified, only the certificate with this ID will be returned.",
				Optional:    true,
			},
			"name": schema.StringAttribute{
				Description: "Certificate Name. If specified, only the certificate with this name will be returned.",
				Optional:    true,
			},
			"certificates": schema.ListNestedAttribute{
				Description: "List of certificates.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id":              schema.StringAttribute{Computed: true},
						"name":            schema.StringAttribute{Computed: true},
						"subject":         schema.StringAttribute{Computed: true},
						"issuer":          schema.StringAttribute{Computed: true},
						"san":             schema.ListAttribute{Computed: true, ElementType: types.StringType},
						"expires":         schema.StringAttribute{Computed: true},
						"uploaded":        schema.StringAttribute{Computed: true},
						"le_auto_renew":   schema.BoolAttribute{Computed: true},
						"le_auto_replace": schema.BoolAttribute{Computed: true},
						"revoked":         schema.BoolAttribute{Computed: true},
						"side":            schema.StringAttribute{Computed: true},
						"links": schema.ListNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"provider": schema.StringAttribute{Computed: true},
									"link":     schema.StringAttribute{Computed: true},
									"region":   schema.StringAttribute{Computed: true},
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
func (d *CertificatesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = providerutil.ConfigureClient(req.ProviderData, &resp.Diagnostics)
}

// Read reads the certificates data source.
func (d *CertificatesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CertificatesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var certificates []client.Certificate

	if !data.ID.IsNull() && !data.ID.IsUnknown() {
		cert, err := d.client.GetCertificate(ctx, data.ConfigID.ValueString(), data.ID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading Certificate",
				"Could not read certificate with ID "+data.ID.ValueString()+": "+err.Error(),
			)
			return
		}
		certificates = []client.Certificate{*cert}
	} else if !data.Name.IsNull() && !data.Name.IsUnknown() {
		allCerts, err := d.client.ListCertificates(ctx, data.ConfigID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading Certificates",
				"Could not read certificates: "+err.Error(),
			)
			return
		}
		for _, cert := range allCerts {
			if cert.Name == data.Name.ValueString() {
				certificates = []client.Certificate{cert}
				break
			}
		}
		if len(certificates) == 0 {
			resp.Diagnostics.AddError(
				"Certificate Not Found",
				"No certificate found with name: "+data.Name.ValueString(),
			)
			return
		}
	} else {
		var err error
		certificates, err = d.client.ListCertificates(ctx, data.ConfigID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Reading Certificates",
				"Could not read certificates: "+err.Error(),
			)
			return
		}
	}

	linkAttrTypes := map[string]attr.Type{
		"provider": types.StringType,
		"link":     types.StringType,
		"region":   types.StringType,
	}

	data.Certificates = make([]CertificateDataModel, len(certificates))
	for i, cert := range certificates {
		san, diags := types.ListValueFrom(ctx, types.StringType, cert.SAN)
		resp.Diagnostics.Append(diags...)

		// Handle links
		links := cert.Links
		if len(links) == 0 {
			links = cert.ProviderLinks
		}

		var linksList types.List
		if len(links) > 0 {
			linkValues := make([]attr.Value, len(links))
			for j, link := range links {
				linkObj, diags := types.ObjectValue(linkAttrTypes, map[string]attr.Value{
					"provider": types.StringValue(link.Provider),
					"link":     types.StringValue(link.Link),
					"region":   types.StringValue(link.Region),
				})
				resp.Diagnostics.Append(diags...)
				linkValues[j] = linkObj
			}
			linksList, diags = types.ListValue(types.ObjectType{AttrTypes: linkAttrTypes}, linkValues)
			resp.Diagnostics.Append(diags...)
		} else {
			linksList = types.ListNull(types.ObjectType{AttrTypes: linkAttrTypes})
		}

		data.Certificates[i] = CertificateDataModel{
			ID:            types.StringValue(cert.ID),
			Name:          types.StringValue(cert.Name),
			Subject:       types.StringValue(cert.Subject),
			Issuer:        types.StringValue(cert.Issuer),
			SAN:           san,
			Expires:       types.StringValue(cert.Expires),
			Uploaded:      types.StringValue(cert.Uploaded),
			LEAutoRenew:   types.BoolValue(cert.LEAutoRenew),
			LEAutoReplace: types.BoolValue(cert.LEAutoReplace),
			Revoked:       types.BoolValue(cert.Revoked),
			Side:          types.StringValue(cert.Side),
			Links:         linksList,
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
