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

func TestDynamicRulesDataSource_Metadata(t *testing.T) {
	d := NewDynamicRulesDataSource()
	req := dsMetadataReq("link11waap")
	resp := dsMetadataResp()
	d.Metadata(context.Background(), req, resp)
	assert.Equal(t, "link11waap_dynamic_rules", resp.TypeName)
}

func TestDynamicRulesDataSource_Schema(t *testing.T) {
	d := NewDynamicRulesDataSource()
	req := dsSchemaReq()
	resp := dsSchemaResp()
	d.Schema(context.Background(), req, resp)
	require.NotNil(t, resp.Schema)
	assert.Contains(t, resp.Schema.Attributes, "config_id")
	assert.Contains(t, resp.Schema.Attributes, "dynamic_rules")
}

func TestDynamicRulesDataSource_Configure_InvalidType(t *testing.T) {
	testDSConfigureWithInvalidType(t, NewDynamicRulesDataSource())
}

func TestDynamicRulesDataSource_Configure_Nil(t *testing.T) {
	testDSConfigureWithNil(t, NewDynamicRulesDataSource())
}

func TestDynamicRulesDataSource_Read_Success(t *testing.T) {
	d := NewDynamicRulesDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ListResponse[client.DynamicRule]{
			Total: 1,
			Items: []client.DynamicRule{
				{
					ID:                 "dr1",
					Name:               "test-rule",
					Description:        "Test dynamic rule",
					Threshold:          100,
					Timeframe:          60,
					TTL:                300,
					Active:             true,
					OffloadIPFiltering: false,
					Target:             "ip",
					Action:             "action-monitor",
					Tags:               []string{"tag1", "tag2"},
					Include: client.RateLimitTagFilter{
						Relation: "OR",
						Tags:     []string{"include-tag"},
					},
					Exclude: client.RateLimitTagFilter{
						Relation: "AND",
						Tags:     []string{"exclude-tag"},
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

func TestDynamicRulesDataSource_Read_NilTags(t *testing.T) {
	d := NewDynamicRulesDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ListResponse[client.DynamicRule]{
			Total: 1,
			Items: []client.DynamicRule{
				{
					ID:        "dr1",
					Name:      "test-rule",
					Active:    true,
					Threshold: 100,
					Timeframe: 60,
					TTL:       300,
					Target:    "ip",
					Action:    "action-monitor",
					Tags:      nil,
					Include:   client.RateLimitTagFilter{Relation: "OR", Tags: nil},
					Exclude:   client.RateLimitTagFilter{Relation: "AND", Tags: nil},
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

func TestDynamicRulesDataSource_Read_APIError(t *testing.T) {
	d := NewDynamicRulesDataSource()
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

func TestDynamicRulesDataSource_Read_MultipleRules(t *testing.T) {
	d := NewDynamicRulesDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ListResponse[client.DynamicRule]{
			Total: 2,
			Items: []client.DynamicRule{
				{
					ID: "dr1", Name: "rule1", Active: true, Timeframe: 60, Threshold: 100, TTL: 300, Target: "ip", Action: "ban",
					Tags: []string{"t1"}, Include: client.RateLimitTagFilter{Relation: "OR", Tags: []string{"i1"}},
					Exclude: client.RateLimitTagFilter{Relation: "AND", Tags: []string{"e1"}},
				},
				{
					ID: "dr2", Name: "rule2", Active: false, Timeframe: 120, Threshold: 50, TTL: 600, Target: "session", Action: "monitor",
					Include: client.RateLimitTagFilter{Relation: "AND"}, Exclude: client.RateLimitTagFilter{Relation: "OR"},
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
