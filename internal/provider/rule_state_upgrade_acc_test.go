package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// upgradeRawState drives the provider's real UpgradeResourceState RPC with a raw
// state JSON document, exactly as Terraform does when it finds a state written by
// an older schema version on disk. It returns the upgraded value decoded against
// the resource's current schema.
//
// This is the closest automated equivalent of "install the new provider over an
// existing workspace"; the only thing it cannot do is run the previous provider
// binary to produce the JSON, so the JSON below is a verbatim copy of the shape
// found in real state files written by schema version 0.
func upgradeRawState(t *testing.T, typeName string, priorStateJSON string) tftypes.Value {
	t.Helper()
	ctx := context.Background()

	server, err := providerserver.NewProtocol6WithError(New("test")())()
	if err != nil {
		t.Fatalf("creating provider server: %v", err)
	}

	schemaResp, err := server.GetProviderSchema(ctx, &tfprotov6.GetProviderSchemaRequest{})
	if err != nil {
		t.Fatalf("getting provider schema: %v", err)
	}
	resourceSchema, ok := schemaResp.ResourceSchemas[typeName]
	if !ok {
		t.Fatalf("no schema for %q", typeName)
	}

	resp, err := server.UpgradeResourceState(ctx, &tfprotov6.UpgradeResourceStateRequest{
		TypeName: typeName,
		Version:  0,
		RawState: &tfprotov6.RawState{JSON: []byte(priorStateJSON)},
	})
	if err != nil {
		t.Fatalf("upgrading resource state: %v", err)
	}
	for _, d := range resp.Diagnostics {
		if d.Severity == tfprotov6.DiagnosticSeverityError {
			t.Fatalf("unexpected error diagnostic: %s: %s", d.Summary, d.Detail)
		}
	}
	if resp.UpgradedState == nil {
		t.Fatal("expected an upgraded state")
	}

	value, err := resp.UpgradedState.Unmarshal(resourceSchema.ValueType())
	if err != nil {
		t.Fatalf("decoding upgraded state against the current schema: %v", err)
	}
	return value
}

// attrValue walks one level into an object value.
func attrValue(t *testing.T, obj tftypes.Value, name string) tftypes.Value {
	t.Helper()
	var attrs map[string]tftypes.Value
	if err := obj.As(&attrs); err != nil {
		t.Fatalf("reading object attributes: %v", err)
	}
	value, ok := attrs[name]
	if !ok {
		t.Fatalf("attribute %q not present", name)
	}
	return value
}

func assertString(t *testing.T, value tftypes.Value, expected string) {
	t.Helper()
	var actual string
	if err := value.As(&actual); err != nil {
		t.Fatalf("reading string: %v", err)
	}
	if actual != expected {
		t.Errorf("expected %q, got %q", expected, actual)
	}
}

func assertStringList(t *testing.T, value tftypes.Value, expected []string) {
	t.Helper()
	var elements []tftypes.Value
	if err := value.As(&elements); err != nil {
		t.Fatalf("reading list: %v", err)
	}
	if len(elements) != len(expected) {
		t.Fatalf("expected %d elements, got %d", len(expected), len(elements))
	}
	for i, element := range elements {
		assertString(t, element, expected[i])
	}
}

// dynamicRuleStateV0JSON is the on-disk shape of a schema version 0 dynamic rule:
// include/exclude are JSON arrays because the blocks were sets.
const dynamicRuleStateV0JSON = `{
  "config_id": "test-config",
  "id": "dr1",
  "name": "burst-protection",
  "description": "",
  "threshold": 100,
  "timeframe": 60,
  "ttl": 3600,
  "active": true,
  "offload_ip_filtering": false,
  "target": "ip",
  "action": "action-monitor",
  "tags": ["api"],
  "include": [{"relation": "OR", "tags": ["facebook"]}],
  "exclude": [{"relation": "AND", "tags": ["tor"]}]
}`

func TestUpgradeResourceState_DynamicRuleV0(t *testing.T) {
	upgraded := upgradeRawState(t, "link11waap_dynamic_rule", dynamicRuleStateV0JSON)

	assertString(t, attrValue(t, upgraded, "config_id"), "test-config")
	assertString(t, attrValue(t, upgraded, "id"), "dr1")
	assertString(t, attrValue(t, upgraded, "name"), "burst-protection")
	assertString(t, attrValue(t, upgraded, "target"), "ip")
	assertString(t, attrValue(t, upgraded, "action"), "action-monitor")
	assertStringList(t, attrValue(t, upgraded, "tags"), []string{"api"})

	include := attrValue(t, upgraded, "include")
	if include.IsNull() {
		t.Fatal("include must not be null after the upgrade")
	}
	assertString(t, attrValue(t, include, "relation"), "OR")
	assertStringList(t, attrValue(t, include, "tags"), []string{"facebook"})

	exclude := attrValue(t, upgraded, "exclude")
	assertString(t, attrValue(t, exclude, "relation"), "AND")
	assertStringList(t, attrValue(t, exclude, "tags"), []string{"tor"})
}

// rateLimitRuleStateV0JSON also covers the case where both filters were omitted,
// which schema version 0 stored as empty arrays.
const rateLimitRuleStateV0JSON = `{
  "config_id": "test-config",
  "id": "rl1",
  "name": "api-rate-limit",
  "description": "",
  "global": false,
  "active": true,
  "timeframe": 60,
  "threshold": 100,
  "ttl": 300,
  "action": "action-monitor",
  "is_action_ban": false,
  "pairwith": "{\"self\":\"self\"}",
  "tags": null,
  "key": [{"attrs": "session", "args": null, "plugins": null, "cookies": null, "headers": null}],
  "include": [{"relation": "OR", "tags": ["api"]}],
  "exclude": []
}`

func TestUpgradeResourceState_RateLimitRuleV0(t *testing.T) {
	upgraded := upgradeRawState(t, "link11waap_rate_limit_rule", rateLimitRuleStateV0JSON)

	assertString(t, attrValue(t, upgraded, "id"), "rl1")
	assertString(t, attrValue(t, upgraded, "name"), "api-rate-limit")
	assertString(t, attrValue(t, upgraded, "pairwith"), `{"self":"self"}`)

	key := attrValue(t, upgraded, "key")
	var keyElements []tftypes.Value
	if err := key.As(&keyElements); err != nil {
		t.Fatalf("reading key list: %v", err)
	}
	if len(keyElements) != 1 {
		t.Fatalf("expected 1 key block, got %d", len(keyElements))
	}
	assertString(t, attrValue(t, keyElements[0], "attrs"), "session")

	include := attrValue(t, upgraded, "include")
	assertString(t, attrValue(t, include, "relation"), "OR")
	assertStringList(t, attrValue(t, include, "tags"), []string{"api"})

	// An omitted block was stored as an empty set and must become a null object,
	// so that a configuration without the block still matches the state.
	if !attrValue(t, upgraded, "exclude").IsNull() {
		t.Error("an omitted exclude block must upgrade to a null object")
	}
}
