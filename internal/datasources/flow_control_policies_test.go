package datasources

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/link11/terraform-provider-link11waap/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func fcStrPtr(s string) *string { return &s }

func TestFlowControlPoliciesDataSource_Metadata(t *testing.T) {
	d := NewFlowControlPoliciesDataSource()
	resp := dsMetadataResp()
	d.Metadata(context.Background(), dsMetadataReq("link11waap"), resp)
	assert.Equal(t, "link11waap_flow_control_policies", resp.TypeName)
}

func TestFlowControlPoliciesDataSource_Schema(t *testing.T) {
	d := NewFlowControlPoliciesDataSource()
	resp := dsSchemaResp()
	d.Schema(context.Background(), dsSchemaReq(), resp)
	require.NotNil(t, resp.Schema)
	assert.Contains(t, resp.Schema.Attributes, "config_id")
	assert.Contains(t, resp.Schema.Attributes, "flow_control_policies")
}

func TestFlowControlPoliciesDataSource_Configure_InvalidType(t *testing.T) {
	testDSConfigureWithInvalidType(t, NewFlowControlPoliciesDataSource())
}

func TestFlowControlPoliciesDataSource_Configure_Nil(t *testing.T) {
	testDSConfigureWithNil(t, NewFlowControlPoliciesDataSource())
}

func TestFlowControlPoliciesDataSource_Read_Success(t *testing.T) {
	d := NewFlowControlPoliciesDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ListResponse[client.FlowControl]{
			Total: 1,
			Items: []client.FlowControl{
				{
					ID:          "fc1",
					Name:        "flow1",
					Description: "Test flow",
					Active:      true,
					Timeframe:   60,
					Tags:        []string{"t1"},
					Include:     []string{"in1"},
					Exclude:     []string{"ex1"},
					Key: []client.FlowControlKeyEntry{
						{Attrs: fcStrPtr("ip")},
					},
					Steps: []client.FlowStepItem{
						{Method: "GET", URI: "/login", Headers: map[string]string{"X-Test": "1"}},
						{Method: "POST", URI: "/checkout"},
					},
				},
			},
		})
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
	})
	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}

func TestFlowControlPoliciesDataSource_Read_NilCollections(t *testing.T) {
	d := NewFlowControlPoliciesDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ListResponse[client.FlowControl]{
			Total: 1,
			Items: []client.FlowControl{
				{
					ID:        "fc1",
					Name:      "flow1",
					Active:    false,
					Timeframe: 30,
					Tags:      nil,
					Include:   nil,
					Exclude:   nil,
					Key:       []client.FlowControlKeyEntry{{Cookies: fcStrPtr("sid")}},
					Steps:     []client.FlowStepItem{{Method: "GET", URI: "/"}},
				},
			},
		})
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
	})
	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}

func TestFlowControlPoliciesDataSource_Read_APIError(t *testing.T) {
	d := NewFlowControlPoliciesDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"error"}`))
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
	})
	assert.True(t, resp.Diagnostics.HasError())
}
