// Package provider acceptance tests. This file covers WP-2552: include/exclude
// used to be set-nested blocks, so Terraform identified them by a hash of their
// whole value and adding a single tag re-rendered the entire block. They are now
// single nested blocks, and the tags list must diff element by element.
package provider

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

// regexpExactlyOneExclude matches the diagnostic raised when the mandatory
// 'exclude' block is missing.
var regexpExactlyOneExclude = regexp.MustCompile(`(?s)Invalid exclude configuration.*Exactly one 'exclude' block must be specified`)

// ruleMockBackend is a minimal in-memory implementation of the dynamic rule and
// rate limit rule CRUD endpoints. Like globalFilterMockBackend it stores and
// echoes back exactly the JSON the provider sends, which is enough to drive a
// full plan/apply/refresh/destroy cycle without spurious diffs.
type ruleMockBackend struct {
	mu    sync.Mutex
	store map[string]json.RawMessage // "<collection>/<entryID>" -> raw rule JSON
}

func newRuleMockBackend() *ruleMockBackend {
	return &ruleMockBackend{store: map[string]json.RawMessage{}}
}

func (b *ruleMockBackend) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Path shape: /api/v4.3/conf/{configID}/{collection}/{entryID}
	// parts: [0]=api [1]=v4.3 [2]=conf [3]=configID [4]=collection [5]=entryID
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 6 {
		http.NotFound(w, r)
		return
	}
	switch parts[4] {
	case "dynamic-rules", "rate-limit-rules":
	default:
		http.NotFound(w, r)
		return
	}
	key := parts[4] + "/" + parts[5]

	b.mu.Lock()
	defer b.mu.Unlock()

	switch r.Method {
	case http.MethodPost, http.MethodPut:
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		b.store[key] = body
		if r.Method == http.MethodPost {
			w.WriteHeader(http.StatusCreated)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	case http.MethodGet:
		stored, ok := b.store[key]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"error":"not found"}`)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(stored)
	case http.MethodDelete:
		delete(b.store, key)
		w.WriteHeader(http.StatusOK)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// startRuleMockBackend redirects the provider's HTTP client at a local mock
// backend and returns once the environment is configured.
func startRuleMockBackend(t *testing.T) {
	t.Helper()

	srv := httptest.NewServer(newRuleMockBackend())
	t.Cleanup(srv.Close)

	target, err := url.Parse(srv.URL)
	if err != nil {
		t.Fatalf("parsing mock server URL: %v", err)
	}

	origTransport := http.DefaultTransport
	http.DefaultTransport = &redirectTransport{target: target, base: origTransport}
	t.Cleanup(func() { http.DefaultTransport = origTransport })

	t.Setenv("LINK11_DOMAIN", acceptanceFakeDomain)
	t.Setenv("LINK11_API_KEY", "test-key")
}

func ruleProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"link11waap": providerserver.NewProtocol6WithError(New("test")()),
	}
}

// TestAccDynamicRuleResource_TagFilterInPlaceUpdate is the WP-2552 regression
// test. Adding one tag to exclude.tags must be an in-place update that keeps the
// rest of the block intact, not a replacement of the whole block.
func TestAccDynamicRuleResource_TagFilterInPlaceUpdate(t *testing.T) {
	startRuleMockBackend(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: ruleProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccDynamicRuleConfig(`["tor", "facebook"]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("link11waap_dynamic_rule.test", "include.relation", "OR"),
					resource.TestCheckResourceAttr("link11waap_dynamic_rule.test", "include.tags.#", "1"),
					resource.TestCheckResourceAttr("link11waap_dynamic_rule.test", "exclude.relation", "AND"),
					resource.TestCheckResourceAttr("link11waap_dynamic_rule.test", "exclude.tags.#", "2"),
					resource.TestCheckResourceAttr("link11waap_dynamic_rule.test", "exclude.tags.0", "tor"),
					resource.TestCheckResourceAttr("link11waap_dynamic_rule.test", "exclude.tags.1", "facebook"),
				),
			},
			{
				// The reported scenario: one tag appended to the list.
				Config: testAccDynamicRuleConfig(`["tor", "facebook", "global-blacklist"]`),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("link11waap_dynamic_rule.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("link11waap_dynamic_rule.test", "exclude.relation", "AND"),
					resource.TestCheckResourceAttr("link11waap_dynamic_rule.test", "exclude.tags.#", "3"),
					resource.TestCheckResourceAttr("link11waap_dynamic_rule.test", "exclude.tags.0", "tor"),
					resource.TestCheckResourceAttr("link11waap_dynamic_rule.test", "exclude.tags.1", "facebook"),
					resource.TestCheckResourceAttr("link11waap_dynamic_rule.test", "exclude.tags.2", "global-blacklist"),
					// The include block is untouched by the change.
					resource.TestCheckResourceAttr("link11waap_dynamic_rule.test", "include.relation", "OR"),
					resource.TestCheckResourceAttr("link11waap_dynamic_rule.test", "include.tags.0", "api"),
				),
			},
			{
				// Re-applying the same configuration must be a no-op.
				Config: testAccDynamicRuleConfig(`["tor", "facebook", "global-blacklist"]`),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
		},
	})
}

// TestAccDynamicRuleResource_MissingExcludeBlock verifies that "exactly one
// block" is still enforced now that the schema can only express "at most one".
func TestAccDynamicRuleResource_MissingExcludeBlock(t *testing.T) {
	startRuleMockBackend(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: ruleProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config:      testAccDynamicRuleNoExcludeConfig,
				ExpectError: regexpExactlyOneExclude,
			},
		},
	})
}

// TestAccDynamicRuleResource_DynamicTagFilterBlock exercises the unknown-value
// path: a `dynamic "exclude"` block whose for_each is unresolved at plan time
// makes the whole block unknown. That must not crash the provider (see the
// dynamic-entry crash fixed for global filters) and must not trip the
// "exactly one block" check at validate time.
func TestAccDynamicRuleResource_DynamicTagFilterBlock(t *testing.T) {
	startRuleMockBackend(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: ruleProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccDynamicRuleDynamicBlockConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("link11waap_dynamic_rule.dynamic", "exclude.relation", "OR"),
					resource.TestCheckResourceAttr("link11waap_dynamic_rule.dynamic", "exclude.tags.#", "1"),
				),
			},
		},
	})
}

// TestAccRateLimitRuleResource_TagFilterInPlaceUpdate covers the same fix for
// the rate limit rule, including the case where both blocks are omitted --
// which is legal there.
func TestAccRateLimitRuleResource_TagFilterInPlaceUpdate(t *testing.T) {
	startRuleMockBackend(t)

	resource.Test(t, resource.TestCase{
		IsUnitTest:               true,
		ProtoV6ProviderFactories: ruleProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccRateLimitRuleConfig(`["tor"]`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("link11waap_rate_limit_rule.test", "exclude.tags.#", "1"),
					resource.TestCheckResourceAttr("link11waap_rate_limit_rule.test", "exclude.tags.0", "tor"),
				),
			},
			{
				Config: testAccRateLimitRuleConfig(`["tor", "global-blacklist"]`),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("link11waap_rate_limit_rule.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("link11waap_rate_limit_rule.test", "exclude.tags.#", "2"),
					resource.TestCheckResourceAttr("link11waap_rate_limit_rule.test", "exclude.tags.1", "global-blacklist"),
					resource.TestCheckResourceAttr("link11waap_rate_limit_rule.test", "include.tags.0", "api"),
				),
			},
		},
	})
}

func testAccDynamicRuleConfig(excludeTags string) string {
	return fmt.Sprintf(`
resource "link11waap_dynamic_rule" "test" {
  config_id            = "test-config"
  name                 = "burst-protection"
  threshold            = 100
  timeframe            = 60
  ttl                  = 3600
  active               = true
  offload_ip_filtering = false
  target               = "ip"
  action               = "action-monitor"

  include {
    relation = "OR"
    tags     = ["api"]
  }

  exclude {
    relation = "AND"
    tags     = %s
  }
}
`, excludeTags)
}

const testAccDynamicRuleNoExcludeConfig = `
resource "link11waap_dynamic_rule" "test" {
  config_id            = "test-config"
  name                 = "burst-protection"
  threshold            = 100
  timeframe            = 60
  ttl                  = 3600
  active               = true
  offload_ip_filtering = false
  target               = "ip"
  action               = "action-monitor"

  include {
    relation = "OR"
    tags     = ["api"]
  }
}
`

const testAccDynamicRuleDynamicBlockConfig = `
resource "link11waap_dynamic_rule" "seed" {
  config_id            = "test-config"
  name                 = "seed-rule"
  threshold            = 100
  timeframe            = 60
  ttl                  = 3600
  active               = false
  offload_ip_filtering = false
  target               = "ip"
  action               = "action-monitor"

  include {
    relation = "OR"
    tags     = ["api"]
  }

  exclude {
    relation = "OR"
    tags     = []
  }
}

resource "link11waap_dynamic_rule" "dynamic" {
  config_id            = "test-config"
  name                 = "dynamic-rule"
  threshold            = 100
  timeframe            = 60
  ttl                  = 3600
  active               = false
  offload_ip_filtering = false
  target               = "ip"
  action               = "action-monitor"

  include {
    relation = "OR"
    tags     = ["api"]
  }

  # link11waap_dynamic_rule.seed.id is Computed (UseStateForUnknown) and so is
  # Unknown until "seed" is created. Splitting an unknown string yields an
  # unknown-length collection, which makes the whole "exclude" block Unknown at
  # the initial plan.
  dynamic "exclude" {
    for_each = { for idx, v in split(",", link11waap_dynamic_rule.seed.id) : idx => v }
    content {
      relation = "OR"
      tags     = ["tor"]
    }
  }
}
`

func testAccRateLimitRuleConfig(excludeTags string) string {
	return fmt.Sprintf(`
resource "link11waap_rate_limit_rule" "test" {
  config_id = "test-config"
  name      = "api-rate-limit"
  global    = false
  active    = true
  timeframe = 60
  threshold = 100
  ttl       = 300
  action    = "action-monitor"

  key {
    attrs = "session"
  }

  include {
    relation = "OR"
    tags     = ["api"]
  }

  exclude {
    relation = "OR"
    tags     = %s
  }
}
`, excludeTags)
}
