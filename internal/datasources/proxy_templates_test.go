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

func TestProxyTemplatesDataSource_Metadata(t *testing.T) {
	d := NewProxyTemplatesDataSource()
	req := dsMetadataReq("link11waap")
	resp := dsMetadataResp()
	d.Metadata(context.Background(), req, resp)
	assert.Equal(t, "link11waap_proxy_templates", resp.TypeName)
}

func TestProxyTemplatesDataSource_Schema(t *testing.T) {
	d := NewProxyTemplatesDataSource()
	req := dsSchemaReq()
	resp := dsSchemaResp()
	d.Schema(context.Background(), req, resp)
	require.NotNil(t, resp.Schema)
	assert.Contains(t, resp.Schema.Attributes, "config_id")
	assert.Contains(t, resp.Schema.Attributes, "proxy_templates")
}

func TestProxyTemplatesDataSource_Configure_InvalidType(t *testing.T) {
	testDSConfigureWithInvalidType(t, NewProxyTemplatesDataSource())
}

func TestProxyTemplatesDataSource_Configure_Nil(t *testing.T) {
	testDSConfigureWithNil(t, NewProxyTemplatesDataSource())
}

func TestProxyTemplatesDataSource_Read_Success(t *testing.T) {
	d := NewProxyTemplatesDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ListResponse[client.ProxyTemplate]{
			Total: 1,
			Items: []client.ProxyTemplate{
				{
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
					AdvancedConfiguration: []client.ProxyTemplateAdvancedConfig{
						{
							Name:          "custom-block",
							Protocol:      []string{"http", "https"},
							Configuration: "proxy_set_header X-Custom true;",
							Description:   "Custom header",
						},
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

func TestProxyTemplatesDataSource_Read_EmptyList(t *testing.T) {
	d := NewProxyTemplatesDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ListResponse[client.ProxyTemplate]{
			Total: 0,
			Items: []client.ProxyTemplate{},
		})
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
	})
	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}

func TestProxyTemplatesDataSource_Read_MultipleTemplates(t *testing.T) {
	d := NewProxyTemplatesDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ListResponse[client.ProxyTemplate]{
			Total: 2,
			Items: []client.ProxyTemplate{
				{
					ID:                            "pt1",
					Name:                          "first",
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
				},
				{
					ID:                            "pt2",
					Name:                          "second",
					XFFHeaderName:                 []string{"X-Forwarded-For"},
					XRealIPHeaderName:             "X-Real-IP",
					ProxyConnectTimeout:           "10",
					ProxyReadTimeout:              "120",
					ProxySendTimeout:              "60",
					UpstreamHost:                  "$host",
					ClientBodyTimeout:             "10",
					ClientBodyBufferSize:          "8",
					ClientHeaderTimeout:           "10",
					ClientHeaderBufferSize:        "64",
					ClientMaxBodySize:             "300",
					KeepaliveTimeout:              "120",
					SendTimeout:                   "10",
					LimitReqRate:                  "2400",
					LimitReqBurst:                 "800",
					LargeClientHeaderBuffersCount: "4",
					LargeClientHeaderBuffersSize:  "64",
					SSLProtocols:                  []string{"TLSv1.3"},
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

func TestProxyTemplatesDataSource_Read_NilSSLProtocols(t *testing.T) {
	d := NewProxyTemplatesDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ListResponse[client.ProxyTemplate]{
			Total: 1,
			Items: []client.ProxyTemplate{
				{
					ID:                            "pt1",
					Name:                          "nil-ssl",
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
					SSLProtocols:                  nil, // nil SSLProtocols -> null list
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

func TestProxyTemplatesDataSource_Read_NilAdvancedConfig(t *testing.T) {
	d := NewProxyTemplatesDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ListResponse[client.ProxyTemplate]{
			Total: 1,
			Items: []client.ProxyTemplate{
				{
					ID:                            "pt1",
					Name:                          "nil-adv",
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
					AdvancedConfiguration:         nil, // nil AdvancedConfiguration -> null list
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

func TestProxyTemplatesDataSource_Read_WithAdvancedConfig(t *testing.T) {
	d := NewProxyTemplatesDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ListResponse[client.ProxyTemplate]{
			Total: 1,
			Items: []client.ProxyTemplate{
				{
					ID:                            "pt1",
					Name:                          "adv-config",
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
							Name:          "block1",
							Protocol:      []string{"http", "https"},
							Configuration: "add_header X-Test true;",
							Description:   "Test block",
						},
						{
							Name:          "block2",
							Protocol:      []string{"https"},
							Configuration: "ssl_stapling on;",
							Description:   "SSL stapling",
						},
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

func TestProxyTemplatesDataSource_Read_WithAdvancedConfig_NilProtocol(t *testing.T) {
	d := NewProxyTemplatesDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(client.ListResponse[client.ProxyTemplate]{
			Total: 1,
			Items: []client.ProxyTemplate{
				{
					ID:                            "pt1",
					Name:                          "adv-nil-proto",
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
							Name:          "block-nil-proto",
							Protocol:      nil, // nil Protocol -> null list
							Configuration: "proxy_pass http://backend;",
							Description:   "Nil protocol block",
						},
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

func TestProxyTemplatesDataSource_Read_APIError(t *testing.T) {
	d := NewProxyTemplatesDataSource()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"message":"server error"}`))
	})
	configureDatasourceWithMock(t, d, handler)

	resp := readDatasource(t, d, map[string]tftypes.Value{
		"config_id": tftypes.NewValue(tftypes.String, "cfg1"),
	})
	assert.True(t, resp.Diagnostics.HasError())
}
