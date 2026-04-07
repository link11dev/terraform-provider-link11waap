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

func TestRateLimitRulesDataSource_Metadata(t *testing.T) {
	d := NewRateLimitRulesDataSource()
	req := dsMetadataReq("link11waap")
	resp := dsMetadataResp()
	d.Metadata(context.Background(), req, resp)
	assert.Equal(t, "link11waap_rate_limit_rules", resp.TypeName)
}

func TestRateLimitRulesDataSource_Schema(t *testing.T) {
	d := NewRateLimitRulesDataSource()
	req := dsSchemaReq()
	resp := dsSchemaResp()
	d.Schema(context.Background(), req, resp)
	require.NotNil(t, resp.Schema)
	assert.Contains(t, resp.Schema.Attributes, "config_id")
	assert.Contains(t, resp.Schema.Attributes, "rate_limit_rules")
}

func TestRateLimitRulesDataSource_Configure_InvalidType(t *testing.T) {
	testDSConfigureWithInvalidType(t, NewRateLimitRulesDataSource())
}

func TestRateLimitRulesDataSource_Configure_Nil(t *testing.T) {
	testDSConfigureWithNil(t, NewRateLimitRulesDataSource())
}

func TestRateLimitRulesDataSource_Read_Success(t *testing.T) {
	d := NewRateLimitRulesDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ListResponse[client.RateLimitRule]{
			Total: 1,
			Items: []client.RateLimitRule{
				{
					ID:          "rl1",
					Name:        "test-rule",
					Description: "Test rate limit",
					Global:      false,
					Active:      true,
					Timeframe:   60,
					Threshold:   100,
					TTL:         300,
					Action:      "action-monitor",
					Tags:        []string{"tag1", "tag2"},
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

func TestRateLimitRulesDataSource_Read_NilTags(t *testing.T) {
	d := NewRateLimitRulesDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ListResponse[client.RateLimitRule]{
			Total: 1,
			Items: []client.RateLimitRule{
				{
					ID:        "rl1",
					Name:      "test-rule",
					Active:    true,
					Timeframe: 60,
					Threshold: 100,
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

func TestRateLimitRulesDataSource_Read_APIError(t *testing.T) {
	d := NewRateLimitRulesDataSource()
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

func TestRateLimitRulesDataSource_Read_MultipleRules(t *testing.T) {
	d := NewRateLimitRulesDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ListResponse[client.RateLimitRule]{
			Total: 2,
			Items: []client.RateLimitRule{
				{
					ID: "rl1", Name: "rule1", Active: true, Timeframe: 60, Threshold: 100, Action: "ban",
					Tags: []string{"t1"}, Include: client.RateLimitTagFilter{Relation: "OR", Tags: []string{"i1"}},
					Exclude: client.RateLimitTagFilter{Relation: "AND", Tags: []string{"e1"}},
				},
				{
					ID: "rl2", Name: "rule2", Active: false, Timeframe: 120, Threshold: 50, Action: "monitor",
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
