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
