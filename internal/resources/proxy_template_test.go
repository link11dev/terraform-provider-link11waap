package resources

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/link11/terraform-provider-link11waap/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProxyTemplateResource(t *testing.T) {
	r := NewProxyTemplateResource()
	require.NotNil(t, r)
	_, ok := r.(*ProxyTemplateResource)
	assert.True(t, ok)
}

func TestProxyTemplateResource_Metadata(t *testing.T) {
	r := &ProxyTemplateResource{}
	ctx := context.Background()

	req := metadataReq("link11waap")
	resp := metadataResp()
	r.Metadata(ctx, req, resp)

	assert.Equal(t, "link11waap_proxy_template", resp.TypeName)
}

func TestProxyTemplateResource_Schema(t *testing.T) {
	r := &ProxyTemplateResource{}
	ctx := context.Background()

	req := schemaReq()
	resp := schemaResp()
	r.Schema(ctx, req, resp)

	schema := resp.Schema
	assert.NotEmpty(t, schema.Attributes)

	expectedAttrs := []string{
		"config_id", "id", "name", "description",
		"acao_header", "xff_header_name", "xrealip_header_name",
		"proxy_connect_timeout", "proxy_read_timeout", "proxy_send_timeout",
		"upstream_host",
		"client_body_timeout", "client_body_buffer_size",
		"client_header_timeout", "client_header_buffer_size",
		"client_max_body_size", "keepalive_timeout", "send_timeout",
		"limit_req_rate", "limit_req_burst", "mask_headers",
		"custom_listener",
		"large_client_header_buffers_count", "large_client_header_buffers_size",
		"conf_specific", "ssl_conf_specific", "ssl_ciphers",
		"ssl_protocols", "advanced_configuration",
	}
	for _, attr := range expectedAttrs {
		_, ok := schema.Attributes[attr]
		assert.True(t, ok, "expected attribute %q in schema", attr)
	}
}

func TestProxyTemplateResource_Configure_NilProvider(t *testing.T) {
	r := &ProxyTemplateResource{}
	ctx := context.Background()

	req := configureReq(nil)
	resp := configureResp()
	r.Configure(ctx, req, resp)

	assert.Nil(t, r.client)
	assert.False(t, resp.Diagnostics.HasError())
}

func TestProxyTemplateResource_ImportState_Valid(t *testing.T) {
	r := &ProxyTemplateResource{}
	resp := testImportState(t, r, "config123/pt456")

	assert.False(t, resp.Diagnostics.HasError())
}

func TestProxyTemplateResource_ImportState_Invalid(t *testing.T) {
	r := &ProxyTemplateResource{}
	resp := testImportState(t, r, "invalidformat")

	assert.True(t, resp.Diagnostics.HasError())
}

func TestProxyTemplateResource_ImportState_TooManyParts(t *testing.T) {
	r := &ProxyTemplateResource{}
	resp := testImportState(t, r, "a/b/c")

	assert.True(t, resp.Diagnostics.HasError())
}

// --- CRUD with failing client ---

func TestProxyTemplateResource_CRUD_WithFailingClient(t *testing.T) {
	r := &ProxyTemplateResource{}

	planVals := map[string]tftypes.Value{
		"config_id":                         tftypes.NewValue(tftypes.String, "cfg1"),
		"id":                                tftypes.NewValue(tftypes.String, nil),
		"name":                              tftypes.NewValue(tftypes.String, "test-pt"),
		"description":                       tftypes.NewValue(tftypes.String, ""),
		"acao_header":                       tftypes.NewValue(tftypes.Bool, false),
		"xff_header_name": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
			tftypes.NewValue(tftypes.String, "X-Forwarded-For"),
		}),
		"xrealip_header_name":               tftypes.NewValue(tftypes.String, "X-Real-IP"),
		"proxy_connect_timeout":             tftypes.NewValue(tftypes.String, "5"),
		"proxy_read_timeout":                tftypes.NewValue(tftypes.String, "60"),
		"proxy_send_timeout":                tftypes.NewValue(tftypes.String, "30"),
		"upstream_host":                     tftypes.NewValue(tftypes.String, "$host"),
		"client_body_timeout":               tftypes.NewValue(tftypes.String, "5"),
		"client_body_buffer_size":           tftypes.NewValue(tftypes.String, "4"),
		"client_header_timeout":             tftypes.NewValue(tftypes.String, "5"),
		"client_header_buffer_size":         tftypes.NewValue(tftypes.String, "32"),
		"client_max_body_size":              tftypes.NewValue(tftypes.String, "150"),
		"keepalive_timeout":                 tftypes.NewValue(tftypes.String, "660"),
		"send_timeout":                      tftypes.NewValue(tftypes.String, "5"),
		"limit_req_rate":                    tftypes.NewValue(tftypes.String, "1200"),
		"limit_req_burst":                   tftypes.NewValue(tftypes.String, "400"),
		"mask_headers":                      tftypes.NewValue(tftypes.String, ""),
		"custom_listener":                   tftypes.NewValue(tftypes.Bool, false),
		"large_client_header_buffers_count": tftypes.NewValue(tftypes.String, "2"),
		"large_client_header_buffers_size":  tftypes.NewValue(tftypes.String, "32"),
		"conf_specific":                     tftypes.NewValue(tftypes.String, ""),
		"ssl_conf_specific":                 tftypes.NewValue(tftypes.String, ""),
		"ssl_ciphers":                       tftypes.NewValue(tftypes.String, "ECDHE-RSA-AES256-GCM-SHA384"),
		"ssl_protocols": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
			tftypes.NewValue(tftypes.String, "TLSv1.2"),
		}),
		"advanced_configuration": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"name":          tftypes.String,
				"protocol":      tftypes.List{ElementType: tftypes.String},
				"configuration": tftypes.String,
				"description":   tftypes.String,
			},
		}}, nil),
	}

	stateVals := map[string]tftypes.Value{
		"config_id":                         tftypes.NewValue(tftypes.String, "cfg1"),
		"id":                                tftypes.NewValue(tftypes.String, "pt1"),
		"name":                              tftypes.NewValue(tftypes.String, "test-pt"),
		"description":                       tftypes.NewValue(tftypes.String, ""),
		"acao_header":                       tftypes.NewValue(tftypes.Bool, false),
		"xff_header_name": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
			tftypes.NewValue(tftypes.String, "X-Forwarded-For"),
		}),
		"xrealip_header_name":               tftypes.NewValue(tftypes.String, "X-Real-IP"),
		"proxy_connect_timeout":             tftypes.NewValue(tftypes.String, "5"),
		"proxy_read_timeout":                tftypes.NewValue(tftypes.String, "60"),
		"proxy_send_timeout":                tftypes.NewValue(tftypes.String, "30"),
		"upstream_host":                     tftypes.NewValue(tftypes.String, "$host"),
		"client_body_timeout":               tftypes.NewValue(tftypes.String, "5"),
		"client_body_buffer_size":           tftypes.NewValue(tftypes.String, "4"),
		"client_header_timeout":             tftypes.NewValue(tftypes.String, "5"),
		"client_header_buffer_size":         tftypes.NewValue(tftypes.String, "32"),
		"client_max_body_size":              tftypes.NewValue(tftypes.String, "150"),
		"keepalive_timeout":                 tftypes.NewValue(tftypes.String, "660"),
		"send_timeout":                      tftypes.NewValue(tftypes.String, "5"),
		"limit_req_rate":                    tftypes.NewValue(tftypes.String, "1200"),
		"limit_req_burst":                   tftypes.NewValue(tftypes.String, "400"),
		"mask_headers":                      tftypes.NewValue(tftypes.String, ""),
		"custom_listener":                   tftypes.NewValue(tftypes.Bool, false),
		"large_client_header_buffers_count": tftypes.NewValue(tftypes.String, "2"),
		"large_client_header_buffers_size":  tftypes.NewValue(tftypes.String, "32"),
		"conf_specific":                     tftypes.NewValue(tftypes.String, ""),
		"ssl_conf_specific":                 tftypes.NewValue(tftypes.String, ""),
		"ssl_ciphers":                       tftypes.NewValue(tftypes.String, "ECDHE-RSA-AES256-GCM-SHA384"),
		"ssl_protocols": tftypes.NewValue(tftypes.List{ElementType: tftypes.String}, []tftypes.Value{
			tftypes.NewValue(tftypes.String, "TLSv1.2"),
		}),
		"advanced_configuration": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{
			AttributeTypes: map[string]tftypes.Type{
				"name":          tftypes.String,
				"protocol":      tftypes.List{ElementType: tftypes.String},
				"configuration": tftypes.String,
				"description":   tftypes.String,
			},
		}}, nil),
	}

	t.Run("Create", func(t *testing.T) { crudCreateWithClient(t, r, planVals) })
	t.Run("Read", func(t *testing.T) { crudReadWithClient(t, r, stateVals) })
	t.Run("Update", func(t *testing.T) { crudUpdateWithClient(t, r, planVals) })
	t.Run("Delete", func(t *testing.T) { crudDeleteWithClient(t, r, stateVals) })
}

// --- Read with mock ---

func TestProxyTemplateResource_Read_WithMock(t *testing.T) {
	r := &ProxyTemplateResource{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ProxyTemplate{
			ID:                            "pt1",
			Name:                          "test-pt",
			Description:                   "A proxy template",
			ACAOHeader:                    true,
			XFFHeaderName:                 []string{"X-Forwarded-For"},
			XRealIPHeaderName:             "X-Real-IP",
			ProxyConnectTimeout:           "5",
			ProxyReadTimeout:              "60",
			ProxySendTimeout:              "30",
			UpstreamHost:                  "$host",
			ClientBodyTimeout:             "5",
			ClientBodyBufferSize:          "4",
			ClientHeaderTimeout:           "5",
			ClientHeaderBufferSize:        "32",
			ClientMaxBodySize:             "150",
			KeepaliveTimeout:              "660",
			SendTimeout:                   "5",
			LimitReqRate:                  "1200",
			LimitReqBurst:                 "400",
			MaskHeaders:                   "",
			CustomListener:                false,
			LargeClientHeaderBuffersCount: "2",
			LargeClientHeaderBuffersSize:  "32",
			ConfSpecific:                  "",
			SSLConfSpecific:               "",
			SSLCiphers:                    "ECDHE-RSA-AES256-GCM-SHA384",
			SSLProtocols:                  []string{"TLSv1.2", "TLSv1.3"},
		})
	})
	configureResourceWithMock(t, r, handler)

	resp := readWithMock(t, r, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"id":        tftypes.NewValue(tftypes.String, "pt1"),
	})

	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}

func TestProxyTemplateResource_Read_WithMock_SSLProtocols(t *testing.T) {
	r := &ProxyTemplateResource{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ProxyTemplate{
			ID:                            "pt1",
			Name:                          "ssl-test",
			XFFHeaderName:                 []string{"X-Forwarded-For"},
			XRealIPHeaderName:             "X-Real-IP",
			ProxyConnectTimeout:           "5",
			ProxyReadTimeout:              "60",
			ProxySendTimeout:              "30",
			UpstreamHost:                  "$host",
			ClientBodyTimeout:             "5",
			ClientBodyBufferSize:          "4",
			ClientHeaderTimeout:           "5",
			ClientHeaderBufferSize:        "32",
			ClientMaxBodySize:             "150",
			KeepaliveTimeout:              "660",
			SendTimeout:                   "5",
			LimitReqRate:                  "1200",
			LimitReqBurst:                 "400",
			LargeClientHeaderBuffersCount: "2",
			LargeClientHeaderBuffersSize:  "32",
			SSLProtocols:                  []string{"TLSv1.1", "TLSv1.2", "TLSv1.3"},
		})
	})
	configureResourceWithMock(t, r, handler)

	resp := readWithMock(t, r, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"id":        tftypes.NewValue(tftypes.String, "pt1"),
	})

	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}

func TestProxyTemplateResource_Read_WithMock_AdvancedConfig(t *testing.T) {
	t.Run("WithProtocol", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			json.NewEncoder(w).Encode(client.ProxyTemplate{
				ID:                            "pt1",
				Name:                          "adv-test",
				XFFHeaderName:                 []string{"X-Forwarded-For"},
				XRealIPHeaderName:             "X-Real-IP",
				ProxyConnectTimeout:           "5",
				ProxyReadTimeout:              "60",
				ProxySendTimeout:              "30",
				UpstreamHost:                  "$host",
				ClientBodyTimeout:             "5",
				ClientBodyBufferSize:          "4",
				ClientHeaderTimeout:           "5",
				ClientHeaderBufferSize:        "32",
				ClientMaxBodySize:             "150",
				KeepaliveTimeout:              "660",
				SendTimeout:                   "5",
				LimitReqRate:                  "1200",
				LimitReqBurst:                 "400",
				LargeClientHeaderBuffersCount: "2",
				LargeClientHeaderBuffersSize:  "32",
				SSLProtocols:                  []string{"TLSv1.2"},
				AdvancedConfiguration: []client.ProxyTemplateAdvancedConfig{
					{
						Name:          "custom-block",
						Protocol:      []string{"http", "https"},
						Configuration: "proxy_set_header X-Custom true;",
						Description:   "Custom header",
					},
				},
			})
		})
		r2 := &ProxyTemplateResource{}
		configureResourceWithMock(t, r2, handler)

		resp := readWithMock(t, r2, map[string]tftypes.Value{
			"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
			"id":        tftypes.NewValue(tftypes.String, "pt1"),
		})

		assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
	})

	t.Run("WithNilProtocol", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			json.NewEncoder(w).Encode(client.ProxyTemplate{
				ID:                            "pt1",
				Name:                          "adv-test-nil-proto",
				XFFHeaderName:                 []string{"X-Forwarded-For"},
				XRealIPHeaderName:             "X-Real-IP",
				ProxyConnectTimeout:           "5",
				ProxyReadTimeout:              "60",
				ProxySendTimeout:              "30",
				UpstreamHost:                  "$host",
				ClientBodyTimeout:             "5",
				ClientBodyBufferSize:          "4",
				ClientHeaderTimeout:           "5",
				ClientHeaderBufferSize:        "32",
				ClientMaxBodySize:             "150",
				KeepaliveTimeout:              "660",
				SendTimeout:                   "5",
				LimitReqRate:                  "1200",
				LimitReqBurst:                 "400",
				LargeClientHeaderBuffersCount: "2",
				LargeClientHeaderBuffersSize:  "32",
				SSLProtocols:                  []string{"TLSv1.2"},
				AdvancedConfiguration: []client.ProxyTemplateAdvancedConfig{
					{
						Name:          "custom-block",
						Protocol:      nil,
						Configuration: "proxy_set_header X-Custom true;",
						Description:   "Custom header nil proto",
					},
				},
			})
		})
		r3 := &ProxyTemplateResource{}
		configureResourceWithMock(t, r3, handler)

		resp := readWithMock(t, r3, map[string]tftypes.Value{
			"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
			"id":        tftypes.NewValue(tftypes.String, "pt1"),
		})

		assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
	})
}

func TestProxyTemplateResource_Read_NotFound(t *testing.T) {
	r := &ProxyTemplateResource{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"not found"}`))
	})
	configureResourceWithMock(t, r, handler)

	resp := readWithMock(t, r, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"id":        tftypes.NewValue(tftypes.String, "missing"),
	})

	assert.False(t, resp.Diagnostics.HasError(), "not found should not produce error, resource should be removed")
}

func TestProxyTemplateResource_Read_APIError(t *testing.T) {
	r := &ProxyTemplateResource{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"server error"}`))
	})
	configureResourceWithMock(t, r, handler)

	resp := readWithMock(t, r, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
		"id":        tftypes.NewValue(tftypes.String, "pt1"),
	})

	assert.True(t, resp.Diagnostics.HasError())
}

// --- Helper function tests ---

func TestAdvancedConfigAttrTypes(t *testing.T) {
	attrTypes := advancedConfigAttrTypes()
	require.NotNil(t, attrTypes)
	assert.Len(t, attrTypes, 4)

	assert.Equal(t, types.StringType, attrTypes["name"])
	assert.Equal(t, types.ListType{ElemType: types.StringType}, attrTypes["protocol"])
	assert.Equal(t, types.StringType, attrTypes["configuration"])
	assert.Equal(t, types.StringType, attrTypes["description"])
}

func TestProxyTemplateResourceModel_FieldTypes(t *testing.T) {
	model := ProxyTemplateResourceModel{
		ConfigID:                      types.StringValue("cfg1"),
		ID:                            types.StringValue("pt1"),
		Name:                          types.StringValue("test-pt"),
		Description:                   types.StringValue("desc"),
		ACAOHeader:                    types.BoolValue(true),
		XFFHeaderName:                 types.ListValueMust(types.StringType, []attr.Value{types.StringValue("X-Forwarded-For")}),
		XRealIPHeaderName:             types.StringValue("X-Real-IP"),
		ProxyConnectTimeout:           types.StringValue("5"),
		ProxyReadTimeout:              types.StringValue("60"),
		ProxySendTimeout:              types.StringValue("30"),
		UpstreamHost:                  types.StringValue("$host"),
		ClientBodyTimeout:             types.StringValue("5"),
		ClientBodyBufferSize:          types.StringValue("4"),
		ClientHeaderTimeout:           types.StringValue("5"),
		ClientHeaderBufferSize:        types.StringValue("32"),
		ClientMaxBodySize:             types.StringValue("150"),
		KeepaliveTimeout:              types.StringValue("660"),
		SendTimeout:                   types.StringValue("5"),
		LimitReqRate:                  types.StringValue("1200"),
		LimitReqBurst:                 types.StringValue("400"),
		MaskHeaders:                   types.StringValue(""),
		CustomListener:                types.BoolValue(false),
		LargeClientHeaderBuffersCount: types.StringValue("2"),
		LargeClientHeaderBuffersSize:  types.StringValue("32"),
		ConfSpecific:                  types.StringValue(""),
		SSLConfSpecific:               types.StringValue(""),
		SSLCiphers:                    types.StringValue("ECDHE"),
		SSLProtocols:                  types.ListNull(types.StringType),
		AdvancedConfiguration:         types.ListNull(types.ObjectType{AttrTypes: advancedConfigAttrTypes()}),
	}

	assert.Equal(t, "cfg1", model.ConfigID.ValueString())
	assert.Equal(t, "pt1", model.ID.ValueString())
	assert.Equal(t, "test-pt", model.Name.ValueString())
	assert.Equal(t, "desc", model.Description.ValueString())
	assert.True(t, model.ACAOHeader.ValueBool())
	assert.False(t, model.XFFHeaderName.IsNull())
	assert.Equal(t, "X-Real-IP", model.XRealIPHeaderName.ValueString())
	assert.False(t, model.CustomListener.ValueBool())
	assert.True(t, model.SSLProtocols.IsNull())
	assert.True(t, model.AdvancedConfiguration.IsNull())
}

// --- buildProxyTemplateAPIModel tests ---

func TestBuildProxyTemplateAPIModel_Basic(t *testing.T) {
	ctx := context.Background()
	model := &ProxyTemplateResourceModel{
		ConfigID:                      types.StringValue("cfg1"),
		ID:                            types.StringValue("pt1"),
		Name:                          types.StringValue("test-pt"),
		Description:                   types.StringValue("desc"),
		ACAOHeader:                    types.BoolValue(true),
		XFFHeaderName:                 types.ListValueMust(types.StringType, []attr.Value{types.StringValue("X-Forwarded-For")}),
		XRealIPHeaderName:             types.StringValue("X-Real-IP"),
		ProxyConnectTimeout:           types.StringValue("5"),
		ProxyReadTimeout:              types.StringValue("60"),
		ProxySendTimeout:              types.StringValue("30"),
		UpstreamHost:                  types.StringValue("$host"),
		ClientBodyTimeout:             types.StringValue("5"),
		ClientBodyBufferSize:          types.StringValue("4"),
		ClientHeaderTimeout:           types.StringValue("5"),
		ClientHeaderBufferSize:        types.StringValue("32"),
		ClientMaxBodySize:             types.StringValue("150"),
		KeepaliveTimeout:              types.StringValue("660"),
		SendTimeout:                   types.StringValue("5"),
		LimitReqRate:                  types.StringValue("1200"),
		LimitReqBurst:                 types.StringValue("400"),
		MaskHeaders:                   types.StringValue("Server"),
		CustomListener:                types.BoolValue(false),
		LargeClientHeaderBuffersCount: types.StringValue("2"),
		LargeClientHeaderBuffersSize:  types.StringValue("32"),
		ConfSpecific:                  types.StringValue(""),
		SSLConfSpecific:               types.StringValue(""),
		SSLCiphers:                    types.StringValue("ECDHE"),
		SSLProtocols:                  types.ListNull(types.StringType),
		AdvancedConfiguration:         types.ListNull(types.ObjectType{AttrTypes: advancedConfigAttrTypes()}),
	}

	pt := buildProxyTemplateAPIModel(ctx, model)

	assert.Equal(t, "pt1", pt.ID)
	assert.Equal(t, "test-pt", pt.Name)
	assert.Equal(t, "desc", pt.Description)
	assert.True(t, pt.ACAOHeader)
	assert.Equal(t, []string{"X-Forwarded-For"}, pt.XFFHeaderName)
	assert.Equal(t, "X-Real-IP", pt.XRealIPHeaderName)
	assert.Equal(t, "5", pt.ProxyConnectTimeout)
	assert.Equal(t, "60", pt.ProxyReadTimeout)
	assert.Equal(t, "30", pt.ProxySendTimeout)
	assert.Equal(t, "$host", pt.UpstreamHost)
	assert.Equal(t, "5", pt.ClientBodyTimeout)
	assert.Equal(t, "4", pt.ClientBodyBufferSize)
	assert.Equal(t, "5", pt.ClientHeaderTimeout)
	assert.Equal(t, "32", pt.ClientHeaderBufferSize)
	assert.Equal(t, "150", pt.ClientMaxBodySize)
	assert.Equal(t, "660", pt.KeepaliveTimeout)
	assert.Equal(t, "5", pt.SendTimeout)
	assert.Equal(t, "1200", pt.LimitReqRate)
	assert.Equal(t, "400", pt.LimitReqBurst)
	assert.Equal(t, "Server", pt.MaskHeaders)
	assert.False(t, pt.CustomListener)
	assert.Equal(t, "2", pt.LargeClientHeaderBuffersCount)
	assert.Equal(t, "32", pt.LargeClientHeaderBuffersSize)
	assert.Equal(t, "", pt.ConfSpecific)
	assert.Equal(t, "", pt.SSLConfSpecific)
	assert.Equal(t, "ECDHE", pt.SSLCiphers)
	assert.Empty(t, pt.AdvancedConfiguration)
}

func TestBuildProxyTemplateAPIModel_WithSSLProtocols(t *testing.T) {
	ctx := context.Background()

	sslList, diags := types.ListValueFrom(ctx, types.StringType, []string{"TLSv1.2", "TLSv1.3"})
	require.False(t, diags.HasError())

	model := &ProxyTemplateResourceModel{
		ConfigID:                      types.StringValue("cfg1"),
		ID:                            types.StringValue("pt1"),
		Name:                          types.StringValue("ssl-test"),
		Description:                   types.StringValue(""),
		ACAOHeader:                    types.BoolValue(false),
		XFFHeaderName:                 types.ListValueMust(types.StringType, []attr.Value{types.StringValue("X-Forwarded-For")}),
		XRealIPHeaderName:             types.StringValue("X-Real-IP"),
		ProxyConnectTimeout:           types.StringValue("5"),
		ProxyReadTimeout:              types.StringValue("60"),
		ProxySendTimeout:              types.StringValue("30"),
		UpstreamHost:                  types.StringValue("$host"),
		ClientBodyTimeout:             types.StringValue("5"),
		ClientBodyBufferSize:          types.StringValue("4"),
		ClientHeaderTimeout:           types.StringValue("5"),
		ClientHeaderBufferSize:        types.StringValue("32"),
		ClientMaxBodySize:             types.StringValue("150"),
		KeepaliveTimeout:              types.StringValue("660"),
		SendTimeout:                   types.StringValue("5"),
		LimitReqRate:                  types.StringValue("1200"),
		LimitReqBurst:                 types.StringValue("400"),
		MaskHeaders:                   types.StringValue(""),
		CustomListener:                types.BoolValue(false),
		LargeClientHeaderBuffersCount: types.StringValue("2"),
		LargeClientHeaderBuffersSize:  types.StringValue("32"),
		ConfSpecific:                  types.StringValue(""),
		SSLConfSpecific:               types.StringValue(""),
		SSLCiphers:                    types.StringValue("ECDHE"),
		SSLProtocols:                  sslList,
		AdvancedConfiguration:         types.ListNull(types.ObjectType{AttrTypes: advancedConfigAttrTypes()}),
	}

	pt := buildProxyTemplateAPIModel(ctx, model)

	assert.Equal(t, []string{"TLSv1.2", "TLSv1.3"}, pt.SSLProtocols)
}

func TestBuildProxyTemplateAPIModel_WithAdvancedConfig(t *testing.T) {
	ctx := context.Background()

	protoList, diags := types.ListValueFrom(ctx, types.StringType, []string{"http", "https"})
	require.False(t, diags.HasError())

	advModels := []AdvancedConfigModel{
		{
			Name:          types.StringValue("custom-block"),
			Protocol:      protoList,
			Configuration: types.StringValue("proxy_set_header X-Custom true;"),
			Description:   types.StringValue("Custom header"),
		},
	}

	advList, diags := types.ListValueFrom(ctx, types.ObjectType{AttrTypes: advancedConfigAttrTypes()}, advModels)
	require.False(t, diags.HasError())

	model := &ProxyTemplateResourceModel{
		ConfigID:                      types.StringValue("cfg1"),
		ID:                            types.StringValue("pt1"),
		Name:                          types.StringValue("adv-test"),
		Description:                   types.StringValue(""),
		ACAOHeader:                    types.BoolValue(false),
		XFFHeaderName:                 types.ListValueMust(types.StringType, []attr.Value{types.StringValue("X-Forwarded-For")}),
		XRealIPHeaderName:             types.StringValue("X-Real-IP"),
		ProxyConnectTimeout:           types.StringValue("5"),
		ProxyReadTimeout:              types.StringValue("60"),
		ProxySendTimeout:              types.StringValue("30"),
		UpstreamHost:                  types.StringValue("$host"),
		ClientBodyTimeout:             types.StringValue("5"),
		ClientBodyBufferSize:          types.StringValue("4"),
		ClientHeaderTimeout:           types.StringValue("5"),
		ClientHeaderBufferSize:        types.StringValue("32"),
		ClientMaxBodySize:             types.StringValue("150"),
		KeepaliveTimeout:              types.StringValue("660"),
		SendTimeout:                   types.StringValue("5"),
		LimitReqRate:                  types.StringValue("1200"),
		LimitReqBurst:                 types.StringValue("400"),
		MaskHeaders:                   types.StringValue(""),
		CustomListener:                types.BoolValue(false),
		LargeClientHeaderBuffersCount: types.StringValue("2"),
		LargeClientHeaderBuffersSize:  types.StringValue("32"),
		ConfSpecific:                  types.StringValue(""),
		SSLConfSpecific:               types.StringValue(""),
		SSLCiphers:                    types.StringValue("ECDHE"),
		SSLProtocols:                  types.ListNull(types.StringType),
		AdvancedConfiguration:         advList,
	}

	pt := buildProxyTemplateAPIModel(ctx, model)

	require.Len(t, pt.AdvancedConfiguration, 1)
	assert.Equal(t, "custom-block", pt.AdvancedConfiguration[0].Name)
	assert.Equal(t, []string{"http", "https"}, pt.AdvancedConfiguration[0].Protocol)
	assert.Equal(t, "proxy_set_header X-Custom true;", pt.AdvancedConfiguration[0].Configuration)
	assert.Equal(t, "Custom header", pt.AdvancedConfiguration[0].Description)
}

func TestBuildProxyTemplateAPIModel_NullSSLProtocols(t *testing.T) {
	ctx := context.Background()
	model := &ProxyTemplateResourceModel{
		ConfigID:                      types.StringValue("cfg1"),
		ID:                            types.StringValue("pt1"),
		Name:                          types.StringValue("null-ssl"),
		Description:                   types.StringValue(""),
		ACAOHeader:                    types.BoolValue(false),
		XFFHeaderName:                 types.ListValueMust(types.StringType, []attr.Value{types.StringValue("X-Forwarded-For")}),
		XRealIPHeaderName:             types.StringValue("X-Real-IP"),
		ProxyConnectTimeout:           types.StringValue("5"),
		ProxyReadTimeout:              types.StringValue("60"),
		ProxySendTimeout:              types.StringValue("30"),
		UpstreamHost:                  types.StringValue("$host"),
		ClientBodyTimeout:             types.StringValue("5"),
		ClientBodyBufferSize:          types.StringValue("4"),
		ClientHeaderTimeout:           types.StringValue("5"),
		ClientHeaderBufferSize:        types.StringValue("32"),
		ClientMaxBodySize:             types.StringValue("150"),
		KeepaliveTimeout:              types.StringValue("660"),
		SendTimeout:                   types.StringValue("5"),
		LimitReqRate:                  types.StringValue("1200"),
		LimitReqBurst:                 types.StringValue("400"),
		MaskHeaders:                   types.StringValue(""),
		CustomListener:                types.BoolValue(false),
		LargeClientHeaderBuffersCount: types.StringValue("2"),
		LargeClientHeaderBuffersSize:  types.StringValue("32"),
		ConfSpecific:                  types.StringValue(""),
		SSLConfSpecific:               types.StringValue(""),
		SSLCiphers:                    types.StringValue("ECDHE"),
		SSLProtocols:                  types.ListNull(types.StringType),
		AdvancedConfiguration:         types.ListNull(types.ObjectType{AttrTypes: advancedConfigAttrTypes()}),
	}

	pt := buildProxyTemplateAPIModel(ctx, model)

	assert.Nil(t, pt.SSLProtocols)
}

func TestBuildProxyTemplateAPIModel_NullAdvancedConfig(t *testing.T) {
	ctx := context.Background()
	model := &ProxyTemplateResourceModel{
		ConfigID:                      types.StringValue("cfg1"),
		ID:                            types.StringValue("pt1"),
		Name:                          types.StringValue("null-adv"),
		Description:                   types.StringValue(""),
		ACAOHeader:                    types.BoolValue(false),
		XFFHeaderName:                 types.ListValueMust(types.StringType, []attr.Value{types.StringValue("X-Forwarded-For")}),
		XRealIPHeaderName:             types.StringValue("X-Real-IP"),
		ProxyConnectTimeout:           types.StringValue("5"),
		ProxyReadTimeout:              types.StringValue("60"),
		ProxySendTimeout:              types.StringValue("30"),
		UpstreamHost:                  types.StringValue("$host"),
		ClientBodyTimeout:             types.StringValue("5"),
		ClientBodyBufferSize:          types.StringValue("4"),
		ClientHeaderTimeout:           types.StringValue("5"),
		ClientHeaderBufferSize:        types.StringValue("32"),
		ClientMaxBodySize:             types.StringValue("150"),
		KeepaliveTimeout:              types.StringValue("660"),
		SendTimeout:                   types.StringValue("5"),
		LimitReqRate:                  types.StringValue("1200"),
		LimitReqBurst:                 types.StringValue("400"),
		MaskHeaders:                   types.StringValue(""),
		CustomListener:                types.BoolValue(false),
		LargeClientHeaderBuffersCount: types.StringValue("2"),
		LargeClientHeaderBuffersSize:  types.StringValue("32"),
		ConfSpecific:                  types.StringValue(""),
		SSLConfSpecific:               types.StringValue(""),
		SSLCiphers:                    types.StringValue("ECDHE"),
		SSLProtocols:                  types.ListNull(types.StringType),
		AdvancedConfiguration:         types.ListNull(types.ObjectType{AttrTypes: advancedConfigAttrTypes()}),
	}

	pt := buildProxyTemplateAPIModel(ctx, model)

	assert.Empty(t, pt.AdvancedConfiguration)
}
