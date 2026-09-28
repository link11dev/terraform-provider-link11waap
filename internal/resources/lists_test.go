package resources

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func TestEmptyListFor(t *testing.T) {
	objType := types.ObjectType{AttrTypes: map[string]attr.Type{"k": types.StringType}}

	tests := []struct {
		name      string
		prior     types.List
		elemType  attr.Type
		wantNull  bool
		wantElems int
	}{
		{
			name:     "prior null stays null",
			prior:    types.ListNull(types.StringType),
			elemType: types.StringType,
			wantNull: true,
		},
		{
			name:     "prior unknown becomes null",
			prior:    types.ListUnknown(types.StringType),
			elemType: types.StringType,
			wantNull: true,
		},
		{
			name:      "prior empty stays empty",
			prior:     types.ListValueMust(types.StringType, []attr.Value{}),
			elemType:  types.StringType,
			wantNull:  false,
			wantElems: 0,
		},
		{
			name:     "prior populated becomes null so drift is reported",
			prior:    types.ListValueMust(types.StringType, []attr.Value{types.StringValue("a")}),
			elemType: types.StringType,
			wantNull: true,
		},
		{
			name:     "zero value list is treated as null",
			prior:    types.List{},
			elemType: types.StringType,
			wantNull: true,
		},
		{
			name:      "element type is carried through for object lists",
			prior:     types.ListValueMust(objType, []attr.Value{}),
			elemType:  objType,
			wantNull:  false,
			wantElems: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := emptyListFor(tt.prior, tt.elemType)

			assert.Equal(t, tt.wantNull, got.IsNull())
			assert.False(t, got.IsUnknown(), "result must always be known")
			assert.Equal(t, tt.elemType, got.ElementType(t.Context()))
			if !tt.wantNull {
				assert.Len(t, got.Elements(), tt.wantElems)
			}
		})
	}
}
