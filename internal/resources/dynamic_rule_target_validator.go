package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.String = dynamicRuleTargetValidator{}

// dynamicRuleTargetSimpleValues are target values that stand on their own,
// with no additional qualifier.
var dynamicRuleTargetSimpleValues = []string{"ip", "asn", "country", "organization"}

// dynamicRuleTargetPrefixes are target types that must be followed by "_"
// and a non-empty qualifier (e.g. "headers_x-test-header").
var dynamicRuleTargetPrefixes = []string{"arguments", "headers", "cookies"}

// dynamicRuleTargetValidator validates that a target attribute is either one
// of the simple target values, or a "<prefix>_<name>" pair where prefix is
// one of arguments, headers, or cookies and name is non-empty.
type dynamicRuleTargetValidator struct{}

// Description describes the validation in plain text formatting.
func (v dynamicRuleTargetValidator) Description(_ context.Context) string {
	return fmt.Sprintf(
		"value must be one of %s, or one of %s followed by an underscore and a name (e.g. \"cookies_some_cookie\")",
		joinQuoted(dynamicRuleTargetSimpleValues),
		joinQuoted(dynamicRuleTargetPrefixes),
	)
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v dynamicRuleTargetValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// ValidateString performs the validation.
func (v dynamicRuleTargetValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()

	for _, simple := range dynamicRuleTargetSimpleValues {
		if value == simple {
			return
		}
	}

	for _, prefix := range dynamicRuleTargetPrefixes {
		name, found := strings.CutPrefix(value, prefix+"_")
		if !found {
			continue
		}
		if name == "" {
			resp.Diagnostics.AddAttributeError(
				req.Path,
				"Invalid Value",
				fmt.Sprintf("%s, got: %q (missing name after %q)", v.Description(ctx), value, prefix+"_"),
			)
		}
		return
	}

	resp.Diagnostics.AddAttributeError(
		req.Path,
		"Invalid Value",
		fmt.Sprintf("%s, got: %q", v.Description(ctx), value),
	)
}

// DynamicRuleTarget returns a validator which ensures that a configured
// string attribute is a valid dynamic rule target.
func DynamicRuleTarget() validator.String {
	return dynamicRuleTargetValidator{}
}

func joinQuoted(values []string) string {
	quoted := make([]string, len(values))
	for i, v := range values {
		quoted[i] = fmt.Sprintf("%q", v)
	}
	return strings.Join(quoted, ", ")
}
