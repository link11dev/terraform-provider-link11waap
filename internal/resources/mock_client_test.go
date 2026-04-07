package resources

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
	"unsafe"

	"github.com/link11/terraform-provider-link11waap/internal/client"
	"github.com/stretchr/testify/require"
)

// newMockClient creates a *client.Client that points at a TLS test server.
// It uses unsafe to inject the test server's http.Client (which trusts the
// test server's self-signed certificate) into the real client's unexported
// httpClient field.
func newMockClient(t *testing.T, handler http.Handler) (*client.Client, *httptest.Server) {
	t.Helper()

	server := httptest.NewTLSServer(handler)
	t.Cleanup(server.Close)

	// Create a real client.Client via the exported New() constructor.
	// The domain must match the TLS test server's address.
	domain := server.Listener.Addr().String()
	c, err := client.New(client.Config{
		Domain:    domain,
		APIKey:    "dGVzdC1rZXk=",
		Timeout:   5 * time.Second,
		RetryMax:  0,
		RetryWait: time.Millisecond,
	})
	require.NoError(t, err)

	// Replace the unexported httpClient field with the test server's client
	// (which trusts the test server's self-signed TLS certificate).
	rv := reflect.ValueOf(c).Elem()
	httpClientField := rv.FieldByName("httpClient")
	// Use unsafe to get a writable pointer to the unexported field.
	ptr := unsafe.Pointer(httpClientField.UnsafeAddr())
	*(**http.Client)(ptr) = server.Client()

	return c, server
}
