// Copyright 2021-present The Atlas Authors. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package branded

import (
	"testing"

	"ariga.io/atlas/sql/branded"
	"ariga.io/atlas/sql/schema"
	"github.com/jonny-novikov/jonnify/fiberfx"
	"github.com/stretchr/testify/require"
)

func TestValidatorValidateColumn(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name       string
		tableName  string
		col        *schema.Column
		wantErrors int
	}{
		{
			name:      "valid TSK column",
			tableName: "tasks",
			col: &schema.Column{
				Name: "id",
				Type: &schema.ColumnType{
					Type: branded.BrandedIDFromNamespace(fiberfx.NS_TASK),
				},
			},
			wantErrors: 0,
		},
		{
			name:      "valid EPC column",
			tableName: "epics",
			col: &schema.Column{
				Name: "id",
				Type: &schema.ColumnType{
					Type: branded.BrandedIDFromNamespace(fiberfx.NS_EPIC),
				},
			},
			wantErrors: 0,
		},
		{
			name:      "invalid namespace",
			tableName: "test",
			col: &schema.Column{
				Name: "id",
				Type: &schema.ColumnType{
					Type: branded.BrandedID("XXX"), // Invalid namespace
				},
			},
			wantErrors: 1,
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
			wantErrors: 0, // Should be ignored
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := v.ValidateColumn(tt.tableName, tt.col)
			require.Len(t, errs, tt.wantErrors)
		})
	}
}

func TestValidatorWithNamingConvention(t *testing.T) {
	v := NewValidator(WithNamingConvention(true))

	tests := []struct {
		name       string
		tableName  string
		colName    string
		ns         fiberfx.Namespace
		wantErrors int
	}{
		{
			name:       "id is valid",
			tableName:  "tasks",
			colName:    "id",
			ns:         fiberfx.NS_TASK,
			wantErrors: 0,
		},
		{
			name:       "namespace_id is valid",
			tableName:  "tasks",
			colName:    "tsk_id",
			ns:         fiberfx.NS_TASK,
			wantErrors: 0,
		},
		{
			name:       "entity_id is valid",
			tableName:  "tasks",
			colName:    "epic_id",
			ns:         fiberfx.NS_EPIC,
			wantErrors: 0,
		},
		{
			name:       "non-standard name warns",
			tableName:  "tasks",
			colName:    "identifier",
			ns:         fiberfx.NS_TASK,
			wantErrors: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			col := &schema.Column{
				Name: tt.colName,
				Type: &schema.ColumnType{
					Type: branded.BrandedIDFromNamespace(tt.ns),
				},
			}
			errs := v.ValidateColumn(tt.tableName, col)
			require.Len(t, errs, tt.wantErrors)
		})
	}
}

func TestValidatorValidateValue(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name     string
		value    string
		expected fiberfx.Namespace
		wantErr  bool
	}{
		{
			name:     "valid TSK ID",
			value:    "TSK0Ij1P13FRDM",
			expected: fiberfx.NS_TASK,
			wantErr:  false,
		},
		{
			name:     "valid EPC ID",
			value:    "EPC0K5M2vuIULY",
			expected: fiberfx.NS_EPIC,
			wantErr:  false,
		},
		{
			name:     "wrong namespace",
			value:    "TSK0Ij1P13FRDM",
			expected: fiberfx.NS_EPIC,
			wantErr:  true,
		},
		{
			name:     "too short",
			value:    "TSK123",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "too long",
			value:    "TSK0Ij1P13FRDMX",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "invalid characters",
			value:    "TSK0Ij1P13FRD!",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "any namespace ok when not specified",
			value:    "TSK0Ij1P13FRDM",
			expected: "",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateValue(tt.value, tt.expected)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidatorValidateForeignKey(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name       string
		fk         *schema.ForeignKey
		wantErrors int
	}{
		{
			name: "matching namespaces",
			fk: &schema.ForeignKey{
				Symbol: "fk_task_epic",
				Columns: []*schema.Column{
					{
						Name: "epic_id",
						Type: &schema.ColumnType{
							Type: branded.BrandedIDFromNamespace(fiberfx.NS_EPIC),
						},
					},
				},
				RefColumns: []*schema.Column{
					{
						Name: "id",
						Type: &schema.ColumnType{
							Type: branded.BrandedIDFromNamespace(fiberfx.NS_EPIC),
						},
					},
				},
			},
			wantErrors: 0,
		},
		{
			name: "mismatched namespaces",
			fk: &schema.ForeignKey{
				Symbol: "fk_bad",
				Columns: []*schema.Column{
					{
						Name: "task_id",
						Type: &schema.ColumnType{
							Type: branded.BrandedIDFromNamespace(fiberfx.NS_TASK),
						},
					},
				},
				RefColumns: []*schema.Column{
					{
						Name: "id",
						Type: &schema.ColumnType{
							Type: branded.BrandedIDFromNamespace(fiberfx.NS_EPIC),
						},
					},
				},
			},
			wantErrors: 1,
		},
		{
			name: "branded to non-branded",
			fk: &schema.ForeignKey{
				Symbol: "fk_mixed",
				Columns: []*schema.Column{
					{
						Name: "epic_id",
						Type: &schema.ColumnType{
							Type: branded.BrandedIDFromNamespace(fiberfx.NS_EPIC),
						},
					},
				},
				RefColumns: []*schema.Column{
					{
						Name: "id",
						Type: &schema.ColumnType{
							Type: &schema.IntegerType{T: "bigint"},
						},
					},
				},
			},
			wantErrors: 1,
		},
		{
			name: "non-branded FK ignored",
			fk: &schema.ForeignKey{
				Symbol: "fk_normal",
				Columns: []*schema.Column{
					{
						Name: "user_id",
						Type: &schema.ColumnType{
							Type: &schema.IntegerType{T: "bigint"},
						},
					},
				},
				RefColumns: []*schema.Column{
					{
						Name: "id",
						Type: &schema.ColumnType{
							Type: &schema.IntegerType{T: "bigint"},
						},
					},
				},
			},
			wantErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := v.ValidateForeignKey(tt.fk)
			require.Len(t, errs, tt.wantErrors)
		})
	}
}

func TestValidatorValidateSchema(t *testing.T) {
	v := NewValidator()

	s := &schema.Schema{
		Name: "public",
		Tables: []*schema.Table{
			{
				Name: "epics",
				Columns: []*schema.Column{
					{
						Name: "id",
						Type: &schema.ColumnType{
							Type: branded.BrandedIDFromNamespace(fiberfx.NS_EPIC),
						},
					},
				},
			},
			{
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
			},
		},
	}

	errs := v.ValidateSchema(s)
	require.Empty(t, errs)
}

func TestValidatorValidateSchemaWithErrors(t *testing.T) {
	v := NewValidator()

	s := &schema.Schema{
		Name: "public",
		Tables: []*schema.Table{
			{
				Name: "bad_table",
				Columns: []*schema.Column{
					{
						Name: "id",
						Type: &schema.ColumnType{
							Type: branded.BrandedID("XXX"), // Invalid namespace
						},
					},
				},
			},
		},
	}

	errs := v.ValidateSchema(s)
	require.Len(t, errs, 1)
	require.Contains(t, errs[0].Message, "unknown namespace")
}

func TestValidationErrorFormat(t *testing.T) {
	// Column error
	err1 := ValidationError{
		Table:   "tasks",
		Column:  "id",
		Message: "invalid namespace",
	}
	require.Equal(t, `column "tasks"."id": invalid namespace`, err1.Error())

	// FK error
	err2 := ValidationError{
		FK:      "fk_epic",
		Message: "namespace mismatch",
	}
	require.Equal(t, `FK "fk_epic": namespace mismatch`, err2.Error())

	// Generic error
	err3 := ValidationError{
		Message: "general error",
	}
	require.Equal(t, "general error", err3.Error())
}
