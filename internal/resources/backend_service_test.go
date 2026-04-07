package resources

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestValidateHostResolution_ValidIPv4(t *testing.T) {
	err := validateHostResolution("192.168.1.1")
	if err != nil {
		t.Errorf("expected no error for valid IPv4 address, got: %v", err)
	}
}

func TestValidateHostResolution_ValidIPv6(t *testing.T) {
	err := validateHostResolution("::1")
	if err != nil {
		t.Errorf("expected no error for valid IPv6 address, got: %v", err)
	}
}

func TestValidateHostResolution_ValidIPv6Full(t *testing.T) {
	err := validateHostResolution("2001:0db8:85a3:0000:0000:8a2e:0370:7334")
	if err != nil {
		t.Errorf("expected no error for valid full IPv6 address, got: %v", err)
	}
}

func TestValidateHostResolution_ValidHostname(t *testing.T) {
	// Stub the lookup function to simulate successful resolution.
	original := lookupHostFunc
	lookupHostFunc = func(_ string) ([]string, error) {
		return []string{"93.184.216.34"}, nil
	}
	defer func() { lookupHostFunc = original }()

	err := validateHostResolution("example.com")
	if err != nil {
		t.Errorf("expected no error for resolvable hostname, got: %v", err)
	}
}

func TestValidateHostResolution_InvalidHostname(t *testing.T) {
	// Stub the lookup function to simulate a DNS failure.
	original := lookupHostFunc
	lookupHostFunc = func(_ string) ([]string, error) {
		return nil, fmt.Errorf("lookup host-that-does-not-exist.invalid: no such host")
	}
	defer func() { lookupHostFunc = original }()

	err := validateHostResolution("host-that-does-not-exist.invalid")
	if err == nil {
		t.Error("expected error for unresolvable hostname, got nil")
	}
}

func TestValidateHostResolution_InvalidHostnameErrorMessage(t *testing.T) {
	original := lookupHostFunc
	lookupHostFunc = func(_ string) ([]string, error) {
		return nil, fmt.Errorf("lookup no-such-host.invalid: no such host")
	}
	defer func() { lookupHostFunc = original }()

	err := validateHostResolution("no-such-host.invalid")
	if err == nil {
		t.Fatal("expected error for unresolvable hostname, got nil")
	}

	expected := `hostname "no-such-host.invalid" does not resolve`
	if got := err.Error(); len(got) < len(expected) {
		t.Errorf("error message too short: %q", got)
	} else if got[:len(expected)] != expected {
		t.Errorf("expected error to start with %q, got %q", expected, got)
	}
}

func TestValidateHostResolution_EmptyStringTreatedAsHostname(t *testing.T) {
	original := lookupHostFunc
	lookupHostFunc = func(_ string) ([]string, error) {
		return nil, fmt.Errorf("lookup : no such host")
	}
	defer func() { lookupHostFunc = original }()

	err := validateHostResolution("")
	if err == nil {
		t.Error("expected error for empty string, got nil")
	}
}

func TestValidateHostResolution_IPAddressSkipsDNS(t *testing.T) {
	// Replace lookup with a function that always fails, to prove IPs skip DNS.
	original := lookupHostFunc
	lookupHostFunc = func(_ string) ([]string, error) {
		t.Error("lookupHostFunc should not be called for an IP address")
		return nil, fmt.Errorf("should not be called")
	}
	defer func() { lookupHostFunc = original }()

	if err := validateHostResolution("10.0.0.1"); err != nil {
		t.Errorf("expected no error for IPv4, got: %v", err)
	}
	if err := validateHostResolution("::1"); err != nil {
		t.Errorf("expected no error for IPv6, got: %v", err)
	}
}

// Additional tests using testify for better coverage

func TestNewBackendServiceResource(t *testing.T) {
	r := NewBackendServiceResource()
	if r == nil {
		t.Fatal("expected non-nil resource")
	}
	_, ok := r.(*BackendServiceResource)
	if !ok {
		t.Fatal("expected *BackendServiceResource")
	}
}

func TestBackendServiceResource_Metadata(t *testing.T) {
	r := &BackendServiceResource{}

	req := metadataReq("link11waap")
	resp := metadataResp()
	r.Metadata(context.Background(), req, resp)

	if resp.TypeName != "link11waap_backend_service" {
		t.Errorf("expected 'link11waap_backend_service', got %q", resp.TypeName)
	}
}

func TestBackendServiceResource_Schema(t *testing.T) {
	r := &BackendServiceResource{}
	ctx := context.Background()

	req := schemaReq()
	resp := schemaResp()
	r.Schema(ctx, req, resp)

	schema := resp.Schema
	if len(schema.Attributes) == 0 {
		t.Fatal("expected non-empty attributes")
	}

	expectedAttrs := []string{
		"config_id", "id", "name", "description", "http11",
		"transport_mode", "sticky", "sticky_cookie_name", "least_conn",
		"mtls_certificate", "mtls_trusted_certificate",
	}
	for _, attr := range expectedAttrs {
		if _, ok := schema.Attributes[attr]; !ok {
			t.Errorf("expected attribute %q in schema", attr)
		}
	}

	// Check blocks
	if _, ok := schema.Blocks["back_hosts"]; !ok {
		t.Error("expected block 'back_hosts' in schema")
	}
}

func TestBackendServiceResource_Configure_NilProvider(t *testing.T) {
	r := &BackendServiceResource{}
	ctx := context.Background()

	req := configureReq(nil)
	resp := configureResp()
	r.Configure(ctx, req, resp)

	if r.client != nil {
		t.Error("expected nil client for nil provider data")
	}
	if resp.Diagnostics.HasError() {
		t.Error("expected no diagnostics errors")
	}
}

func TestBackendServiceResource_ImportState_Valid(t *testing.T) {
	r := &BackendServiceResource{}
	resp := testImportState(t, r, "config123/bs456")

	if resp.Diagnostics.HasError() {
		t.Errorf("expected no errors, got: %v", resp.Diagnostics)
	}
}

func TestBackendServiceResource_ImportState_Invalid(t *testing.T) {
	r := &BackendServiceResource{}
	resp := testImportState(t, r, "invalid")

	if !resp.Diagnostics.HasError() {
		t.Error("expected error for invalid import ID")
	}
}

func TestBackendServiceResource_ImportState_TooManyParts(t *testing.T) {
	r := &BackendServiceResource{}
	resp := testImportState(t, r, "a/b/c")

	if !resp.Diagnostics.HasError() {
		t.Error("expected error for too many parts")
	}
}

func TestIntSliceToInt64(t *testing.T) {
	tests := []struct {
		name  string
		input []int
		want  []int64
	}{
		{"empty", []int{}, []int64{}},
		{"single", []int{80}, []int64{80}},
		{"multiple", []int{80, 443, 8080}, []int64{80, 443, 8080}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := intSliceToInt64(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("expected %d elements, got %d", len(tt.want), len(got))
			}
			for i, v := range got {
				if v != tt.want[i] {
					t.Errorf("index %d: expected %d, got %d", i, tt.want[i], v)
				}
			}
		})
	}
}

func TestInt64SliceToInt(t *testing.T) {
	tests := []struct {
		name  string
		input []int64
		want  []int
	}{
		{"empty", []int64{}, []int{}},
		{"single", []int64{443}, []int{443}},
		{"multiple", []int64{80, 443, 8443}, []int{80, 443, 8443}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := int64SliceToInt(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("expected %d elements, got %d", len(tt.want), len(got))
			}
			for i, v := range got {
				if v != tt.want[i] {
					t.Errorf("index %d: expected %d, got %d", i, tt.want[i], v)
				}
			}
		})
	}
}

func TestBackendHostAttrTypes(t *testing.T) {
	attrTypes := backendHostAttrTypes()
	expectedKeys := []string{
		"host", "http_ports", "https_ports", "weight",
		"max_fails", "fail_timeout", "down", "monitor_state", "backup",
	}
	if len(attrTypes) != len(expectedKeys) {
		t.Errorf("expected %d attr types, got %d", len(expectedKeys), len(attrTypes))
	}
	for _, key := range expectedKeys {
		if _, ok := attrTypes[key]; !ok {
			t.Errorf("expected key %q in attr types", key)
		}
	}
}

func TestIntSliceToInt64_Roundtrip(t *testing.T) {
	original := []int{80, 443, 8080, 8443}
	int64Slice := intSliceToInt64(original)
	result := int64SliceToInt(int64Slice)

	if len(result) != len(original) {
		t.Fatalf("expected %d elements after roundtrip, got %d", len(original), len(result))
	}
	for i, v := range result {
		if v != original[i] {
			t.Errorf("roundtrip index %d: expected %d, got %d", i, original[i], v)
		}
	}
}

func TestBuildBackendServiceAPIModel_BasicFields(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	plan := &BackendServiceResourceModel{
		ID:                     types.StringValue("bs-1"),
		Name:                   types.StringValue("test-bs"),
		Description:            types.StringValue("desc"),
		HTTP11:                 types.BoolValue(true),
		TransportMode:          types.StringValue("default"),
		Sticky:                 types.StringValue("none"),
		StickyCookieName:       types.StringValue(""),
		LeastConn:              types.BoolValue(false),
		BackHosts:              types.SetNull(types.ObjectType{AttrTypes: backendHostAttrTypes()}),
		MtlsCertificate:        types.StringNull(),
		MtlsTrustedCertificate: types.StringNull(),
	}

	bs := buildBackendServiceAPIModel(ctx, plan, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	if bs.ID != "bs-1" {
		t.Errorf("expected ID='bs-1', got '%s'", bs.ID)
	}
	if bs.Name != "test-bs" {
		t.Errorf("expected Name='test-bs', got '%s'", bs.Name)
	}
	if bs.HTTP11 != true {
		t.Errorf("expected HTTP11=true, got %v", bs.HTTP11)
	}
	if bs.TransportMode != "default" {
		t.Errorf("expected TransportMode='default', got '%s'", bs.TransportMode)
	}
	if bs.MtlsCertificate != "" {
		t.Errorf("expected empty MtlsCertificate, got '%s'", bs.MtlsCertificate)
	}
}

func TestBuildBackendServiceAPIModel_WithMTLS(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	plan := &BackendServiceResourceModel{
		ID:                     types.StringValue("bs-2"),
		Name:                   types.StringValue("test"),
		Description:            types.StringValue(""),
		HTTP11:                 types.BoolValue(false),
		TransportMode:          types.StringValue("https"),
		Sticky:                 types.StringValue("none"),
		StickyCookieName:       types.StringValue(""),
		LeastConn:              types.BoolValue(false),
		BackHosts:              types.SetNull(types.ObjectType{AttrTypes: backendHostAttrTypes()}),
		MtlsCertificate:        types.StringValue("mtls-cert-1"),
		MtlsTrustedCertificate: types.StringValue("mtls-ca-1"),
	}

	bs := buildBackendServiceAPIModel(ctx, plan, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	if bs.MtlsCertificate != "mtls-cert-1" {
		t.Errorf("expected MtlsCertificate='mtls-cert-1', got '%s'", bs.MtlsCertificate)
	}
	if bs.MtlsTrustedCertificate != "mtls-ca-1" {
		t.Errorf("expected MtlsTrustedCertificate='mtls-ca-1', got '%s'", bs.MtlsTrustedCertificate)
	}
}

func TestBuildBackendServiceAPIModel_WithBackHosts(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	// Stub lookup to succeed for hostname
	original := lookupHostFunc
	lookupHostFunc = func(_ string) ([]string, error) {
		return []string{"1.2.3.4"}, nil
	}
	defer func() { lookupHostFunc = original }()

	httpPorts, _ := types.ListValueFrom(ctx, types.Int64Type, []int64{80})
	httpsPorts, _ := types.ListValueFrom(ctx, types.Int64Type, []int64{443})

	hostModels := []BackendHostModel{
		{
			Host:         types.StringValue("1.2.3.4"),
			HTTPPorts:    httpPorts,
			HTTPSPorts:   httpsPorts,
			Weight:       types.Int64Value(1),
			MaxFails:     types.Int64Value(3),
			FailTimeout:  types.Int64Value(30),
			Down:         types.BoolValue(false),
			MonitorState: types.StringValue(""),
			Backup:       types.BoolValue(false),
		},
	}

	hostSet, d := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: backendHostAttrTypes()}, hostModels)
	if d.HasError() {
		t.Fatalf("error creating host set: %v", d)
	}

	plan := &BackendServiceResourceModel{
		ID:                     types.StringValue("bs-3"),
		Name:                   types.StringValue("test"),
		Description:            types.StringValue(""),
		HTTP11:                 types.BoolValue(true),
		TransportMode:          types.StringValue("default"),
		Sticky:                 types.StringValue("none"),
		StickyCookieName:       types.StringValue(""),
		LeastConn:              types.BoolValue(false),
		BackHosts:              hostSet,
		MtlsCertificate:        types.StringNull(),
		MtlsTrustedCertificate: types.StringNull(),
	}

	bs := buildBackendServiceAPIModel(ctx, plan, &diags)

	if diags.HasError() {
		t.Fatalf("unexpected error: %v", diags)
	}
	if len(bs.BackHosts) != 1 {
		t.Fatalf("expected 1 back host, got %d", len(bs.BackHosts))
	}
	if bs.BackHosts[0].Host != "1.2.3.4" {
		t.Errorf("expected Host='1.2.3.4', got '%s'", bs.BackHosts[0].Host)
	}
	if len(bs.BackHosts[0].HTTPPorts) != 1 || bs.BackHosts[0].HTTPPorts[0] != 80 {
		t.Errorf("expected HTTPPorts=[80], got %v", bs.BackHosts[0].HTTPPorts)
	}
}

func TestBuildBackendServiceAPIModel_HostResolutionFailure(t *testing.T) {
	ctx := context.Background()
	var diags diag.Diagnostics

	// Stub lookup to fail
	original := lookupHostFunc
	lookupHostFunc = func(_ string) ([]string, error) {
		return nil, fmt.Errorf("no such host")
	}
	defer func() { lookupHostFunc = original }()

	httpPorts, _ := types.ListValueFrom(ctx, types.Int64Type, []int64{80})
	httpsPorts, _ := types.ListValueFrom(ctx, types.Int64Type, []int64{443})

	hostModels := []BackendHostModel{
		{
			Host:         types.StringValue("unresolvable.invalid"),
			HTTPPorts:    httpPorts,
			HTTPSPorts:   httpsPorts,
			Weight:       types.Int64Value(1),
			MaxFails:     types.Int64Value(3),
			FailTimeout:  types.Int64Value(30),
			Down:         types.BoolValue(false),
			MonitorState: types.StringValue(""),
			Backup:       types.BoolValue(false),
		},
	}

	hostSet, d := types.SetValueFrom(ctx, types.ObjectType{AttrTypes: backendHostAttrTypes()}, hostModels)
	if d.HasError() {
		t.Fatalf("error creating host set: %v", d)
	}

	plan := &BackendServiceResourceModel{
		ID:                     types.StringValue("bs-4"),
		Name:                   types.StringValue("test"),
		Description:            types.StringValue(""),
		HTTP11:                 types.BoolValue(true),
		TransportMode:          types.StringValue("default"),
		Sticky:                 types.StringValue("none"),
		StickyCookieName:       types.StringValue(""),
		LeastConn:              types.BoolValue(false),
		BackHosts:              hostSet,
		MtlsCertificate:        types.StringNull(),
		MtlsTrustedCertificate: types.StringNull(),
	}

	bs := buildBackendServiceAPIModel(ctx, plan, &diags)

	if !diags.HasError() {
		t.Error("expected error for unresolvable hostname")
	}
	if bs != nil {
		t.Error("expected nil return for failed host resolution")
	}
}
