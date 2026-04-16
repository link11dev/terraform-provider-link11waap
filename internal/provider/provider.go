// Package provider implements the Terraform provider for Link11 WAAP.
package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/link11/terraform-provider-link11waap/internal/client"
	"github.com/link11/terraform-provider-link11waap/internal/datasources"
	"github.com/link11/terraform-provider-link11waap/internal/resources"
)

var _ provider.Provider = &Link11WaapProvider{}

// Link11WaapProvider implements the Terraform provider for Link11 WAAP.
type Link11WaapProvider struct {
	version string
}

// Link11WaapProviderModel describes the provider configuration model.
type Link11WaapProviderModel struct {
	Domain types.String `tfsdk:"domain"`
	APIKey types.String `tfsdk:"api_key"`
}

// New returns a new instance of the Link11 WAAP provider.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &Link11WaapProvider{
			version: version,
		}
	}
}

// Metadata returns the provider type name and version.
func (p *Link11WaapProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "link11waap"
	resp.Version = p.version
}

// Schema defines the schema for the provider configuration.
func (p *Link11WaapProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider for Link11 WAAP (Web Application and API Protection).",
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				Description: "The Link11 WAAP domain (e.g., 'customer.app.reblaze.io'). Can also be set via LINK11_DOMAIN environment variable.",
				Optional:    true,
			},
			"api_key": schema.StringAttribute{
				Description: "The API key for authentication. Can also be set via LINK11_API_KEY environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}
}

// Configure initializes the provider with the configured domain and API key, creating a client for use by resources and data sources.
func (p *Link11WaapProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config Link11WaapProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get domain from config or environment
	domain := os.Getenv("LINK11_DOMAIN")
	if !config.Domain.IsNull() {
		domain = config.Domain.ValueString()
	}
	if domain == "" {
		resp.Diagnostics.AddError(
			"Missing Domain",
			"The provider requires a domain to be configured. Set it in the provider block or via LINK11_DOMAIN environment variable.",
		)
		return
	}

	// Get API key from config or environment
	apiKey := os.Getenv("LINK11_API_KEY")
	if !config.APIKey.IsNull() {
		apiKey = config.APIKey.ValueString()
	}
	if apiKey == "" {
		resp.Diagnostics.AddError(
			"Missing API Key",
			"The provider requires an API key to be configured. Set it in the provider block or via LINK11_API_KEY environment variable.",
		)
		return
	}

	// Create client
	c, err := client.New(client.Config{
		Domain:    domain,
		APIKey:    apiKey,
		UserAgent: "terraform-provider-link11waap/" + p.version,
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Client",
			"An error occurred while creating the API client: "+err.Error(),
		)
		return
	}

	resp.DataSourceData = c
	resp.ResourceData = c
}

// Resources returns the resource implementations for the provider.
func (p *Link11WaapProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		// Existing
		resources.NewServerGroupResource,
		resources.NewCertificateResource,
		resources.NewLoadBalancerCertificateResource,
		resources.NewLoadBalancerRegionsResource,
		resources.NewPublishResource,
		resources.NewSecurityPolicyResource,
		resources.NewACLProfileResource,
		resources.NewUserResource,
		resources.NewBackendServiceResource,
		resources.NewMobileApplicationGroupResource,
		resources.NewRateLimitRuleResource,
		resources.NewEdgeFunctionResource,
		resources.NewProxyTemplateResource,
		resources.NewGlobalFilterResource,
		resources.NewPlanetTrustedNetsResource,
	}
}

// DataSources returns the data source implementations for the provider.
func (p *Link11WaapProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		// Existing
		datasources.NewConfigDataSource,
		datasources.NewServerGroupsDataSource,
		datasources.NewCertificatesDataSource,
		datasources.NewLoadBalancersDataSource,
		datasources.NewLoadBalancerRegionsDataSource,
		datasources.NewSecurityPoliciesDataSource,
		datasources.NewACLProfilesDataSource,
		datasources.NewUsersDataSource,
		datasources.NewBackendServicesDataSource,
		datasources.NewMobileApplicationGroupsDataSource,
		datasources.NewRateLimitRulesDataSource,
		datasources.NewEdgeFunctionsDataSource,
		datasources.NewProxyTemplatesDataSource,
		datasources.NewGlobalFiltersDataSource,
		datasources.NewGlobalFilterDataSource,
		datasources.NewPlanetTrustedNetsDataSource,
	}
}
