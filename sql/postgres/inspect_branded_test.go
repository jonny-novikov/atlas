// Copyright 2021-present The Atlas Authors. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package postgres

import (
	"testing"

	"ariga.io/atlas/sql/branded"
	"ariga.io/atlas/sql/schema"
	"github.com/jonny-novikov/jonnify/fiberfx"
	"github.com/stretchr/testify/require"
)

func TestConvertBrandedIDColumn(t *testing.T) {
	tests := []struct {
		name        string
		column      *schema.Column
		wantBranded bool
		wantNS      string
	}{
		{
			name: "varchar(14) with branded_id comment",
			column: schema.NewColumn("id").
				SetType(&schema.StringType{T: "character varying", Size: 14}).
				SetComment("branded_id:TSK"),
			wantBranded: true,
			wantNS:      "TSK",
		},
		{
			name: "varchar(14) without comment",
			column: schema.NewColumn("id").
				SetType(&schema.StringType{T: "character varying", Size: 14}),
			wantBranded: false,
		},
		{
			name: "varchar(14) with regular comment",
			column: schema.NewColumn("id").
				SetType(&schema.StringType{T: "character varying", Size: 14}).
				SetComment("Just a regular ID column"),
			wantBranded: false,
		},
		{
			name: "varchar(255) with branded_id comment - wrong size",
			column: schema.NewColumn("name").
				SetType(&schema.StringType{T: "character varying", Size: 255}).
				SetComment("branded_id:TSK"),
			wantBranded: false,
		},
		{
			name: "integer with branded_id comment - wrong type",
			column: schema.NewColumn("count").
				SetType(&schema.IntegerType{T: "integer"}).
				SetComment("branded_id:TSK"),
			wantBranded: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			convertBrandedIDColumn(tt.column)

			isBranded := IsBrandedIDColumn(tt.column)
			require.Equal(t, tt.wantBranded, isBranded)

			if tt.wantBranded {
				ns, ok := GetBrandedIDNamespace(tt.column)
				require.True(t, ok)
				require.Equal(t, tt.wantNS, ns)
			}
		})
	}
}

func TestConvertBrandedIDColumnsInTable(t *testing.T) {
	table := schema.NewTable("tasks").
		AddColumns(
			schema.NewColumn("id").
				SetType(&schema.StringType{T: "character varying", Size: 14}).
				SetComment("branded_id:TSK"),
			schema.NewColumn("epic_id").
				SetType(&schema.StringType{T: "character varying", Size: 14}).
				SetComment("branded_id:EPC"),
			schema.NewColumn("title").
				SetType(&schema.StringType{T: "text"}),
		)

	ConvertBrandedIDColumnsInTable(table)

	// Check id column
	idCol, ok := table.Column("id")
	require.True(t, ok)
	require.True(t, IsBrandedIDColumn(idCol))
	ns, _ := GetBrandedIDNamespace(idCol)
	require.Equal(t, "TSK", ns)

	// Check epic_id column
	epicCol, ok := table.Column("epic_id")
	require.True(t, ok)
	require.True(t, IsBrandedIDColumn(epicCol))
	ns, _ = GetBrandedIDNamespace(epicCol)
	require.Equal(t, "EPC", ns)

	// Check title column (should not be branded)
	titleCol, ok := table.Column("title")
	require.True(t, ok)
	require.False(t, IsBrandedIDColumn(titleCol))
}

func TestConvertBrandedIDColumns(t *testing.T) {
	s := schema.New("public").
		AddTables(
			schema.NewTable("tasks").
				AddColumns(
					schema.NewColumn("id").
						SetType(&schema.StringType{T: "character varying", Size: 14}).
						SetComment("branded_id:TSK"),
				),
			schema.NewTable("epics").
				AddColumns(
					schema.NewColumn("id").
						SetType(&schema.StringType{T: "character varying", Size: 14}).
						SetComment("branded_id:EPC"),
				),
		)

	ConvertBrandedIDColumns(s)

	// Check tasks.id
	tasksTable, ok := s.Table("tasks")
	require.True(t, ok)
	idCol, ok := tasksTable.Column("id")
	require.True(t, ok)
	require.True(t, IsBrandedIDColumn(idCol))

	// Check epics.id
	epicsTable, ok := s.Table("epics")
	require.True(t, ok)
	epicIdCol, ok := epicsTable.Column("id")
	require.True(t, ok)
	require.True(t, IsBrandedIDColumn(epicIdCol))
}

func TestSetBrandedIDComment(t *testing.T) {
	col := schema.NewColumn("id").
		SetType(branded.BrandedID("TSK"))

	SetBrandedIDComment(col)

	// Check the comment was set
	var comment string
	for _, a := range col.Attrs {
		if c, ok := a.(*schema.Comment); ok {
			comment = c.Text
			break
		}
	}
	require.Equal(t, "branded_id:TSK", comment)
}

func TestIsBrandedIDCompatible(t *testing.T) {
	tests := []struct {
		name string
		typ  schema.Type
		want bool
	}{
		{
			name: "varchar(14)",
			typ:  &schema.StringType{T: "character varying", Size: 14},
			want: true,
		},
		{
			name: "varchar(255)",
			typ:  &schema.StringType{T: "character varying", Size: 255},
			want: false,
		},
		{
			name: "text",
			typ:  &schema.StringType{T: "text"},
			want: false,
		},
		{
			name: "integer",
			typ:  &schema.IntegerType{T: "integer"},
			want: false,
		},
		{
			name: "uuid",
			typ:  &schema.UUIDType{T: "uuid"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isBrandedIDCompatible(tt.typ)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestGetBrandedIDNamespace_NotBranded(t *testing.T) {
	col := schema.NewColumn("id").
		SetType(&schema.StringType{T: "text"})

	ns, ok := GetBrandedIDNamespace(col)
	require.False(t, ok)
	require.Empty(t, ns)
}

func TestSetBrandedIDComment_NotBranded(t *testing.T) {
	col := schema.NewColumn("id").
		SetType(&schema.StringType{T: "text"})

	// Should not panic, just do nothing
	SetBrandedIDComment(col)

	// Verify no comment was set
	var hasComment bool
	for _, a := range col.Attrs {
		if _, ok := a.(*schema.Comment); ok {
			hasComment = true
			break
		}
	}
	require.False(t, hasComment)
}

// TestFormatTypeWithBrandedID tests that FormatType correctly handles BrandedIDType.
func TestFormatTypeWithBrandedID(t *testing.T) {
	bid := branded.BrandedID("TSK")
	sql, err := FormatType(bid)
	require.NoError(t, err)
	require.Equal(t, "character varying(14)", sql)
}

// TestAllNamespacesValid verifies we can create branded IDs for all known namespaces.
func TestAllNamespacesValid(t *testing.T) {
	namespaces := []fiberfx.Namespace{
		fiberfx.NS_TASK,
		fiberfx.NS_EPIC,
		fiberfx.NS_FEATURE,
		fiberfx.NS_PLAN,
		fiberfx.NS_KB,
	}

	for _, ns := range namespaces {
		t.Run(string(ns), func(t *testing.T) {
			bid := branded.BrandedIDFromNamespace(ns)
			require.Equal(t, ns, bid.Namespace)

			sql, err := FormatType(bid)
			require.NoError(t, err)
			require.Equal(t, "character varying(14)", sql)
		})
	}
}
