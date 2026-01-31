// Copyright 2021-present The Atlas Authors. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package postgres

import (
	"fmt"

	"ariga.io/atlas/sql/branded"
	"ariga.io/atlas/sql/schema"
)

// BrandedConstraintOption configures branded ID constraint generation.
type BrandedConstraintOption func(*brandedConstraintConfig)

type brandedConstraintConfig struct {
	enabled       bool
	constraintFmt string
}

// WithBrandedConstraints enables or disables CHECK constraint generation.
func WithBrandedConstraints(enabled bool) BrandedConstraintOption {
	return func(c *brandedConstraintConfig) {
		c.enabled = enabled
	}
}

// WithConstraintFormat sets the constraint name format.
// Default: "chk_%s_%s_branded" (table, column)
func WithConstraintFormat(format string) BrandedConstraintOption {
	return func(c *brandedConstraintConfig) {
		c.constraintFmt = format
	}
}

// BrandedConstraintGenerator generates CHECK constraints for branded ID columns.
type BrandedConstraintGenerator struct {
	config brandedConstraintConfig
}

// NewBrandedConstraintGenerator creates a new generator with options.
func NewBrandedConstraintGenerator(opts ...BrandedConstraintOption) *BrandedConstraintGenerator {
	g := &BrandedConstraintGenerator{
		config: brandedConstraintConfig{
			enabled:       true,
			constraintFmt: "chk_%s_%s_branded",
		},
	}
	for _, opt := range opts {
		opt(&g.config)
	}
	return g
}

// GenerateForTable generates CHECK constraints for all branded ID columns in a table.
func (g *BrandedConstraintGenerator) GenerateForTable(t *schema.Table) []*schema.Check {
	if !g.config.enabled {
		return nil
	}

	var checks []*schema.Check
	for _, col := range t.Columns {
		if check := g.GenerateForColumn(t.Name, col); check != nil {
			checks = append(checks, check)
		}
	}
	return checks
}

// GenerateForColumn generates a CHECK constraint for a branded ID column.
// Returns nil if the column is not a branded ID type.
func (g *BrandedConstraintGenerator) GenerateForColumn(tableName string, col *schema.Column) *schema.Check {
	if !g.config.enabled {
		return nil
	}

	bt, ok := col.Type.Type.(*branded.BrandedIDType)
	if !ok {
		return nil
	}

	constraintName := fmt.Sprintf(g.config.constraintFmt, tableName, col.Name)
	ns := string(bt.Namespace)

	// Generate PostgreSQL regex CHECK constraint
	// Format: ^{NS}[0-9A-Za-z]{11}$
	var expr string
	if col.Type.Null {
		// Allow NULL values
		expr = fmt.Sprintf("%s IS NULL OR %s ~ '^%s[0-9A-Za-z]{11}$'", col.Name, col.Name, ns)
	} else {
		expr = fmt.Sprintf("%s ~ '^%s[0-9A-Za-z]{11}$'", col.Name, ns)
	}

	return &schema.Check{
		Name: constraintName,
		Expr: expr,
	}
}

// GenerateSQL generates the ALTER TABLE statement for adding a CHECK constraint.
func (g *BrandedConstraintGenerator) GenerateSQL(tableName string, col *schema.Column) string {
	check := g.GenerateForColumn(tableName, col)
	if check == nil {
		return ""
	}
	return fmt.Sprintf(
		`ALTER TABLE %q ADD CONSTRAINT %q CHECK (%s);`,
		tableName,
		check.Name,
		check.Expr,
	)
}

// GenerateAllSQL generates all CHECK constraint statements for a table.
func (g *BrandedConstraintGenerator) GenerateAllSQL(t *schema.Table) []string {
	checks := g.GenerateForTable(t)
	var sqls []string
	for _, check := range checks {
		sqls = append(sqls, fmt.Sprintf(
			`ALTER TABLE %q ADD CONSTRAINT %q CHECK (%s);`,
			t.Name,
			check.Name,
			check.Expr,
		))
	}
	return sqls
}
