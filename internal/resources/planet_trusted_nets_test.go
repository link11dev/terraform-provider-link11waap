package resources

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/link11/terraform-provider-link11waap/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var trustedNetsObjType = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"source":  tftypes.String,
		"address": tftypes.String,
		"gf_id":   tftypes.String,
		"comment": tftypes.String,
	},
}

var trustedNetsListType = tftypes.List{ElementType: trustedNetsObjType}

func makeTrustedNet(source, address, gfID, comment tftypes.Value) tftypes.Value {
	return tftypes.NewValue(trustedNetsObjType, map[string]tftypes.Value{
		"source":  source,
		"address": address,
		"gf_id":   gfID,
		"comment": comment,
	})
}

func strVal(s string) tftypes.Value {
	return tftypes.NewValue(tftypes.String, s)
}

func TestNewPlanetTrustedNetsResource(t *testing.T) {
	r := NewPlanetTrustedNetsResource()
	require.NotNil(t, r)
	_, ok := r.(*PlanetTrustedNetsResource)
	assert.True(t, ok)
}

func TestPlanetTrustedNetsResource_Metadata(t *testing.T) {
	r := &PlanetTrustedNetsResource{}
	req := metadataReq("link11waap")
	resp := metadataResp()
	r.Metadata(context.Background(), req, resp)
	assert.Equal(t, "link11waap_planet_trusted_nets", resp.TypeName)
}

func TestPlanetTrustedNetsResource_Schema(t *testing.T) {
	r := &PlanetTrustedNetsResource{}
	sReq := schemaReq()
	sResp := schemaResp()
	r.Schema(context.Background(), sReq, sResp)

	for _, name := range []string{"config_id", "id", "name"} {
		_, ok := sResp.Schema.Attributes[name]
		assert.True(t, ok, "expected attribute %q", name)
	}

	cfgAttr, ok := sResp.Schema.Attributes["config_id"].(rschema.StringAttribute)
	require.True(t, ok)
	assert.True(t, cfgAttr.Required)

	block, ok := sResp.Schema.Blocks["trusted_nets"].(rschema.ListNestedBlock)
	require.True(t, ok, "trusted_nets should be a ListNestedBlock")

	sourceAttr, ok := block.NestedObject.Attributes["source"].(rschema.StringAttribute)
	require.True(t, ok)
	assert.True(t, sourceAttr.Required)
	assert.NotEmpty(t, sourceAttr.Validators, "source must have OneOf validator")

	for _, name := range []string{"address", "gf_id", "comment"} {
		attr, ok := block.NestedObject.Attributes[name].(rschema.StringAttribute)
		require.True(t, ok, "expected nested attribute %q", name)
		assert.True(t, attr.Optional, "%s should be Optional", name)
		assert.True(t, attr.Computed, "%s should be Computed", name)
		assert.NotNil(t, attr.Default, "%s should have default", name)
	}
}

func TestPlanetTrustedNetsResource_ConfigValidators(t *testing.T) {
	r := &PlanetTrustedNetsResource{}
	validators := r.ConfigValidators(context.Background())
	require.Len(t, validators, 1)
	_, ok := validators[0].(*trustedNetsConfigValidator)
	assert.True(t, ok)
}

func TestTrustedNetsConfigValidator_Description(t *testing.T) {
	v := &trustedNetsConfigValidator{}
	assert.NotEmpty(t, v.Description(context.Background()))
}

func TestTrustedNetsConfigValidator_MarkdownDescription(t *testing.T) {
	v := &trustedNetsConfigValidator{}
	assert.NotEmpty(t, v.MarkdownDescription(context.Background()))
}

func TestTrustedNetsConfigValidator_ValidateResource(t *testing.T) {
	r := &PlanetTrustedNetsResource{}
	v := &trustedNetsConfigValidator{}
	ctx := context.Background()

	testCases := []struct {
		name      string
		entry     tftypes.Value
		expectErr bool
	}{
		{
			name:  "valid ip with address",
			entry: makeTrustedNet(strVal("ip"), strVal("1.2.3.4"), strVal(""), strVal("")),
		},
		{
			name:  "valid ip with cidr",
			entry: makeTrustedNet(strVal("ip"), strVal("10.0.0.0/8"), strVal(""), strVal("")),
		},
		{
			name:      "ip with empty address",
			entry:     makeTrustedNet(strVal("ip"), strVal(""), strVal(""), strVal("")),
			expectErr: true,
		},
		{
			name:      "ip with non-empty gf_id",
			entry:     makeTrustedNet(strVal("ip"), strVal("1.2.3.4"), strVal("gf1"), strVal("")),
			expectErr: true,
		},
		{
			name:      "ip with invalid address",
			entry:     makeTrustedNet(strVal("ip"), strVal("not-an-ip"), strVal(""), strVal("")),
			expectErr: true,
		},
		{
			name:  "valid global_filter",
			entry: makeTrustedNet(strVal("global_filter"), strVal(""), strVal("gf1"), strVal("")),
		},
		{
			name:      "global_filter with address",
			entry:     makeTrustedNet(strVal("global_filter"), strVal("1.2.3.4"), strVal("gf1"), strVal("")),
			expectErr: true,
		},
		{
			name:      "global_filter with empty gf_id",
			entry:     makeTrustedNet(strVal("global_filter"), strVal(""), strVal(""), strVal("")),
			expectErr: true,
		},
		{
			name: "unknown source is skipped",
			entry: makeTrustedNet(
				tftypes.NewValue(tftypes.String, tftypes.UnknownValue),
				strVal(""), strVal(""), strVal(""),
			),
		},
		{
			name: "ip with unknown address is allowed",
			entry: makeTrustedNet(
				strVal("ip"),
				tftypes.NewValue(tftypes.String, tftypes.UnknownValue), // unknown address
				strVal(""),
				strVal(""),
			),
		},
		{
			name: "global_filter with unknown gf_id is allowed",
			entry: makeTrustedNet(
				strVal("global_filter"),
				strVal(""),
				tftypes.NewValue(tftypes.String, tftypes.UnknownValue), // unknown gf_id
				strVal(""),
			),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := buildConfig(ctx, t, r, map[string]tftypes.Value{
				"config_id":    strVal("cfg1"),
				"trusted_nets": tftypes.NewValue(trustedNetsListType, []tftypes.Value{tc.entry}),
			})
			req := resource.ValidateConfigRequest{Config: config}
			resp := &resource.ValidateConfigResponse{}
			v.ValidateResource(ctx, req, resp)

			if tc.expectErr {
				assert.True(t, resp.Diagnostics.HasError(), "expected error, got: %v", resp.Diagnostics)
			} else {
				assert.False(t, resp.Diagnostics.HasError(), "unexpected errors: %v", resp.Diagnostics)
			}
		})
	}
}

func TestPlanetTrustedNetsResource_Configure_NilProvider(t *testing.T) {
	r := &PlanetTrustedNetsResource{}
	req := configureReq(nil)
	resp := configureResp()
	r.Configure(context.Background(), req, resp)
	assert.Nil(t, r.client)
	assert.False(t, resp.Diagnostics.HasError())
}

func TestPlanetTrustedNetsResource_Configure_InvalidType(t *testing.T) {
	testConfigureWithInvalidType(t, &PlanetTrustedNetsResource{})
}

func TestPlanetTrustedNetsResource_Create_Success(t *testing.T) {
	r := &PlanetTrustedNetsResource{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		assert.Equal(t, http.MethodPut, req.Method)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	})
	configureResourceWithMock(t, r, handler)

	ctx := context.Background()
	plan := buildTerraformPlan(ctx, t, r, map[string]tftypes.Value{
		"config_id": strVal("cfg1"),
		"trusted_nets": tftypes.NewValue(trustedNetsListType, []tftypes.Value{
			makeTrustedNet(strVal("ip"), strVal("1.2.3.4"), strVal(""), strVal("office")),
		}),
	})

	sReq := schemaReq()
	sResp := schemaResp()
	r.Schema(ctx, sReq, sResp)

	emptyState := tfsdk.State{
		Schema: sResp.Schema,
		Raw:    tftypes.NewValue(sResp.Schema.Type().TerraformType(ctx), nil),
	}
	createResp := &resource.CreateResponse{State: emptyState}
	r.Create(ctx, resource.CreateRequest{Plan: plan}, createResp)
	require.False(t, createResp.Diagnostics.HasError(), "errors: %v", createResp.Diagnostics)

	var got PlanetTrustedNetsResourceModel
	diags := createResp.State.Get(ctx, &got)
	require.False(t, diags.HasError())
	assert.Equal(t, defaultPlanetEntryID, got.ID.ValueString())
	assert.Equal(t, defaultPlanetEntryID, got.Name.ValueString())
}

func TestPlanetTrustedNetsResource_Create_APIError(t *testing.T) {
	r := &PlanetTrustedNetsResource{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"boom"}`))
	})
	configureResourceWithMock(t, r, handler)

	ctx := context.Background()
	plan := buildTerraformPlan(ctx, t, r, map[string]tftypes.Value{
		"config_id":    strVal("cfg1"),
		"trusted_nets": tftypes.NewValue(trustedNetsListType, []tftypes.Value{}),
	})

	sReq := schemaReq()
	sResp := schemaResp()
	r.Schema(ctx, sReq, sResp)

	emptyState := tfsdk.State{
		Schema: sResp.Schema,
		Raw:    tftypes.NewValue(sResp.Schema.Type().TerraformType(ctx), nil),
	}
	createResp := &resource.CreateResponse{State: emptyState}
	r.Create(ctx, resource.CreateRequest{Plan: plan}, createResp)
	assert.True(t, createResp.Diagnostics.HasError())
}

func TestPlanetTrustedNetsResource_Read_Success(t *testing.T) {
	r := &PlanetTrustedNetsResource{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(client.ListResponse[client.Planet]{
			Total: 1,
			Items: []client.Planet{
				{
					ID:   "__default__",
					Name: "__default__",
					TrustedNets: []client.TrustedNet{
						{Source: "ip", Address: "1.2.3.4", Comment: "office"},
					},
				},
			},
		})
	})
	configureResourceWithMock(t, r, handler)

	resp := readWithMock(t, r, map[string]tftypes.Value{
		"config_id":    strVal("cfg1"),
		"id":           strVal(defaultPlanetEntryID),
		"name":         strVal(defaultPlanetEntryID),
		"trusted_nets": tftypes.NewValue(trustedNetsListType, []tftypes.Value{}),
	})
	assert.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)
}

func TestPlanetTrustedNetsResource_Read_NotFound(t *testing.T) {
	r := &PlanetTrustedNetsResource{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(client.ListResponse[client.Planet]{
			Total: 0,
			Items: []client.Planet{},
		})
	})
	configureResourceWithMock(t, r, handler)

	resp := readWithMock(t, r, map[string]tftypes.Value{
		"config_id":    strVal("cfg1"),
		"id":           strVal(defaultPlanetEntryID),
		"name":         strVal(defaultPlanetEntryID),
		"trusted_nets": tftypes.NewValue(trustedNetsListType, []tftypes.Value{}),
	})
	assert.False(t, resp.Diagnostics.HasError())
}

func TestPlanetTrustedNetsResource_Read_APIError(t *testing.T) {
	r := &PlanetTrustedNetsResource{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"boom"}`))
	})
	configureResourceWithMock(t, r, handler)

	resp := readWithMock(t, r, map[string]tftypes.Value{
		"config_id":    strVal("cfg1"),
		"id":           strVal(defaultPlanetEntryID),
		"name":         strVal(defaultPlanetEntryID),
		"trusted_nets": tftypes.NewValue(trustedNetsListType, []tftypes.Value{}),
	})
	assert.True(t, resp.Diagnostics.HasError())
}

func TestPlanetTrustedNetsResource_Update_Success(t *testing.T) {
	r := &PlanetTrustedNetsResource{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		assert.Equal(t, http.MethodPut, req.Method)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	})
	configureResourceWithMock(t, r, handler)

	ctx := context.Background()
	plan := buildTerraformPlan(ctx, t, r, map[string]tftypes.Value{
		"config_id": strVal("cfg1"),
		"trusted_nets": tftypes.NewValue(trustedNetsListType, []tftypes.Value{
			makeTrustedNet(strVal("ip"), strVal("2.3.4.5"), strVal(""), strVal("updated")),
		}),
	})

	sReq := schemaReq()
	sResp := schemaResp()
	r.Schema(ctx, sReq, sResp)

	state := tfsdk.State{
		Schema: sResp.Schema,
		Raw:    tftypes.NewValue(sResp.Schema.Type().TerraformType(ctx), nil),
	}
	updateResp := &resource.UpdateResponse{State: state}
	r.Update(ctx, resource.UpdateRequest{Plan: plan}, updateResp)
	require.False(t, updateResp.Diagnostics.HasError(), "errors: %v", updateResp.Diagnostics)

	var got PlanetTrustedNetsResourceModel
	diags := updateResp.State.Get(ctx, &got)
	require.False(t, diags.HasError())
	assert.Equal(t, defaultPlanetEntryID, got.ID.ValueString())
	assert.Equal(t, defaultPlanetEntryID, got.Name.ValueString())
}

func TestPlanetTrustedNetsResource_Update_APIError(t *testing.T) {
	r := &PlanetTrustedNetsResource{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"boom"}`))
	})
	configureResourceWithMock(t, r, handler)

	ctx := context.Background()
	plan := buildTerraformPlan(ctx, t, r, map[string]tftypes.Value{
		"config_id":    strVal("cfg1"),
		"trusted_nets": tftypes.NewValue(trustedNetsListType, []tftypes.Value{}),
	})

	sReq := schemaReq()
	sResp := schemaResp()
	r.Schema(ctx, sReq, sResp)

	state := tfsdk.State{
		Schema: sResp.Schema,
		Raw:    tftypes.NewValue(sResp.Schema.Type().TerraformType(ctx), nil),
	}
	updateResp := &resource.UpdateResponse{State: state}
	r.Update(ctx, resource.UpdateRequest{Plan: plan}, updateResp)
	assert.True(t, updateResp.Diagnostics.HasError())
}

func TestPlanetTrustedNetsResource_Delete_NoOp(t *testing.T) {
	r := &PlanetTrustedNetsResource{}
	ctx := context.Background()
	state := buildTerraformState(ctx, t, r, map[string]tftypes.Value{
		"config_id":    strVal("cfg1"),
		"trusted_nets": tftypes.NewValue(trustedNetsListType, []tftypes.Value{}),
	})
	resp := &resource.DeleteResponse{}
	assert.NotPanics(t, func() {
		r.Delete(ctx, resource.DeleteRequest{State: state}, resp)
	})
	assert.False(t, resp.Diagnostics.HasError())
}

func TestPlanetTrustedNetsResource_ImportState(t *testing.T) {
	r := &PlanetTrustedNetsResource{}
	resp := testImportState(t, r, "cfg123")
	require.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)

	ctx := context.Background()
	var got PlanetTrustedNetsResourceModel
	diags := resp.State.Get(ctx, &got)
	require.False(t, diags.HasError())
	assert.Equal(t, "cfg123", got.ConfigID.ValueString())
	assert.Equal(t, defaultPlanetEntryID, got.ID.ValueString())
	assert.Equal(t, defaultPlanetEntryID, got.Name.ValueString())
}

func TestPlanetTrustedNetsResource_ImportState_WithDefaultSuffix(t *testing.T) {
	r := &PlanetTrustedNetsResource{}
	resp := testImportState(t, r, "cfg123/__default__")
	require.False(t, resp.Diagnostics.HasError(), "errors: %v", resp.Diagnostics)

	ctx := context.Background()
	var got PlanetTrustedNetsResourceModel
	diags := resp.State.Get(ctx, &got)
	require.False(t, diags.HasError())
	assert.Equal(t, "cfg123", got.ConfigID.ValueString())
	assert.Equal(t, defaultPlanetEntryID, got.ID.ValueString())
	assert.Equal(t, defaultPlanetEntryID, got.Name.ValueString())
}

func TestPlanetTrustedNetsResource_ImportState_InvalidSuffix(t *testing.T) {
	r := &PlanetTrustedNetsResource{}
	resp := testImportState(t, r, "cfg123/not-default")
	assert.True(t, resp.Diagnostics.HasError())
}

func TestBuildPlanetBody(t *testing.T) {
	plan := &PlanetTrustedNetsResourceModel{
		ConfigID: types.StringValue("cfg1"),
		TrustedNets: []TrustedNetModel{
			{
				Source:  types.StringValue("ip"),
				Address: types.StringValue("1.2.3.4"),
				GfID:    types.StringValue(""),
				Comment: types.StringValue("office"),
			},
			{
				Source:  types.StringValue("global_filter"),
				Address: types.StringValue(""),
				GfID:    types.StringValue("gf1"),
				Comment: types.StringValue("filter"),
			},
		},
	}

	planet := buildPlanetBody(plan)
	require.NotNil(t, planet)
	assert.Equal(t, defaultPlanetEntryID, planet.ID)
	assert.Equal(t, defaultPlanetEntryID, planet.Name)
	assert.Equal(t, json.RawMessage(defaultIchallengeJSON), planet.Ichallenge)
	assert.Equal(t, defaultNoHostCertName, planet.NoHostCertName)
	assert.Equal(t, defaultNoHostSSLCiphers, planet.NoHostSSLCiphers)
	assert.Equal(t, []string{"TLSv1.2", "TLSv1.3"}, planet.NoHostSSLProtocols)

	require.Len(t, planet.TrustedNets, 2)
	assert.Equal(t, "ip", planet.TrustedNets[0].Source)
	assert.Equal(t, "1.2.3.4", planet.TrustedNets[0].Address)
	assert.Equal(t, "", planet.TrustedNets[0].GfID)
	assert.Equal(t, "office", planet.TrustedNets[0].Comment)

	assert.Equal(t, "global_filter", planet.TrustedNets[1].Source)
	assert.Equal(t, "", planet.TrustedNets[1].Address)
	assert.Equal(t, "gf1", planet.TrustedNets[1].GfID)
	assert.Equal(t, "filter", planet.TrustedNets[1].Comment)
}

func TestBuildPlanetBody_EmptyTrustedNets(t *testing.T) {
	plan := &PlanetTrustedNetsResourceModel{
		ConfigID:    types.StringValue("cfg1"),
		TrustedNets: []TrustedNetModel{},
	}

	planet := buildPlanetBody(plan)
	require.NotNil(t, planet)
	assert.Empty(t, planet.TrustedNets)
	assert.NotNil(t, planet.TrustedNets)
}
