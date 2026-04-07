package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// readResp is a lightweight wrapper to pass to stringSliceToList etc.
type readResp = resource.ReadResponse

// metadataReq builds a resource.MetadataRequest with the given provider type name.
func metadataReq(providerTypeName string) resource.MetadataRequest {
	return resource.MetadataRequest{ProviderTypeName: providerTypeName}
}

// metadataResp builds an empty resource.MetadataResponse.
func metadataResp() *resource.MetadataResponse {
	return &resource.MetadataResponse{}
}

// schemaReq builds an empty resource.SchemaRequest.
func schemaReq() resource.SchemaRequest {
	return resource.SchemaRequest{}
}

// schemaResp builds an empty resource.SchemaResponse.
func schemaResp() *resource.SchemaResponse {
	return &resource.SchemaResponse{}
}

// configureReq builds a resource.ConfigureRequest with the given provider data.
func configureReq(providerData any) resource.ConfigureRequest {
	return resource.ConfigureRequest{ProviderData: providerData}
}

// configureResp builds an empty resource.ConfigureResponse.
func configureResp() *resource.ConfigureResponse {
	return &resource.ConfigureResponse{}
}

// testImportState is a helper that tests ImportState on a resource with the given ID.
// It creates the minimal required state and invokes ImportState.
func testImportState(t *testing.T, r resource.ResourceWithImportState, importID string) *resource.ImportStateResponse {
	t.Helper()
	ctx := context.Background()

	// Get the schema from the resource first
	schemaResource, ok := r.(resource.Resource)
	if !ok {
		t.Fatal("resource does not implement resource.Resource")
	}
	sReq := schemaReq()
	sResp := schemaResp()
	schemaResource.Schema(ctx, sReq, sResp)

	// Create an empty state with the schema
	state := tfsdk.State{
		Schema: sResp.Schema,
		Raw:    tftypes.NewValue(sResp.Schema.Type().TerraformType(ctx), nil),
	}

	req := resource.ImportStateRequest{ID: importID}
	resp := &resource.ImportStateResponse{
		State: state,
	}

	r.ImportState(ctx, req, resp)
	return resp
}
