// Package providerutil provides shared utilities for Terraform provider configuration.
package providerutil

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"

	"github.com/link11/terraform-provider-link11waap/internal/client"
)

// ConfigureClient extracts and validates the *client.Client from provider data.
// Returns nil if providerData is nil (provider not yet configured).
// Adds a diagnostic error and returns nil if the type assertion fails.
func ConfigureClient(providerData any, diags *diag.Diagnostics) *client.Client {
	if providerData == nil {
		return nil
	}

	c, ok := providerData.(*client.Client)
	if !ok {
		diags.AddError(
			"Client Provider Data Error",
			fmt.Sprintf("invalid provider data supplied, expected *client.Client, got %T", providerData),
		)
		return nil
	}

	return c
}
