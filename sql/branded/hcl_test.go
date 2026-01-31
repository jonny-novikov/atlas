// Copyright 2021-present The Atlas Authors. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package branded

import (
	"testing"

	"ariga.io/atlas/schemahcl"
	"ariga.io/atlas/sql/schema"
	"github.com/jonny-novikov/jonnify/fiberfx"
	"github.com/stretchr/testify/require"
)

func TestBrandedIDTypeSpec(t *testing.T) {
	spec := BrandedIDTypeSpec()
	require.NotNil(t, spec)
	require.Equal(t, TypeBrandedID, spec.Name)
	require.Equal(t, TypeBrandedID, spec.T)
	require.Len(t, spec.Attributes, 1)
	require.Equal(t, "namespace", spec.Attributes[0].Name)
	require.True(t, spec.Attributes[0].Required)
}

func TestFromSpec(t *testing.T) {
	tests := []struct {
		name    string
		typ     *schemahcl.Type
		wantNS  fiberfx.Namespace
		wantErr bool
	}{
		{
			name: "valid TSK namespace",
			typ: &schemahcl.Type{
				T:     TypeBrandedID,
				Attrs: []*schemahcl.Attr{schemahcl.StringAttr("namespace", "TSK")},
			},
			wantNS:  fiberfx.NS_TASK,
			wantErr: false,
		},
		{
			name: "valid EPC namespace",
			typ: &schemahcl.Type{
				T:     TypeBrandedID,
				Attrs: []*schemahcl.Attr{schemahcl.StringAttr("namespace", "EPC")},
			},
			wantNS:  fiberfx.NS_EPIC,
			wantErr: false,
		},
		{
			name: "valid FTR namespace",
			typ: &schemahcl.Type{
				T:     TypeBrandedID,
				Attrs: []*schemahcl.Attr{schemahcl.StringAttr("namespace", "FTR")},
			},
			wantNS:  fiberfx.NS_FEATURE,
			wantErr: false,
		},
		{
			name: "wrong type name",
			typ: &schemahcl.Type{
				T:     "varchar",
				Attrs: []*schemahcl.Attr{schemahcl.StringAttr("namespace", "TSK")},
			},
			wantErr: true,
		},
		{
			name: "missing namespace",
			typ: &schemahcl.Type{
				T: TypeBrandedID,
			},
			wantErr: true,
		},
		{
			name: "invalid namespace",
			typ: &schemahcl.Type{
				T:     TypeBrandedID,
				Attrs: []*schemahcl.Attr{schemahcl.StringAttr("namespace", "XXX")},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := fromSpec(tt.typ)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			bid, ok := result.(*BrandedIDType)
			require.True(t, ok)
			require.Equal(t, tt.wantNS, bid.Namespace)
		})
	}
}

func TestToSpec(t *testing.T) {
	tests := []struct {
		name     string
		typ      schema.Type
		wantNS   string
		wantType string
		wantErr  bool
	}{
		{
			name:     "TSK branded ID",
			typ:      BrandedID("TSK"),
			wantNS:   "TSK",
			wantType: TypeBrandedID,
			wantErr:  false,
		},
		{
			name:     "EPC branded ID",
			typ:      BrandedIDFromNamespace(fiberfx.NS_EPIC),
			wantNS:   "EPC",
			wantType: TypeBrandedID,
			wantErr:  false,
		},
		{
			name:    "wrong type",
			typ:     &schema.StringType{T: "varchar", Size: 14},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := toSpec(tt.typ)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			require.Equal(t, tt.wantType, result.T)
			require.Len(t, result.Attrs, 1)
			require.Equal(t, "namespace", result.Attrs[0].K)

			ns, err := result.Attrs[0].String()
			require.NoError(t, err)
			require.Equal(t, tt.wantNS, ns)
		})
	}
}

func TestRoundTrip(t *testing.T) {
	// Test that we can convert from schema type to HCL type and back
	namespaces := []fiberfx.Namespace{
		fiberfx.NS_TASK,
		fiberfx.NS_EPIC,
		fiberfx.NS_FEATURE,
		fiberfx.NS_PLAN,
		fiberfx.NS_KB,
	}

	for _, ns := range namespaces {
		t.Run(string(ns), func(t *testing.T) {
			// Create schema type
			original := BrandedIDFromNamespace(ns)

			// Convert to HCL spec
			spec, err := toSpec(original)
			require.NoError(t, err)
			require.Equal(t, TypeBrandedID, spec.T)

			// Convert back to schema type
			result, err := fromSpec(spec)
			require.NoError(t, err)

			// Verify they match
			bid, ok := result.(*BrandedIDType)
			require.True(t, ok)
			require.Equal(t, original.Namespace, bid.Namespace)
		})
	}
}

func TestIsBrandedIDType(t *testing.T) {
	tests := []struct {
		name string
		typ  *schemahcl.Type
		want bool
	}{
		{
			name: "branded_id type",
			typ:  &schemahcl.Type{T: TypeBrandedID},
			want: true,
		},
		{
			name: "varchar type",
			typ:  &schemahcl.Type{T: "varchar"},
			want: false,
		},
		{
			name: "nil type",
			typ:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsBrandedIDType(tt.typ)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestNamespaceFromType(t *testing.T) {
	tests := []struct {
		name string
		typ  *schemahcl.Type
		want string
	}{
		{
			name: "with namespace",
			typ: &schemahcl.Type{
				T:     TypeBrandedID,
				Attrs: []*schemahcl.Attr{schemahcl.StringAttr("namespace", "TSK")},
			},
			want: "TSK",
		},
		{
			name: "non-branded type",
			typ:  &schemahcl.Type{T: "varchar"},
			want: "",
		},
		{
			name: "branded without namespace attr",
			typ:  &schemahcl.Type{T: TypeBrandedID},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NamespaceFromType(tt.typ)
			require.Equal(t, tt.want, got)
		})
	}
}
