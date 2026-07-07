package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestFullHoursValidator(t *testing.T) {
	cases := map[string]struct {
		value     types.Int64
		wantError bool
	}{
		"one hour":          {types.Int64Value(3600), false},
		"two hours":         {types.Int64Value(7200), false},
		"zero":              {types.Int64Value(0), true},
		"negative":          {types.Int64Value(-3600), true},
		"not a multiple":    {types.Int64Value(300), true},
		"null is skipped":   {types.Int64Null(), false},
		"unknown skipped":   {types.Int64Unknown(), false},
		"partial hour":      {types.Int64Value(5400), true},
		"large valid value": {types.Int64Value(36000), false},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			req := validator.Int64Request{
				Path:        path.Root("ttl"),
				ConfigValue: tc.value,
			}
			resp := &validator.Int64Response{}
			FullHours().ValidateInt64(context.Background(), req, resp)
			if resp.Diagnostics.HasError() != tc.wantError {
				t.Errorf("value %v: expected error=%v, got diagnostics=%v", tc.value, tc.wantError, resp.Diagnostics)
			}
		})
	}
}
