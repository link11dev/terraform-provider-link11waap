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

func sampleCfProfile(id, name string) client.ContentFilterProfile {
	return client.ContentFilterProfile{
		ID:             id,
		Name:           name,
		Description:    "desc",
		IgnoreAlphanum: true,
		MaskingSeed:    "seed",
		ContentType:    []string{"application/json"},
		GraphqlPath:    "$.q",
		Tags:           []string{"t1"},
		Action:         "act1",
		Args: client.ContentFilterProfileSection{
			MaxCount: 10, MaxLength: 100, EnableMaxCount: true,
			Names: []client.ContentFilterEntryMatch{{Key: "k", Mask: true, Active: true, IgnoreCFRuleTags: []string{"skip"}}},
		},
		Headers:     client.ContentFilterProfileSection{},
		Cookies:     client.ContentFilterProfileSection{},
		Path:        client.ContentFilterProfileSection{},
		URL:         client.ContentFilterProfileSection{},
		AllSections: client.ContentFilterProfileSection{},
		Decoding:    client.ContentFilterDecoding{Base64: true},
	}
}

func TestContentFilterProfilesDataSource_Metadata(t *testing.T) {
	d := NewContentFilterProfilesDataSource()
	req := dsMetadataReq("link11waap")
	resp := dsMetadataResp()
	d.Metadata(context.Background(), req, resp)
	assert.Equal(t, "link11waap_content_filter_profiles", resp.TypeName)
}

func TestContentFilterProfilesDataSource_Schema(t *testing.T) {
	d := NewContentFilterProfilesDataSource()
	req := dsSchemaReq()
	resp := dsSchemaResp()
	d.Schema(context.Background(), req, resp)
	require.NotNil(t, resp.Schema)
	assert.Contains(t, resp.Schema.Attributes, "config_id")
	assert.Contains(t, resp.Schema.Attributes, "id")
	assert.Contains(t, resp.Schema.Attributes, "name")
	assert.Contains(t, resp.Schema.Attributes, "content_filter_profiles")
}

func TestContentFilterProfilesDataSource_Configure_InvalidType(t *testing.T) {
	testDSConfigureWithInvalidType(t, NewContentFilterProfilesDataSource())
}

func TestContentFilterProfilesDataSource_Configure_Nil(t *testing.T) {
	testDSConfigureWithNil(t, NewContentFilterProfilesDataSource())
}

func TestContentFilterProfilesDataSource_Read_Success(t *testing.T) {
	d := NewContentFilterProfilesDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ListResponse[client.ContentFilterProfile]{
			Total: 2,
			Items: []client.ContentFilterProfile{
				sampleCfProfile("cfp1", "profile1"),
				sampleCfProfile("cfp2", "profile2"),
			},
		})
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
	})
	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}

func TestContentFilterProfilesDataSource_Read_Empty(t *testing.T) {
	d := NewContentFilterProfilesDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ListResponse[client.ContentFilterProfile]{
			Total: 0,
			Items: []client.ContentFilterProfile{},
		})
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
	})
	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}

func TestContentFilterProfilesDataSource_Read_APIError(t *testing.T) {
	d := NewContentFilterProfilesDataSource()
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

func TestContentFilterProfilesDataSource_Read_ByName(t *testing.T) {
	d := NewContentFilterProfilesDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ListResponse[client.ContentFilterProfile]{
			Total: 2,
			Items: []client.ContentFilterProfile{
				sampleCfProfile("cfp1", "profile1"),
				sampleCfProfile("cfp2", "profile2"),
			},
		})
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"name":      tftypes.NewValue(tftypes.String, "profile2"),
	})
	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}

func TestContentFilterProfilesDataSource_Read_ByName_NotFound(t *testing.T) {
	d := NewContentFilterProfilesDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ListResponse[client.ContentFilterProfile]{
			Total: 1,
			Items: []client.ContentFilterProfile{sampleCfProfile("cfp1", "profile1")},
		})
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"name":      tftypes.NewValue(tftypes.String, "missing"),
	})
	assert.True(t, resp.Diagnostics.HasError())
}

func TestContentFilterProfilesDataSource_Read_ByID(t *testing.T) {
	d := NewContentFilterProfilesDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v4.3/conf/cfg1/content-filter-profiles/cfp1", r.URL.Path)
		json.NewEncoder(w).Encode(sampleCfProfile("cfp1", "profile1"))
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"id":        tftypes.NewValue(tftypes.String, "cfp1"),
	})
	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}

func TestContentFilterProfilesDataSource_Read_ByID_APIError(t *testing.T) {
	d := NewContentFilterProfilesDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"error"}`))
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"id":        tftypes.NewValue(tftypes.String, "missing"),
	})
	assert.True(t, resp.Diagnostics.HasError())
}
