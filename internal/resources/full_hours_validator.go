package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

var _ validator.Int64 = fullHoursValidator{}

// fullHoursValidator validates that an int64 attribute is a positive
// multiple of 3600, i.e. a whole number of hours expressed in seconds.
type fullHoursValidator struct{}

// Description describes the validation in plain text formatting.
func (v fullHoursValidator) Description(_ context.Context) string {
	return "value must be a positive integer representing full hours in seconds (a multiple of 3600)"
}

// MarkdownDescription describes the validation in Markdown formatting.
func (v fullHoursValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

// ValidateInt64 performs the validation.
func (v fullHoursValidator) ValidateInt64(ctx context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueInt64()
	if value <= 0 || value%3600 != 0 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Value",
			fmt.Sprintf("%s, got: %d", v.Description(ctx), value),
		)
	}
}

// FullHours returns a validator which ensures that a configured int64
// attribute is a positive multiple of 3600 (i.e. a whole number of hours in seconds).
func FullHours() validator.Int64 {
	return fullHoursValidator{}
}
