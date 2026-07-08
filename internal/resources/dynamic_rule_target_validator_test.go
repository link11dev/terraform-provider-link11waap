package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestDynamicRuleTargetValidator(t *testing.T) {
	cases := map[string]struct {
		value     types.String
		wantError bool
	}{
		"ip":                      {types.StringValue("ip"), false},
		"asn":                     {types.StringValue("asn"), false},
		"country":                 {types.StringValue("country"), false},
		"organization":            {types.StringValue("organization"), false},
		"arguments with name":     {types.StringValue("arguments_dddd"), false},
		"headers with name":       {types.StringValue("headers_x-test-header"), false},
		"cookies with name":       {types.StringValue("cookies_some_cookie"), false},
		"arguments missing name":  {types.StringValue("arguments"), true},
		"arguments empty name":    {types.StringValue("arguments_"), true},
		"headers missing name":    {types.StringValue("headers"), true},
		"cookies missing name":    {types.StringValue("cookies"), true},
		"unknown target":          {types.StringValue("session"), true},
		"unknown prefixed target": {types.StringValue("body_foo"), true},
		"null is skipped":         {types.StringNull(), false},
		"unknown is skipped":      {types.StringUnknown(), false},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			req := validator.StringRequest{
				Path:        path.Root("target"),
				ConfigValue: tc.value,
			}
			resp := &validator.StringResponse{}
			DynamicRuleTarget().ValidateString(context.Background(), req, resp)
			if resp.Diagnostics.HasError() != tc.wantError {
				t.Errorf("value %v: expected error=%v, got diagnostics=%v", tc.value, tc.wantError, resp.Diagnostics)
			}
		})
	}
}
