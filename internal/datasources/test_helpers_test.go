package datasources

import (
	"context"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/require"
)

// dsMetadataReq builds a datasource.MetadataRequest with the given provider type name.
func dsMetadataReq(providerTypeName string) datasource.MetadataRequest {
	return datasource.MetadataRequest{ProviderTypeName: providerTypeName}
}

// dsMetadataResp builds an empty datasource.MetadataResponse.
func dsMetadataResp() *datasource.MetadataResponse {
	return &datasource.MetadataResponse{}
}

// dsSchemaReq builds an empty datasource.SchemaRequest.
func dsSchemaReq() datasource.SchemaRequest {
	return datasource.SchemaRequest{}
}

// dsSchemaResp builds an empty datasource.SchemaResponse.
func dsSchemaResp() *datasource.SchemaResponse {
	return &datasource.SchemaResponse{}
}

// dsConfigureReq builds a datasource.ConfigureRequest with the given provider data.
func dsConfigureReq(providerData any) datasource.ConfigureRequest {
	return datasource.ConfigureRequest{ProviderData: providerData}
}

// dsConfigureResp builds an empty datasource.ConfigureResponse.
func dsConfigureResp() *datasource.ConfigureResponse {
	return &datasource.ConfigureResponse{}
}

// configureDatasourceWithMock configures a datasource with a mock client backed by a TLS test server.
func configureDatasourceWithMock(t *testing.T, d datasource.DataSource, handler http.Handler) {
	t.Helper()
	ctx := context.Background()
	c, _ := newMockClient(t, handler)

	req := datasource.ConfigureRequest{ProviderData: c}
	resp := &datasource.ConfigureResponse{}
	if cr, ok := d.(datasource.DataSourceWithConfigure); ok {
		cr.Configure(ctx, req, resp)
	}
	require.False(t, resp.Diagnostics.HasError(), "configure errors: %v", resp.Diagnostics)
}

// buildDatasourceConfig creates a tfsdk.Config from a datasource schema and tftypes values.
func buildDatasourceConfig(ctx context.Context, t *testing.T, d datasource.DataSource, values map[string]tftypes.Value) tfsdk.Config {
	t.Helper()

	sReq := dsSchemaReq()
	sResp := dsSchemaResp()
	d.Schema(ctx, sReq, sResp)

	tfType := sResp.Schema.Type().TerraformType(ctx)
	objType, ok := tfType.(tftypes.Object)
	if !ok {
		t.Fatalf("expected tftypes.Object, got %T", tfType)
	}

	// Fill in missing attributes with null values
	fullValues := make(map[string]tftypes.Value)
	for attrName, attrType := range objType.AttributeTypes {
		fullValues[attrName] = tftypes.NewValue(attrType, nil)
	}
	for k, v := range values {
		fullValues[k] = v
	}

	raw := tftypes.NewValue(tfType, fullValues)

	return tfsdk.Config{
		Schema: sResp.Schema,
		Raw:    raw,
	}
}

// readDatasource performs a full Read operation on a datasource using a mock API server.
func readDatasource(t *testing.T, d datasource.DataSource, configValues map[string]tftypes.Value) *datasource.ReadResponse {
	t.Helper()
	ctx := context.Background()

	config := buildDatasourceConfig(ctx, t, d, configValues)

	sReq := dsSchemaReq()
	sResp := dsSchemaResp()
	d.Schema(ctx, sReq, sResp)

	readResp := &datasource.ReadResponse{
		State: tfsdk.State{
			Schema: sResp.Schema,
			Raw:    tftypes.NewValue(sResp.Schema.Type().TerraformType(ctx), nil),
		},
	}
	d.Read(ctx, datasource.ReadRequest{Config: config}, readResp)
	return readResp
}

// testDSConfigureWithInvalidType tests Configure with a wrong provider data type.
func testDSConfigureWithInvalidType(t *testing.T, d datasource.DataSource) {
	t.Helper()
	ctx := context.Background()

	cr, ok := d.(datasource.DataSourceWithConfigure)
	require.True(t, ok, "datasource does not implement DataSourceWithConfigure")

	req := dsConfigureReq("not-a-client") // string instead of *client.Client
	resp := dsConfigureResp()
	cr.Configure(ctx, req, resp)

	require.True(t, resp.Diagnostics.HasError(), "expected error for invalid provider data type")
}

// testDSConfigureWithNil tests Configure with nil provider data.
func testDSConfigureWithNil(t *testing.T, d datasource.DataSource) {
	t.Helper()
	ctx := context.Background()

	cr, ok := d.(datasource.DataSourceWithConfigure)
	require.True(t, ok, "datasource does not implement DataSourceWithConfigure")

	req := dsConfigureReq(nil)
	resp := dsConfigureResp()
	cr.Configure(ctx, req, resp)

	require.False(t, resp.Diagnostics.HasError(), "nil provider data should not produce error")
}
