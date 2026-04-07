package resources

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-uuid"
)

// generateID generates a new UUID-based resource ID with dashes.
func generateID() string {
	id, err := uuid.GenerateUUID()
	if err != nil {
		panic(fmt.Sprintf("failed to generate UUID: %v", err))
	}
	return id
}

// generateIDNoDash generates a new UUID-based resource ID without dashes.
// Used for edge_function resources that don't allow dashes in IDs.
func generateIDNoDash() string {
	id, err := uuid.GenerateUUID()
	if err != nil {
		panic(fmt.Sprintf("failed to generate UUID: %v", err))
	}
	return strings.ReplaceAll(id, "-", "")
}

// generateCertificateID generates a new ID suitable for certificate resources.
// Certificate IDs must match ^[a-z](?:[-a-z0-9]{0,61}[a-z0-9])?$ so must start with a letter.
func generateCertificateID() string {
	id, err := uuid.GenerateUUID()
	if err != nil {
		panic(fmt.Sprintf("failed to generate UUID: %v", err))
	}
	return "c" + strings.ReplaceAll(id, "-", "")
}
