package resources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/link11/terraform-provider-link11waap/internal/client"
	"github.com/link11/terraform-provider-link11waap/internal/providerutil"
)

var _ resource.Resource = &PublishResource{}

// PublishResource implements the publish resource.
type PublishResource struct {
	client *client.Client
}

// PublishResourceModel describes the publish resource data model.
type PublishResourceModel struct {
	ConfigID types.String `tfsdk:"config_id"`
	Buckets  types.List   `tfsdk:"buckets"`
	Triggers types.Map    `tfsdk:"triggers"`
	ID       types.String `tfsdk:"id"`
}

// BucketModel describes the data model for a publish target bucket.
type BucketModel struct {
	Name types.String `tfsdk:"name"`
	URL  types.String `tfsdk:"url"`
}

// NewPublishResource creates a new publish resource instance.
func NewPublishResource() resource.Resource {
	return &PublishResource{}
}

// Metadata returns the resource type name.
func (r *PublishResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_publish"
}

// Schema defines the schema for the publish resource.
func (r *PublishResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Publishes configuration changes in Link11 WAAP.",
		Attributes: map[string]schema.Attribute{
			"config_id": schema.StringAttribute{
				Description: "The configuration ID to publish.",
				Required:    true,
			},
			"buckets": schema.ListNestedAttribute{
				Description: "Target buckets for publishing. If not specified, uses default bucket.",
				Optional:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Description: "Bucket name.",
							Required:    true,
						},
						"url": schema.StringAttribute{
							Description: "Bucket URL.",
							Required:    true,
						},
					},
				},
			},
			"triggers": schema.MapAttribute{
				Description: "Map of arbitrary strings that, when changed, will trigger republish.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"id": schema.StringAttribute{
				Description: "Resource identifier.",
				Computed:    true,
			},
		},
	}
}

// Configure configures the resource with the provider client.
func (r *PublishResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = providerutil.ConfigureClient(req.ProviderData, &resp.Diagnostics)
}

// Create creates a new publish resource.
func (r *PublishResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PublishResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	buckets := r.getBuckets(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Publish(ctx, plan.ConfigID.ValueString(), buckets)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Publishing Configuration",
			"Could not publish configuration: "+err.Error(),
		)
		return
	}

	plan.ID = plan.ConfigID
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Read reads the publish resource.
func (r *PublishResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PublishResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Publish is a "fire and forget" resource - nothing to read
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update updates the publish resource.
func (r *PublishResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan PublishResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	buckets := r.getBuckets(ctx, &plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Publish(ctx, plan.ConfigID.ValueString(), buckets)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Publishing Configuration",
			"Could not publish configuration: "+err.Error(),
		)
		return
	}

	plan.ID = plan.ConfigID
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// Delete deletes the publish resource.
func (r *PublishResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Nothing to delete - publish is just an action
}

func (r *PublishResource) getBuckets(ctx context.Context, model *PublishResourceModel, diags *diag.Diagnostics) []client.PublishBucket {
	if model.Buckets.IsNull() || model.Buckets.IsUnknown() {
		return []client.PublishBucket{}
	}

	var bucketModels []BucketModel
	diags.Append(model.Buckets.ElementsAs(ctx, &bucketModels, false)...)
	if diags.HasError() {
		return nil
	}

	buckets := make([]client.PublishBucket, len(bucketModels))
	for i, b := range bucketModels {
		buckets[i] = client.PublishBucket{
			Name: b.Name.ValueString(),
			URL:  b.URL.ValueString(),
		}
	}
	return buckets
}
