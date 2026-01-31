// verify_test.go - Cross-module import verification for fiberfx package
package branded_test

import (
	"testing"

	"github.com/jonny-novikov/jonnify/fiberfx"
)

// TestIDParseFormat verifies parse/format round-trip
func TestIDParseFormat(t *testing.T) {
	original := "TSK0Ij1P13FRDM"
	id, err := fiberfx.Parse(original)
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if id.String() != original {
		t.Errorf("Expected %q, got %q", original, id.String())
	}
}

// TestNamespaceRegistry verifies namespace registration
func TestNamespaceRegistry(t *testing.T) {
	// Pre-registered namespaces
	namespaces := []fiberfx.Namespace{
		fiberfx.NS_TASK,
		fiberfx.NS_USER,
		fiberfx.NS_SESSION,
	}
	for _, ns := range namespaces {
		if !ns.Valid() {
			t.Errorf("Namespace %q should be valid", ns)
		}
		if len(ns) != 3 {
			t.Errorf("Namespace %q should be 3 characters", ns)
		}
	}
}

// TestGenerator verifies ID generation
func TestGenerator(t *testing.T) {
	gen := fiberfx.NewGenerator(1)
	id1 := gen.New(fiberfx.NS_TASK)
	id2 := gen.New(fiberfx.NS_TASK)

	if id1.String() == id2.String() {
		t.Error("Generated IDs should be unique")
	}
	if id1.Namespace() != fiberfx.NS_TASK {
		t.Errorf("Expected namespace TSK, got %s", id1.Namespace())
	}
}

// TestDefaultGenerate verifies default generator
func TestDefaultGenerate(t *testing.T) {
	id := fiberfx.Generate(fiberfx.NS_USER)
	if id.IsZero() {
		t.Error("Generated ID should not be zero")
	}
	if id.Namespace() != fiberfx.NS_USER {
		t.Errorf("Expected namespace USR, got %s", id.Namespace())
	}
}

// TestCodemojexNamespaces verifies codemojex domain namespaces
func TestCodemojexNamespaces(t *testing.T) {
	// Players domain
	namespaces := []struct {
		ns       fiberfx.Namespace
		expected string
	}{
		{fiberfx.NS_PLAYER, "PLR"},
		{fiberfx.NS_PLAYER_RESOURCE, "RSC"},
		{fiberfx.NS_ROOM, "ROM"},
		{fiberfx.NS_GAME, "GAM"},
		{fiberfx.NS_GUESS, "GUS"},
		{fiberfx.NS_TRANSACTION, "TXN"},
		{fiberfx.NS_DEPLOYMENT, "DPL"},
		{fiberfx.NS_SHARE, "SHR"},
	}

	for _, tc := range namespaces {
		if string(tc.ns) != tc.expected {
			t.Errorf("Expected %q, got %q", tc.expected, tc.ns)
		}
		if !tc.ns.Valid() {
			t.Errorf("Namespace %q should be valid", tc.ns)
		}
	}
}

// TestCodemojexIDGeneration verifies ID generation with codemojex namespaces
func TestCodemojexIDGeneration(t *testing.T) {
	gen := fiberfx.NewGenerator(1)

	// Generate player ID
	playerID := gen.New(fiberfx.NS_PLAYER)
	if playerID.Namespace() != fiberfx.NS_PLAYER {
		t.Errorf("Expected PLR namespace, got %s", playerID.Namespace())
	}
	if len(playerID.String()) != 14 {
		t.Errorf("Expected 14 chars, got %d", len(playerID.String()))
	}

	// Generate room ID
	roomID := gen.New(fiberfx.NS_ROOM)
	if roomID.Namespace() != fiberfx.NS_ROOM {
		t.Errorf("Expected ROM namespace, got %s", roomID.Namespace())
	}

	// IDs should be unique
	if playerID.String() == roomID.String() {
		t.Error("Generated IDs should be unique")
	}
}
