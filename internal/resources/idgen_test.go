package resources

import (
	"regexp"
	"testing"
)

func TestGenerateID(t *testing.T) {
	id := generateID()
	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	if !uuidRegex.MatchString(id) {
		t.Errorf("generateID() = %q, does not match UUID format", id)
	}
}

func TestGenerateIDNoDash(t *testing.T) {
	id := generateIDNoDash()
	noDashRegex := regexp.MustCompile(`^[0-9a-f]{32}$`)
	if !noDashRegex.MatchString(id) {
		t.Errorf("generateIDNoDash() = %q, does not match 32-char hex format", id)
	}
}

func TestGenerateIDNoDashContainsNoDash(t *testing.T) {
	id := generateIDNoDash()
	for _, c := range id {
		if c == '-' {
			t.Errorf("generateIDNoDash() = %q, contains a dash", id)
			break
		}
	}
}

func TestGenerateCertificateID(t *testing.T) {
	id := generateCertificateID()
	if len(id) != 33 {
		t.Errorf("generateCertificateID() = %q, expected length 33 but got %d", id, len(id))
	}
	if id[0] != 'c' {
		t.Errorf("generateCertificateID() = %q, expected to start with 'c'", id)
	}
	certRegex := regexp.MustCompile(`^[a-z](?:[-a-z0-9]{0,61}[a-z0-9])?$`)
	if !certRegex.MatchString(id) {
		t.Errorf("generateCertificateID() = %q, does not match certificate ID regex", id)
	}
}

func TestGenerateIDUniqueness(t *testing.T) {
	id1 := generateID()
	id2 := generateID()
	if id1 == id2 {
		t.Errorf("generateID() returned same value twice: %q", id1)
	}
}

func TestGenerateIDNoDashMatchesEdgeFunctionRegex(t *testing.T) {
	id := generateIDNoDash()
	edgeFuncRegex := regexp.MustCompile(`^[A-Za-z0-9_]*$`)
	if !edgeFuncRegex.MatchString(id) {
		t.Errorf("generateIDNoDash() = %q, does not match edge_function regex ^[A-Za-z0-9_]*$", id)
	}
}

func TestGenerateIDMatchesServerGroupRegex(t *testing.T) {
	id := generateID()
	serverGroupRegex := regexp.MustCompile(`^[A-Za-z0-9\-\_]*$`)
	if !serverGroupRegex.MatchString(id) {
		t.Errorf("generateID() = %q, does not match server_group regex ^[A-Za-z0-9\\-\\_]*$", id)
	}
}
