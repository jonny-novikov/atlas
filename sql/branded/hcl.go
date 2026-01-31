// Copyright 2021-present The Atlas Authors. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package branded

import (
	"fmt"
	"reflect"

	"ariga.io/atlas/schemahcl"
	"ariga.io/atlas/sql/schema"
	"github.com/jonny-novikov/jonnify/fiberfx"
)

// TypeBrandedID is the HCL type name for branded IDs.
const TypeBrandedID = "branded_id"

// BrandedIDTypeSpec returns the TypeSpec for the branded_id HCL type.
// The type is used as: branded_id("TSK"), branded_id("EPC"), etc.
func BrandedIDTypeSpec() *schemahcl.TypeSpec {
	return schemahcl.NewTypeSpec(
		TypeBrandedID,
		schemahcl.WithAttributes(&schemahcl.TypeAttr{
			Name:     "namespace",
			Kind:     reflect.String,
			Required: true,
		}),
		schemahcl.WithFromSpec(fromSpec),
		schemahcl.WithToSpec(toSpec),
	)
}

// fromSpec converts a schemahcl.Type to a schema.Type (HCL → database).
// Input: branded_id("TSK") from HCL
// Output: *BrandedIDType{Namespace: "TSK"}
func fromSpec(t *schemahcl.Type) (schema.Type, error) {
	if t.T != TypeBrandedID {
		return nil, fmt.Errorf("branded: expected type %q, got %q", TypeBrandedID, t.T)
	}
	// Get the namespace attribute (first positional argument)
	ns, err := namespaceFromAttrs(t.Attrs)
	if err != nil {
		return nil, err
	}
	// Validate the namespace
	if !fiberfx.IsValidNamespace(fiberfx.Namespace(ns)) {
		return nil, fmt.Errorf("branded: invalid namespace %q; valid namespaces: %v", ns, fiberfx.AllNamespaces())
	}
	return BrandedID(ns), nil
}

// toSpec converts a schema.Type to a schemahcl.Type (database → HCL).
// Input: *BrandedIDType{Namespace: "TSK"}
// Output: branded_id("TSK") for HCL
func toSpec(t schema.Type) (*schemahcl.Type, error) {
	bid, ok := t.(*BrandedIDType)
	if !ok {
		return nil, fmt.Errorf("branded: expected *BrandedIDType, got %T", t)
	}
	return &schemahcl.Type{
		T: TypeBrandedID,
		Attrs: []*schemahcl.Attr{
			schemahcl.StringAttr("namespace", string(bid.Namespace)),
		},
	}, nil
}

// namespaceFromAttrs extracts the namespace from HCL type attributes.
// The namespace is the first positional argument: branded_id("TSK").
func namespaceFromAttrs(attrs []*schemahcl.Attr) (string, error) {
	for _, a := range attrs {
		if a.K == "namespace" {
			return a.String()
		}
	}
	return "", fmt.Errorf("branded: missing namespace attribute")
}

// IsBrandedIDType checks if a schemahcl.Type represents a branded ID.
func IsBrandedIDType(t *schemahcl.Type) bool {
	return t != nil && t.T == TypeBrandedID
}

// NamespaceFromType extracts the namespace from a schemahcl.Type.
// Returns empty string if the type is not a branded ID or has no namespace.
func NamespaceFromType(t *schemahcl.Type) string {
	if !IsBrandedIDType(t) {
		return ""
	}
	ns, _ := namespaceFromAttrs(t.Attrs)
	return ns
}
