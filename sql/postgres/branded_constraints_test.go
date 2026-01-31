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

func TestBrandedConstraintGeneratorForColumn(t *testing.T) {
	g := NewBrandedConstraintGenerator()

	tests := []struct {
		name       string
		tableName  string
		col        *schema.Column
		wantCheck  bool
		wantName   string
		wantExpr   string
	}{
		{
			name:      "TSK column",
			tableName: "tasks",
			col: &schema.Column{
				Name: "id",
				Type: &schema.ColumnType{
					Type: branded.BrandedIDFromNamespace(fiberfx.NS_TASK),
					Null: false,
				},
			},
			wantCheck: true,
			wantName:  "chk_tasks_id_branded",
			wantExpr:  "id ~ '^TSK[0-9A-Za-z]{11}$'",
		},
		{
			name:      "nullable EPC column",
			tableName: "tasks",
			col: &schema.Column{
				Name: "epic_id",
				Type: &schema.ColumnType{
					Type: branded.BrandedIDFromNamespace(fiberfx.NS_EPIC),
					Null: true,
				},
			},
			wantCheck: true,
			wantName:  "chk_tasks_epic_id_branded",
			wantExpr:  "epic_id IS NULL OR epic_id ~ '^EPC[0-9A-Za-z]{11}$'",
		},
		{
			name:      "non-branded column",
			tableName: "users",
			col: &schema.Column{
				Name: "name",
				Type: &schema.ColumnType{
					Type: &schema.StringType{T: "varchar", Size: 255},
				},
			},
			wantCheck: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check := g.GenerateForColumn(tt.tableName, tt.col)
			if !tt.wantCheck {
				require.Nil(t, check)
				return
			}
			require.NotNil(t, check)
			require.Equal(t, tt.wantName, check.Name)
			require.Equal(t, tt.wantExpr, check.Expr)
		})
	}
}

func TestBrandedConstraintGeneratorForTable(t *testing.T) {
	g := NewBrandedConstraintGenerator()

	table := &schema.Table{
		Name: "tasks",
		Columns: []*schema.Column{
			{
				Name: "id",
				Type: &schema.ColumnType{
					Type: branded.BrandedIDFromNamespace(fiberfx.NS_TASK),
				},
			},
			{
				Name: "epic_id",
				Type: &schema.ColumnType{
					Type: branded.BrandedIDFromNamespace(fiberfx.NS_EPIC),
					Null: true,
				},
			},
			{
				Name: "title",
				Type: &schema.ColumnType{
					Type: &schema.StringType{T: "varchar", Size: 255},
				},
			},
		},
	}

	checks := g.GenerateForTable(table)
	require.Len(t, checks, 2) // Only branded columns

	// Verify check names
	names := make(map[string]bool)
	for _, check := range checks {
		names[check.Name] = true
	}
	require.True(t, names["chk_tasks_id_branded"])
	require.True(t, names["chk_tasks_epic_id_branded"])
}

func TestBrandedConstraintGeneratorDisabled(t *testing.T) {
	g := NewBrandedConstraintGenerator(WithBrandedConstraints(false))

	col := &schema.Column{
		Name: "id",
		Type: &schema.ColumnType{
			Type: branded.BrandedIDFromNamespace(fiberfx.NS_TASK),
		},
	}

	check := g.GenerateForColumn("tasks", col)
	require.Nil(t, check)
}

func TestBrandedConstraintGeneratorCustomFormat(t *testing.T) {
	g := NewBrandedConstraintGenerator(WithConstraintFormat("branded_%s_%s_check"))

	col := &schema.Column{
		Name: "id",
		Type: &schema.ColumnType{
			Type: branded.BrandedIDFromNamespace(fiberfx.NS_TASK),
		},
	}

	check := g.GenerateForColumn("tasks", col)
	require.NotNil(t, check)
	require.Equal(t, "branded_tasks_id_check", check.Name)
}

func TestBrandedConstraintGeneratorGenerateSQL(t *testing.T) {
	g := NewBrandedConstraintGenerator()

	col := &schema.Column{
		Name: "id",
		Type: &schema.ColumnType{
			Type: branded.BrandedIDFromNamespace(fiberfx.NS_TASK),
		},
	}

	sql := g.GenerateSQL("tasks", col)
	require.Equal(t, `ALTER TABLE "tasks" ADD CONSTRAINT "chk_tasks_id_branded" CHECK (id ~ '^TSK[0-9A-Za-z]{11}$');`, sql)
}

func TestBrandedConstraintGeneratorGenerateAllSQL(t *testing.T) {
	g := NewBrandedConstraintGenerator()

	table := &schema.Table{
		Name: "tasks",
		Columns: []*schema.Column{
			{
				Name: "id",
				Type: &schema.ColumnType{
					Type: branded.BrandedIDFromNamespace(fiberfx.NS_TASK),
				},
			},
			{
				Name: "epic_id",
				Type: &schema.ColumnType{
					Type: branded.BrandedIDFromNamespace(fiberfx.NS_EPIC),
					Null: true,
				},
			},
		},
	}

	sqls := g.GenerateAllSQL(table)
	require.Len(t, sqls, 2)

	// Check SQL content
	require.Contains(t, sqls[0], "ADD CONSTRAINT")
	require.Contains(t, sqls[0], "CHECK")
}
