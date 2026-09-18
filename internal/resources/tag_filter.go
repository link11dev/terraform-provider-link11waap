package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/link11/terraform-provider-link11waap/internal/client"
)

// RateLimitTagFilterModel describes the data model for a tag filter (include/exclude).
// Shared by the dynamic rule and rate limit rule resources.
type RateLimitTagFilterModel struct {
	Relation types.String `tfsdk:"relation"`
	Tags     types.List   `tfsdk:"tags"`
}

// tagFilterAttrTypes defines the attribute types for a tag filter object
var tagFilterAttrTypes = map[string]attr.Type{
	"relation": types.StringType,
	"tags":     types.ListType{ElemType: types.StringType},
}

// tagFilterBlockSchema returns the include/exclude block definition shared by
// link11waap_dynamic_rule and link11waap_rate_limit_rule.
//
// The block is a SingleNestedBlock rather than a SetNestedBlock so that Terraform
// diffs 'tags' element by element: a set element is identified by a hash of its
// entire value, so changing one tag used to replace the whole block.
//
// Neither the block itself nor its attributes can express their own
// requiredness here: Terraform blocks cannot be marked Required, and a Required
// attribute inside a SingleNestedBlock reports an attribute-level error when the
// whole block is omitted, which would mask the block-level message. Presence is
// enforced by validateTagFilterBlock instead, which owns every message the user
// sees. objectvalidator.IsRequired() is deliberately not used: it would raise a
// second, differently-worded diagnostic for the same mistake.
func tagFilterBlockSchema(description string) schema.SingleNestedBlock {
	return schema.SingleNestedBlock{
		Description: description,
		Attributes: map[string]schema.Attribute{
			"relation": schema.StringAttribute{
				Description: "Relation between tags. Valid values: OR, AND. Required when the block is present.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("OR", "AND"),
				},
			},
			"tags": schema.ListAttribute{
				Description: "List of tag identifiers. Required when the block is present.",
				Optional:    true,
				ElementType: types.StringType,
			},
		},
	}
}

// validateTagFilterBlock re-implements, for SingleNestedBlock, the guarantees the
// former SetNestedBlock schema gave us: the block must be present (when required)
// and, when present, relation and tags must both be set.
//
// Terraform blocks cannot be marked Required in the schema, so "exactly one block"
// can only be enforced in code.
func validateTagFilterBlock(ctx context.Context, obj types.Object, p path.Path, required bool, diags *diag.Diagnostics) {
	if obj.IsUnknown() {
		// Value comes from a dynamic block or an unresolved expression; it is
		// re-checked when the plan is applied.
		return
	}
	if obj.IsNull() {
		if required {
			diags.AddAttributeError(
				p,
				fmt.Sprintf("Invalid %s configuration", p.String()),
				fmt.Sprintf("Exactly one '%s' block must be specified.", p.String()),
			)
		}
		return
	}

	var m RateLimitTagFilterModel
	diags.Append(obj.As(ctx, &m, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return
	}
	if m.Relation.IsNull() {
		diags.AddAttributeError(
			p.AtName("relation"),
			"Missing relation",
			"'relation' is required when the block is present. Valid values: OR, AND.",
		)
	}
	if m.Tags.IsNull() {
		diags.AddAttributeError(
			p.AtName("tags"),
			"Missing tags",
			"'tags' is required when the block is present. Use tags = [] for an empty filter.",
		)
	}
}

// extractTagFilter converts a Terraform object to an API RateLimitTagFilter.
// A null or unknown object yields the API's neutral filter, which is what an
// omitted block has always been sent as.
func extractTagFilter(ctx context.Context, obj types.Object) (client.RateLimitTagFilter, diag.Diagnostics) {
	var diags diag.Diagnostics
	if obj.IsNull() || obj.IsUnknown() {
		return client.RateLimitTagFilter{Relation: "OR", Tags: []string{}}, diags
	}

	var model RateLimitTagFilterModel
	diags.Append(obj.As(ctx, &model, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return client.RateLimitTagFilter{}, diags
	}

	relation := "OR"
	if !model.Relation.IsNull() && !model.Relation.IsUnknown() {
		relation = model.Relation.ValueString()
	}

	tags := []string{}
	if !model.Tags.IsNull() && !model.Tags.IsUnknown() {
		diags.Append(model.Tags.ElementsAs(ctx, &tags, false)...)
	}

	return client.RateLimitTagFilter{Relation: relation, Tags: tags}, diags
}

// tagFilterToObject converts an API RateLimitTagFilter to a Terraform object.
func tagFilterToObject(ctx context.Context, filter client.RateLimitTagFilter) (types.Object, diag.Diagnostics) {
	tags := filter.Tags
	if tags == nil {
		tags = []string{}
	}
	tagsList, diags := types.ListValueFrom(ctx, types.StringType, tags)
	if diags.HasError() {
		return types.ObjectNull(tagFilterAttrTypes), diags
	}
	obj, d := types.ObjectValue(tagFilterAttrTypes, map[string]attr.Value{
		"relation": types.StringValue(filter.Relation),
		"tags":     tagsList,
	})
	diags.Append(d...)
	if diags.HasError() {
		return types.ObjectNull(tagFilterAttrTypes), diags
	}
	return obj, diags
}

// tagFilterSetToObject converts a schema version 0 include/exclude value (a set of
// zero or more tag filter objects) into the version 1 value (a single object).
//
// Both resources only ever supported a single filter, so a set holding more than
// one element can only come from a configuration the old schema silently accepted.
// Sets are unordered, so which element survives is not meaningful: a warning is
// emitted so the practitioner verifies the result.
func tagFilterSetToObject(ctx context.Context, set types.Set, p path.Path) (types.Object, diag.Diagnostics) {
	var diags diag.Diagnostics

	if set.IsNull() || set.IsUnknown() || len(set.Elements()) == 0 {
		return types.ObjectNull(tagFilterAttrTypes), diags
	}

	elements := set.Elements()
	obj, ok := elements[0].(types.Object)
	if !ok {
		diags.AddAttributeError(
			p,
			"Unexpected State Value",
			fmt.Sprintf("Expected a %s object in the prior state, got %T. The state cannot be upgraded automatically.", p.String(), elements[0]),
		)
		return types.ObjectNull(tagFilterAttrTypes), diags
	}

	if len(elements) > 1 {
		diags.AddAttributeWarning(
			p,
			"Multiple Tag Filter Blocks Discarded",
			fmt.Sprintf(
				"The prior state held %d '%s' blocks, but only one is supported. One block was kept and the others were discarded. "+
					"Review the plan before applying and remove the extra blocks from your configuration.",
				len(elements), p.String(),
			),
		)
	}

	// Normalise a null tags list to an empty list so the upgraded state cannot
	// differ from a configuration that writes tags = [].
	attrs := obj.Attributes()
	if tagsAttr, found := attrs["tags"]; !found || tagsAttr == nil || tagsAttr.IsNull() {
		emptyTags, d := types.ListValueFrom(ctx, types.StringType, []string{})
		diags.Append(d...)
		if diags.HasError() {
			return types.ObjectNull(tagFilterAttrTypes), diags
		}
		normalised := map[string]attr.Value{
			"relation": types.StringNull(),
			"tags":     emptyTags,
		}
		if relation, found := attrs["relation"]; found && relation != nil {
			normalised["relation"] = relation
		}
		newObj, d := types.ObjectValue(tagFilterAttrTypes, normalised)
		diags.Append(d...)
		if diags.HasError() {
			return types.ObjectNull(tagFilterAttrTypes), diags
		}
		return newObj, diags
	}

	return obj, diags
}
