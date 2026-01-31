// Copyright 2021-present The Atlas Authors. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package postgres

import (
	"testing"

	"ariga.io/atlas/schemahcl"
	"ariga.io/atlas/sql/branded"
	"github.com/jonny-novikov/jonnify/fiberfx"
	"github.com/stretchr/testify/require"
)

func TestTypeRegistryBrandedID(t *testing.T) {
	// Verify branded_id TypeSpec is registered
	specs := TypeRegistry.Specs()
	var found bool
	for _, s := range specs {
		if s.Name == branded.TypeBrandedID {
			found = true
			require.Equal(t, branded.TypeBrandedID, s.T)
			require.Len(t, s.Attributes, 1)
			require.Equal(t, "namespace", s.Attributes[0].Name)
			break
		}
	}
	require.True(t, found, "branded_id TypeSpec should be registered")
}

func TestTypeRegistryBrandedIDConvert(t *testing.T) {
	// Test converting BrandedIDType to schemahcl.Type
	bid := branded.BrandedID("TSK")

	spec, err := TypeRegistry.Convert(bid)
	require.NoError(t, err)
	require.Equal(t, branded.TypeBrandedID, spec.T)
	require.Len(t, spec.Attrs, 1)
	require.Equal(t, "namespace", spec.Attrs[0].K)

	ns, err := spec.Attrs[0].String()
	require.NoError(t, err)
	require.Equal(t, "TSK", ns)
}

func TestTypeRegistryBrandedIDType(t *testing.T) {
	// Test converting schemahcl.Type back to BrandedIDType
	spec := &schemahcl.Type{
		T:     branded.TypeBrandedID,
		Attrs: []*schemahcl.Attr{schemahcl.StringAttr("namespace", "EPC")},
	}

	result, err := TypeRegistry.Type(spec, nil)
	require.NoError(t, err)

	bid, ok := result.(*branded.BrandedIDType)
	require.True(t, ok)
	require.Equal(t, fiberfx.NS_EPIC, bid.Namespace)
}

func TestTypeRegistryBrandedIDRoundTrip(t *testing.T) {
	namespaces := []fiberfx.Namespace{
		fiberfx.NS_TASK,
		fiberfx.NS_EPIC,
		fiberfx.NS_FEATURE,
		fiberfx.NS_PLAN,
	}

	for _, ns := range namespaces {
		t.Run(string(ns), func(t *testing.T) {
			// Create original BrandedIDType
			original := branded.BrandedIDFromNamespace(ns)

			// Convert to schemahcl.Type
			spec, err := TypeRegistry.Convert(original)
			require.NoError(t, err)
			require.Equal(t, branded.TypeBrandedID, spec.T)

			// Convert back to schema.Type
			result, err := TypeRegistry.Type(spec, nil)
			require.NoError(t, err)

			// Verify round-trip preserves namespace
			bid, ok := result.(*branded.BrandedIDType)
			require.True(t, ok)
			require.Equal(t, original.Namespace, bid.Namespace)
		})
	}
}

func TestHCLParseBrandedID(t *testing.T) {
	// Test parsing HCL with branded_id type
	hcl := `
table "tasks" {
  column "id" {
    type = branded_id("TSK")
  }
  column "epic_id" {
    type = branded_id("EPC")
    null = true
  }
  column "title" {
    type = varchar(255)
  }
}
`
	type doc struct {
		Tables []*struct {
			Name    string `spec:",name"`
			Columns []*struct {
				Name string          `spec:",name"`
				Type *schemahcl.Type `spec:"type"`
				Null bool            `spec:"null"`
			} `spec:"column"`
		} `spec:"table"`
	}

	var d doc
	err := schemahcl.New(
		schemahcl.WithTypes("table.column.type", TypeRegistry.Specs()),
	).EvalBytes([]byte(hcl), &d, nil)
	require.NoError(t, err)
	require.Len(t, d.Tables, 1)

	table := d.Tables[0]
	require.Equal(t, "tasks", table.Name)
	require.Len(t, table.Columns, 3)

	// Verify id column
	idCol := table.Columns[0]
	require.Equal(t, "id", idCol.Name)
	require.True(t, branded.IsBrandedIDType(idCol.Type))
	require.Equal(t, "TSK", branded.NamespaceFromType(idCol.Type))

	// Verify epic_id column
	epicCol := table.Columns[1]
	require.Equal(t, "epic_id", epicCol.Name)
	require.True(t, branded.IsBrandedIDType(epicCol.Type))
	require.Equal(t, "EPC", branded.NamespaceFromType(epicCol.Type))
	require.True(t, epicCol.Null)

	// Verify title column (not branded)
	titleCol := table.Columns[2]
	require.Equal(t, "title", titleCol.Name)
	require.False(t, branded.IsBrandedIDType(titleCol.Type))
}

func TestHCLMarshalBrandedID(t *testing.T) {
	// Test that we can marshal a schema with branded_id back to HCL
	bid := branded.BrandedID("TSK")

	spec, err := TypeRegistry.Convert(bid)
	require.NoError(t, err)

	// The spec should be printable
	printed, err := TypeRegistry.PrintType(spec)
	require.NoError(t, err)
	require.Equal(t, `branded_id("TSK")`, printed)
}
