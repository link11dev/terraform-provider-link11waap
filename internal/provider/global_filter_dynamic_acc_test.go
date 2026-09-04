// Package provider acceptance tests. This file exercises the fix
// (dynamic blocks crashing the provider with a "Value Conversion Error")
// through a real Terraform plan/apply cycle, driven by terraform-plugin-testing
// against an in-process mock backend rather than the live WAAP API.
package provider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// acceptanceFakeDomain is the fake provider "domain" used to select requests
// that redirectTransport should reroute to the local mock backend.
const acceptanceFakeDomain = "acctest.invalid"

// redirectTransport rewrites requests targeting acceptanceFakeDomain to a
// local httptest server, so the provider's real (unmodified) HTTP client can
// be exercised end-to-end -- including the framework's own value decoding --
// without depending on the live WAAP API. It leaves every other request
// (there are none in this test) untouched.
type redirectTransport struct {
	target *url.URL
	base   http.RoundTripper
}

func (rt *redirectTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Host == acceptanceFakeDomain {
		req = req.Clone(req.Context())
		req.URL.Scheme = rt.target.Scheme
		req.URL.Host = rt.target.Host
	}
	return rt.base.RoundTrip(req)
}

// globalFilterMockBackend is a minimal in-memory implementation of the
// global-filters CRUD endpoints (POST/GET/PUT/DELETE
// /conf/{configID}/global-filters/{entryID}), sufficient to drive a full
// Terraform plan/apply/refresh/destroy lifecycle. It stores and echoes back
// exactly the JSON the provider sends, which is enough to round-trip through
// the resource's Read without producing spurious diffs.
type globalFilterMockBackend struct {
	mu    sync.Mutex
	store map[string]json.RawMessage // entryID -> raw filter JSON
}

func newGlobalFilterMockBackend() *globalFilterMockBackend {
	return &globalFilterMockBackend{store: map[string]json.RawMessage{}}
}

func (b *globalFilterMockBackend) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Path shape: /api/v4.3/conf/{configID}/global-filters/{entryID}
	// parts: [0]=api [1]=v4.3 [2]=conf [3]=configID [4]=global-filters [5]=entryID
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 6 || parts[4] != "global-filters" {
		http.NotFound(w, r)
		return
	}
	entryID := parts[5]

	b.mu.Lock()
	defer b.mu.Unlock()

	switch r.Method {
	case http.MethodPost, http.MethodPut:
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		b.store[entryID] = body
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusCreated)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	case http.MethodGet:
		stored, ok := b.store[entryID]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"error":"not found"}`)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(stored)
	case http.MethodDelete:
		delete(b.store, entryID)
		w.WriteHeader(http.StatusOK)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// TestAccGlobalFilterResource_DynamicEntryBlocks is the regression
// test: it drives a real `terraform plan`/`apply` (via terraform-plugin-testing
// and a locally installed/downloaded Terraform CLI) over a config containing a
// `dynamic "entry"` block whose for_each is unknown at initial plan time --
// reproducing the exact crash reported in the ticket ("Value Conversion
// Error ... Path: rule.entry ... Received unknown value, however the target
// type cannot handle unknown values").
//
// The `for_each` is derived by splitting the *seed* filter's generated id
// (`link11waap_global_filter.seed.id`), which is Computed/UseStateForUnknown
// and therefore genuinely Unknown until the seed resource is created. Splitting
// an unknown string yields an unknown-length collection, so Terraform cannot
// determine the `entry` block's instance count at the initial plan -- this is
// the same mechanism that previously crashed the provider, now exercised
// through the real framework decode path (not just this package's own Go
// functions, unlike the unit tests in internal/resources).
func TestAccGlobalFilterResource_DynamicEntryBlocks(t *testing.T) {
	backend := newGlobalFilterMockBackend()
	srv := httptest.NewServer(backend)
	defer srv.Close()

	target, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatalf("parsing mock server URL: %v", err)
	}

	origTransport := http.DefaultTransport
	http.DefaultTransport = &redirectTransport{target: target, base: origTransport}
	t.Cleanup(func() { http.DefaultTransport = origTransport })

	t.Setenv("LINK11_DOMAIN", acceptanceFakeDomain)
	t.Setenv("LINK11_API_KEY", "test-key")

	resource.Test(t, resource.TestCase{
		// IsUnitTest lets this run as part of the normal `go test` suite
		// (against the local mock backend above) without requiring the
		// TF_ACC=1 opt-in that gates tests against a real backend.
		IsUnitTest: true,
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"link11waap": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalFilterDynamicEntryConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Static-block seed filter: the known-good baseline.
					resource.TestCheckResourceAttr("link11waap_global_filter.seed", "rule.entry.#", "1"),
					resource.TestCheckResourceAttr("link11waap_global_filter.seed", "rule.entry.0.type", "ip"),
					resource.TestCheckResourceAttr("link11waap_global_filter.seed", "rule.entry.0.value", "10.0.0.1"),
					// Dynamic-block filter: must produce the same shape of
					// result as the static one, with no crash on plan/apply.
					resource.TestCheckResourceAttr("link11waap_global_filter.dynamic", "rule.entry.#", "1"),
					resource.TestCheckResourceAttr("link11waap_global_filter.dynamic", "rule.entry.0.type", "ip"),
					resource.TestCheckResourceAttrSet("link11waap_global_filter.dynamic", "rule.entry.0.value"),
				),
			},
		},
	})
}

const testAccGlobalFilterDynamicEntryConfig = `
resource "link11waap_global_filter" "seed" {
  config_id = "test-config"
  name      = "seed-filter"
  active    = false
  action    = "action-monitor"

  rule {
    relation = "OR"
    entry {
      type  = "ip"
      value = "10.0.0.1"
    }
  }
}

resource "link11waap_global_filter" "dynamic" {
  config_id = "test-config"
  name      = "dynamic-filter"
  active    = false
  action    = "action-monitor"

  rule {
    relation = "OR"

    dynamic "entry" {
      # link11waap_global_filter.seed.id is Computed (UseStateForUnknown) and
      # therefore Unknown until "seed" is created. Splitting an unknown
      # string yields an unknown-length list, so the whole for_each -- and
      # therefore the whole "entry" collection -- is Unknown at the initial
      # plan. This is the exact shape that used to crash the provider.
      for_each = { for idx, v in split(",", link11waap_global_filter.seed.id) : idx => v }
      content {
        type  = "ip"
        value = "10.0.0.${entry.key}"
      }
    }
  }
}
`
