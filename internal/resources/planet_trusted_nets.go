// Package resources contains the implementation of Terraform resources for Link11 WAAP.
package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/netip"
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

const defaultPlanetEntryID = "__default__"

// defaultIchallengeJSON is the verbatim ichallenge object embedded as a Go constant.
// It is sent as part of every PUT request but is never exposed in the Terraform schema.
var defaultIchallengeJSON = json.RawMessage(`{"attrs":{"html":{"dir":"ltr"}},"palette":["#4285F4","#FBBC05","#34A853","#F5D57A"],"position":{"style":"logo"},"lang":{"en":{"click":{"0":"Click inside","1":"Click outside"},"title":"Human verification in process...","src":"data:image/png;base64,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"}}}`)

const defaultNoHostCertName = "placeholder-certificate"
const defaultNoHostSSLCiphers = "ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:DHE-RSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-SHA384:ECDHE-ECDSA-AES256-SHA384:ECDHE-ECDSA-AES256-SHA:ECDHE-RSA-AES256-SHA:DHE-RSA-AES256-SHA256:DHE-RSA-AES256-SHA:!SHA1:!SHA256:!SHA384:!DSS:!aNULL"

var defaultNoHostSSLProtocols = []string{"TLSv1.2", "TLSv1.3"}

var (
	_ resource.Resource                     = &PlanetTrustedNetsResource{}
	_ resource.ResourceWithImportState      = &PlanetTrustedNetsResource{}
	_ resource.ResourceWithConfigValidators = &PlanetTrustedNetsResource{}
)

// PlanetTrustedNetsResource implements the planet_trusted_nets resource.
type PlanetTrustedNetsResource struct {
	client *client.Client
}

// PlanetTrustedNetsResourceModel describes the resource data model.
type PlanetTrustedNetsResourceModel struct {
	ConfigID    types.String      `tfsdk:"config_id"`
	ID          types.String      `tfsdk:"id"`
	Name        types.String      `tfsdk:"name"`
	TrustedNets []TrustedNetModel `tfsdk:"trusted_nets"`
}

// TrustedNetModel describes a single trusted_nets block entry.
type TrustedNetModel struct {
	Source  types.String `tfsdk:"source"`
	Address types.String `tfsdk:"address"`
	GfID    types.String `tfsdk:"gf_id"`
	Comment types.String `tfsdk:"comment"`
}

// NewPlanetTrustedNetsResource creates a new resource instance.
func NewPlanetTrustedNetsResource() resource.Resource {
	return &PlanetTrustedNetsResource{}
}

// Metadata returns the resource type name.
func (r *PlanetTrustedNetsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_planet_trusted_nets"
}

// Schema defines the schema for the resource.
func (r *PlanetTrustedNetsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages the trusted networks list of the planet in Link11 WAAP.",
		Attributes: map[string]schema.Attribute{
			"config_id": schema.StringAttribute{
				Description: "The configuration ID.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"id": schema.StringAttribute{
				Description: "The planet entry ID (always __default__).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The planet entry name (always __default__).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
		Blocks: map[string]schema.Block{
			"trusted_nets": schema.ListNestedBlock{
				Description: "List of trusted network entries.",
				NestedObject: schema.NestedBlockObject{
					Attributes: map[string]schema.Attribute{
						"source": schema.StringAttribute{
							Description: "Source type: 'ip' or 'global_filter'.",
							Required:    true,
							Validators: []validator.String{
								stringvalidator.OneOf("ip", "global_filter"),
							},
						},
						"address": schema.StringAttribute{
							Description: "IP address or CIDR block. Required when source is 'ip'.",
							Optional:    true,
							Computed:    true,
							Default:     stringdefault.StaticString(""),
						},
						"gf_id": schema.StringAttribute{
							Description: "Global filter ID. Required when source is 'global_filter'.",
							Optional:    true,
							Computed:    true,
							Default:     stringdefault.StaticString(""),
						},
						"comment": schema.StringAttribute{
							Description: "Human-readable comment.",
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

// ConfigValidators returns the validators for cross-field validation.
func (r *PlanetTrustedNetsResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		&trustedNetsConfigValidator{},
	}
}

// trustedNetsConfigValidator validates cross-field constraints on trusted_nets entries.
type trustedNetsConfigValidator struct{}

func (v *trustedNetsConfigValidator) Description(_ context.Context) string {
	return "Validates trusted_nets entries: source=ip requires a valid address, source=global_filter requires gf_id."
}

func (v *trustedNetsConfigValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v *trustedNetsConfigValidator) ValidateResource(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config PlanetTrustedNetsResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	for i, entry := range config.TrustedNets {
		if entry.Source.IsUnknown() {
			continue
		}

		source := entry.Source.ValueString()
		address := entry.Address.ValueString()
		gfID := entry.GfID.ValueString()

		switch source {
		case "ip":
			if !entry.GfID.IsUnknown() && gfID != "" {
				resp.Diagnostics.AddAttributeError(
					path.Root("trusted_nets").AtListIndex(i).AtName("gf_id"),
					"Invalid Attribute Combination",
					"gf_id must be empty when source is 'ip'.",
				)
			}
			if !entry.Address.IsUnknown() && address == "" {
				resp.Diagnostics.AddAttributeError(
					path.Root("trusted_nets").AtListIndex(i).AtName("address"),
					"Missing Required Attribute",
					"address is required when source is 'ip'.",
				)
			} else if !entry.Address.IsUnknown() && address != "" {
				if _, err := netip.ParsePrefix(address); err != nil {
					if net.ParseIP(address) == nil {
						resp.Diagnostics.AddAttributeError(
							path.Root("trusted_nets").AtListIndex(i).AtName("address"),
							"Invalid Address",
							fmt.Sprintf("address must be a valid IP address or CIDR block, got: %s", address),
						)
					}
				}
			}

		case "global_filter":
			if !entry.Address.IsUnknown() && address != "" {
				resp.Diagnostics.AddAttributeError(
					path.Root("trusted_nets").AtListIndex(i).AtName("address"),
					"Invalid Attribute Combination",
					"address must be empty when source is 'global_filter'.",
				)
			}
			if !entry.GfID.IsUnknown() && gfID == "" {
				resp.Diagnostics.AddAttributeError(
					path.Root("trusted_nets").AtListIndex(i).AtName("gf_id"),
					"Missing Required Attribute",
					"gf_id is required when source is 'global_filter'.",
				)
			}
		}
	}
}

// Configure configures the resource with the provider client.
func (r *PlanetTrustedNetsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = providerutil.ConfigureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates the planet trusted_nets resource by issuing a PUT.
func (r *PlanetTrustedNetsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PlanetTrustedNetsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	planet := buildPlanetBody(&plan)

	err := r.client.UpsertPlanet(ctx, plan.ConfigID.ValueString(), defaultPlanetEntryID, planet)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Planet Trusted Nets",
			"Could not create planet trusted nets: "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(defaultPlanetEntryID)
	plan.Name = types.StringValue(defaultPlanetEntryID)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read reads the planet trusted_nets resource from the API.
func (r *PlanetTrustedNetsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PlanetTrustedNetsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	planet, err := r.client.GetPlanet(ctx, state.ConfigID.ValueString(), defaultPlanetEntryID)
	if err != nil {
		if client.IsNotFoundError(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(
			"Error Reading Planet Trusted Nets",
			"Could not read planet trusted nets: "+err.Error(),
		)
		return
	}

	state.ID = types.StringValue(planet.ID)
	state.Name = types.StringValue(planet.Name)

	state.TrustedNets = make([]TrustedNetModel, len(planet.TrustedNets))
	for i, tn := range planet.TrustedNets {
		state.TrustedNets[i] = TrustedNetModel{
			Source:  types.StringValue(tn.Source),
			Address: types.StringValue(tn.Address),
			GfID:    types.StringValue(tn.GfID),
			Comment: types.StringValue(tn.Comment),
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the planet trusted_nets resource by issuing a PUT.
func (r *PlanetTrustedNetsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan PlanetTrustedNetsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	planet := buildPlanetBody(&plan)

	err := r.client.UpsertPlanet(ctx, plan.ConfigID.ValueString(), defaultPlanetEntryID, planet)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Planet Trusted Nets",
			"Could not update planet trusted nets: "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(defaultPlanetEntryID)
	plan.Name = types.StringValue(defaultPlanetEntryID)

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete is a no-op because there is no DELETE endpoint for planets.
// It simply removes the resource from Terraform state.
func (r *PlanetTrustedNetsResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}

// ImportState imports an existing planet trusted_nets resource.
// The import ID may be either:
//   - "<config_id>"
//   - "<config_id>/__default__"
func (r *PlanetTrustedNetsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	configID := parts[0]
	if len(parts) == 2 && parts[1] != defaultPlanetEntryID {
		resp.Diagnostics.AddError(
			"Invalid Import ID",
			fmt.Sprintf("Import ID format must be <config_id> or <config_id>/%s, got: %s", defaultPlanetEntryID, req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("config_id"), configID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), defaultPlanetEntryID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), defaultPlanetEntryID)...)
}

// buildPlanetBody assembles the full Planet struct with user-provided trusted_nets
// and hardcoded defaults for all other fields.
func buildPlanetBody(plan *PlanetTrustedNetsResourceModel) *client.Planet {
	trustedNets := make([]client.TrustedNet, len(plan.TrustedNets))
	for i, tn := range plan.TrustedNets {
		trustedNets[i] = client.TrustedNet{
			Source:  tn.Source.ValueString(),
			Comment: tn.Comment.ValueString(),
			Address: tn.Address.ValueString(),
			GfID:    tn.GfID.ValueString(),
		}
	}

	return &client.Planet{
		ID:                 defaultPlanetEntryID,
		Name:               defaultPlanetEntryID,
		TrustedNets:        trustedNets,
		Ichallenge:         defaultIchallengeJSON,
		NoHostCertName:     defaultNoHostCertName,
		NoHostSSLCiphers:   defaultNoHostSSLCiphers,
		NoHostSSLProtocols: defaultNoHostSSLProtocols,
	}
}
