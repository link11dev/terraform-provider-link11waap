package resources

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// emptyListFor decides what an empty list looks like in state.
//
// Terraform distinguishes an unset list from an explicitly empty one, and for an
// Optional attribute that is not Computed it requires state to match
// configuration exactly. The provider's round trip does not preserve that
// distinction. Most list fields in internal/client carry `omitempty`, so an
// empty slice is never transmitted. The section matcher lists (names, regex,
// text) do not, and send `[]` or null as configured, but the flatten code only
// looks at the length of what comes back, so an absent key, null and `[]` all
// read as the same zero-length slice. (JSON itself can tell nil from `[]`;
// relying on that would be the alternative fix, but it changes what goes on
// the wire and it is unverified whether the API stores and echoes an empty
// array distinctly from an absent one.)
//
// So the distinction is recovered from the prior value instead — the plan on
// create/update, the previous state on read. This function mirrors it: a
// configuration that said `[]` keeps an empty list, while an unset, unknown, or
// previously populated value becomes null. Returning null unconditionally makes
// `x = []` fail with "Provider produced inconsistent result after apply" on
// create and re-plan forever on refresh; returning an empty list
// unconditionally breaks the opposite case, where the attribute was never set.
//
// A prior value that still has elements yields null, so a genuine upstream
// deletion is reported as drift rather than silently preserved.
//
// # Wiring it into a resource
//
// Which call sites need it depends on how the resource builds its state:
//
//   - Resources that set state straight from the plan on create and update, and
//     only map the API struct in Read, need it in Read alone. acl_profile.go is
//     the example. Note that this shape also hides any server-side
//     normalization until the next refresh, which is a separate problem with it.
//   - Resources that map the API struct into state on create and update as well
//     — because they have provider- or server-generated values to propagate,
//     such as the per-entry ids in content_filter_profile.go — need the prior
//     value threaded through the whole flatten path, since flatten then runs on
//     all three operations.
//
// Attributes that are Computed without being Optional do not need it: Terraform
// does not compare those against configuration. Neither do data sources.
func emptyListFor(prior types.List, elemType attr.Type) types.List {
	if !prior.IsNull() && !prior.IsUnknown() && len(prior.Elements()) == 0 {
		return types.ListValueMust(elemType, []attr.Value{})
	}
	return types.ListNull(elemType)
}
