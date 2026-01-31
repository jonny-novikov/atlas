// Copyright 2021-present The Atlas Authors. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

// Package branded provides validation for branded ID columns in Atlas schemas.
package branded

import (
	"fmt"
	"strings"

	"ariga.io/atlas/sql/branded"
	"ariga.io/atlas/sql/schema"
	"github.com/jonny-novikov/jonnify/fiberfx"
)

// ValidationError represents a branded ID validation error.
type ValidationError struct {
	Table   string
	Column  string
	FK      string
	Message string
}

func (e ValidationError) Error() string {
	if e.FK != "" {
		return fmt.Sprintf("FK %q: %s", e.FK, e.Message)
	}
	if e.Column != "" {
		return fmt.Sprintf("column %q.%q: %s", e.Table, e.Column, e.Message)
	}
	return e.Message
}

// ValidatorOption configures a Validator.
type ValidatorOption func(*Validator)

// WithStrict enables strict mode that rejects unknown namespaces.
func WithStrict(strict bool) ValidatorOption {
	return func(v *Validator) {
		v.strict = strict
	}
}

// WithNamingConvention enables naming convention warnings.
func WithNamingConvention(enabled bool) ValidatorOption {
	return func(v *Validator) {
		v.checkNaming = enabled
	}
}

// Validator validates branded ID columns and values.
type Validator struct {
	strict      bool // If true, reject unknown namespaces
	checkNaming bool // If true, warn on naming convention violations
}

// NewValidator creates a new branded ID validator.
func NewValidator(opts ...ValidatorOption) *Validator {
	v := &Validator{
		strict:      true,
		checkNaming: false,
	}
	for _, opt := range opts {
		opt(v)
	}
	return v
}

// ValidateSchema validates all branded ID columns in a schema.
func (v *Validator) ValidateSchema(s *schema.Schema) []ValidationError {
	var errs []ValidationError
	for _, t := range s.Tables {
		errs = append(errs, v.ValidateTable(t)...)
	}
	return errs
}

// ValidateTable validates all branded ID columns in a table.
func (v *Validator) ValidateTable(t *schema.Table) []ValidationError {
	var errs []ValidationError

	// Validate columns
	for _, col := range t.Columns {
		colErrs := v.ValidateColumn(t.Name, col)
		errs = append(errs, colErrs...)
	}

	// Validate foreign keys
	for _, fk := range t.ForeignKeys {
		fkErrs := v.ValidateForeignKey(fk)
		errs = append(errs, fkErrs...)
	}

	return errs
}

// ValidateColumn validates a branded ID column definition.
func (v *Validator) ValidateColumn(tableName string, col *schema.Column) []ValidationError {
	var errs []ValidationError

	bt, ok := col.Type.Type.(*branded.BrandedIDType)
	if !ok {
		return nil // Not a branded ID column
	}

	// 1. Validate namespace is known
	if v.strict && !fiberfx.IsValidNamespace(fiberfx.Namespace(bt.Namespace)) {
		errs = append(errs, ValidationError{
			Table:   tableName,
			Column:  col.Name,
			Message: fmt.Sprintf("unknown namespace %q; valid: %v", bt.Namespace, fiberfx.AllNamespaces()),
		})
	}

	// 2. Validate column naming convention (optional)
	if v.checkNaming && !v.isValidColumnName(col.Name, bt.Namespace) {
		errs = append(errs, ValidationError{
			Table:   tableName,
			Column:  col.Name,
			Message: fmt.Sprintf("recommended naming is 'id' or '%s_id'", strings.ToLower(string(bt.Namespace))),
		})
	}

	return errs
}

// isValidColumnName checks if column name follows convention.
// Primary keys should be "id", foreign keys should be "{namespace}_id" or "{entity}_id".
func (v *Validator) isValidColumnName(colName string, ns fiberfx.Namespace) bool {
	if colName == "id" {
		return true
	}
	// Accept {ns}_id pattern (e.g., tsk_id, epic_id)
	nsLower := strings.ToLower(string(ns))
	if colName == nsLower+"_id" {
		return true
	}
	// Accept {entity}_id pattern (e.g., task_id, epic_id, feature_id)
	if strings.HasSuffix(colName, "_id") {
		return true
	}
	return false
}

// ValidateValue validates a branded ID value.
func (v *Validator) ValidateValue(value string, expected fiberfx.Namespace) error {
	if len(value) != fiberfx.BrandedLen {
		return fmt.Errorf("branded ID must be %d characters, got %d", fiberfx.BrandedLen, len(value))
	}

	if !fiberfx.Valid(value) {
		return fmt.Errorf("branded ID %q has invalid format", value)
	}

	// Validate namespace if expected
	if expected != "" {
		ns := fiberfx.Namespace(value[:fiberfx.NamespaceLen])
		if ns != expected {
			return fmt.Errorf("expected namespace %q, got %q", expected, ns)
		}
	}

	return nil
}

// ValidateForeignKey validates namespace consistency in foreign keys.
func (v *Validator) ValidateForeignKey(fk *schema.ForeignKey) []ValidationError {
	var errs []ValidationError

	for i, col := range fk.Columns {
		if i >= len(fk.RefColumns) {
			continue
		}
		refCol := fk.RefColumns[i]

		bt, ok := col.Type.Type.(*branded.BrandedIDType)
		if !ok {
			continue // Not a branded ID column
		}

		refBt, ok := refCol.Type.Type.(*branded.BrandedIDType)
		if !ok {
			errs = append(errs, ValidationError{
				FK:      fk.Symbol,
				Message: fmt.Sprintf("column %q is branded_id but reference %q is not", col.Name, refCol.Name),
			})
			continue
		}

		if bt.Namespace != refBt.Namespace {
			errs = append(errs, ValidationError{
				FK:      fk.Symbol,
				Message: fmt.Sprintf("namespace mismatch %q (%s) -> %q (%s)", col.Name, bt.Namespace, refCol.Name, refBt.Namespace),
			})
		}
	}

	return errs
}
