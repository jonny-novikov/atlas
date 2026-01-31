// Copyright 2021-present The Atlas Authors. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package branded

import (
	"testing"

	"github.com/jonny-novikov/jonnify/fiberfx"
	"github.com/stretchr/testify/require"
)

func TestProviderGenerate(t *testing.T) {
	p := NewProvider(1)

	// Generate IDs for different namespaces
	namespaces := []fiberfx.Namespace{
		fiberfx.NS_TASK,
		fiberfx.NS_EPIC,
		fiberfx.NS_FEATURE,
		fiberfx.NS_PLAN,
	}

	for _, ns := range namespaces {
		t.Run(string(ns), func(t *testing.T) {
			id := p.Generate(ns)
			require.False(t, id.IsZero())
			require.Equal(t, ns, id.Namespace())
			require.Len(t, id.String(), fiberfx.BrandedLen)
			require.True(t, p.IsValid(id.String()))
		})
	}
}

func TestProviderGenerateUnique(t *testing.T) {
	p := NewProvider(1)

	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id := p.Generate(fiberfx.NS_TASK)
		require.False(t, ids[id.String()], "ID should be unique")
		ids[id.String()] = true
	}
}

func TestProviderParse(t *testing.T) {
	p := NewProvider(1)

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "valid TSK",
			value:   "TSK0Ij1P13FRDM",
			wantErr: false,
		},
		{
			name:    "valid EPC",
			value:   "EPC0K5M2vuIULY",
			wantErr: false,
		},
		{
			name:    "too short",
			value:   "TSK123",
			wantErr: true,
		},
		{
			name:    "invalid format",
			value:   "INVALID_FORMAT",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := p.Parse(tt.value)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.value, id.String())
			}
		})
	}
}

func TestProviderParseWithNamespace(t *testing.T) {
	p := NewProvider(1)

	// Correct namespace
	id, err := p.ParseWithNamespace("TSK0Ij1P13FRDM", fiberfx.NS_TASK)
	require.NoError(t, err)
	require.Equal(t, fiberfx.NS_TASK, id.Namespace())

	// Wrong namespace
	_, err = p.ParseWithNamespace("TSK0Ij1P13FRDM", fiberfx.NS_EPIC)
	require.Error(t, err)
	require.Contains(t, err.Error(), "expected namespace")
}

func TestProviderIsValid(t *testing.T) {
	p := NewProvider(1)

	require.True(t, p.IsValid("TSK0Ij1P13FRDM"))
	require.True(t, p.IsValid("EPC0K5M2vuIULY"))
	require.False(t, p.IsValid("short"))
	require.False(t, p.IsValid("INVALID_FORMAT!"))
}

func TestDefaultProvider(t *testing.T) {
	p := DefaultProvider()
	require.NotNil(t, p)

	// Should be able to generate IDs
	id := p.Generate(fiberfx.NS_TASK)
	require.False(t, id.IsZero())
	require.Equal(t, fiberfx.NS_TASK, id.Namespace())
}

func TestProviderGenerator(t *testing.T) {
	p := NewProvider(42)
	gen := p.Generator()
	require.NotNil(t, gen)

	// Should be able to use generator methods
	id := gen.New(fiberfx.NS_TASK)
	require.False(t, id.IsZero())
	require.Equal(t, fiberfx.NS_TASK, id.Namespace())
}
