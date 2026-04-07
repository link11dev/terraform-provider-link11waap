package datasources

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
func newMockClient(t *testing.T, handler http.Handler) (*client.Client, *httptest.Server) {
	t.Helper()

	server := httptest.NewTLSServer(handler)
	t.Cleanup(server.Close)

	domain := server.Listener.Addr().String()
	c, err := client.New(client.Config{
		Domain:    domain,
		APIKey:    "dGVzdC1rZXk=",
		Timeout:   5 * time.Second,
		RetryMax:  0,
		RetryWait: time.Millisecond,
	})
	require.NoError(t, err)

	rv := reflect.ValueOf(c).Elem()
	httpClientField := rv.FieldByName("httpClient")
	ptr := unsafe.Pointer(httpClientField.UnsafeAddr())
	*(**http.Client)(ptr) = server.Client()

	return c, server
}
