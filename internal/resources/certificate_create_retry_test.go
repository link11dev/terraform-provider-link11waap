package resources

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/link11/terraform-provider-link11waap/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createWithMock performs a full Create operation using a mock API server.
func createWithMock(t *testing.T, r resource.Resource, planValues map[string]tftypes.Value) *resource.CreateResponse {
	t.Helper()
	ctx := context.Background()

	plan := buildTerraformPlan(ctx, t, r, planValues)

	sReq := schemaReq()
	sResp := schemaResp()
	r.Schema(ctx, sReq, sResp)

	emptyState := tfsdk.State{
		Schema: sResp.Schema,
		Raw:    tftypes.NewValue(sResp.Schema.Type().TerraformType(ctx), nil),
	}

	createReq := resource.CreateRequest{Plan: plan}
	createResp := &resource.CreateResponse{State: emptyState}

	r.Create(ctx, createReq, createResp)
	return createResp
}

// certificateCreatePlanValues returns the base plan values for creating a certificate.
func certificateCreatePlanValues() map[string]tftypes.Value {
	return map[string]tftypes.Value{
		"config_id":       tftypes.NewValue(tftypes.String, "cfg1"),
		"id":              tftypes.NewValue(tftypes.String, nil),
		"cert_body":       tftypes.NewValue(tftypes.String, "cert-body"),
		"private_key":     tftypes.NewValue(tftypes.String, "key"),
		"le_auto_renew":   tftypes.NewValue(tftypes.Bool, false),
		"le_auto_replace": tftypes.NewValue(tftypes.Bool, false),
		"side":            tftypes.NewValue(tftypes.String, "server"),
	}
}

// withShortCertificateRetryDelay shrinks the retry delay for the duration of a test.
func withShortCertificateRetryDelay(t *testing.T) {
	t.Helper()
	original := certificateCreateReadRetryDelay
	certificateCreateReadRetryDelay = time.Millisecond
	t.Cleanup(func() { certificateCreateReadRetryDelay = original })
}

// TestCertificateResource_Create_RetriesTransientReadAfterWriteFailure simulates the
// API race condition where GetCertificate briefly 500s right after a successful
// POST, and asserts that Create retries the read instead of failing immediately.
func TestCertificateResource_Create_RetriesTransientReadAfterWriteFailure(t *testing.T) {
	withShortCertificateRetryDelay(t)

	r := &CertificateResource{}
	getAttempts := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodPost {
			w.WriteHeader(http.StatusCreated)
			return
		}

		getAttempts++
		if getAttempts < certificateCreateReadMaxRetries {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"message":"KeyError: '\"c123\" does not exist'"}`))
			return
		}

		json.NewEncoder(w).Encode(client.Certificate{
			ID:   "c123",
			Name: "test-cert",
		})
	})
	c, _ := newMockClient(t, handler)

	req := resource.ConfigureRequest{ProviderData: c}
	cfgResp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), req, cfgResp)
	require.False(t, cfgResp.Diagnostics.HasError())

	createResp := createWithMock(t, r, certificateCreatePlanValues())

	assert.False(t, createResp.Diagnostics.HasError(), "errors: %v", createResp.Diagnostics)
	assert.Equal(t, certificateCreateReadMaxRetries, getAttempts, "expected the read to fail until the final retry")

	var state CertificateResourceModel
	assert.False(t, createResp.State.Get(context.Background(), &state).HasError())
	assert.Equal(t, "test-cert", state.Name.ValueString())
}

// TestCertificateResource_Create_FailsAfterExhaustingRetries asserts that once all
// retries are exhausted, Create surfaces the read error instead of retrying forever.
func TestCertificateResource_Create_FailsAfterExhaustingRetries(t *testing.T) {
	withShortCertificateRetryDelay(t)

	r := &CertificateResource{}
	getAttempts := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodPost {
			w.WriteHeader(http.StatusCreated)
			return
		}

		getAttempts++
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"KeyError: '\"c123\" does not exist'"}`))
	})
	c, _ := newMockClient(t, handler)

	req := resource.ConfigureRequest{ProviderData: c}
	cfgResp := &resource.ConfigureResponse{}
	r.Configure(context.Background(), req, cfgResp)
	require.False(t, cfgResp.Diagnostics.HasError())

	createResp := createWithMock(t, r, certificateCreatePlanValues())

	assert.True(t, createResp.Diagnostics.HasError(), "expected error once retries are exhausted")
	assert.Equal(t, certificateCreateReadMaxRetries+1, getAttempts, "expected the initial read plus all retries")
}
