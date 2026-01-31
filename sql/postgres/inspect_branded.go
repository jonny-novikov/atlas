// Copyright 2021-present The Atlas Authors. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package postgres

import (
	"ariga.io/atlas/sql/branded"
	"ariga.io/atlas/sql/schema"
)

// ConvertBrandedIDColumns inspects column comments and converts matching
// VARCHAR(14) columns with "branded_id:NS" comments to BrandedIDType.
// This enables round-trip preservation of branded ID metadata through
// schema inspect/apply cycles.
func ConvertBrandedIDColumns(s *schema.Schema) {
	for _, t := range s.Tables {
		ConvertBrandedIDColumnsInTable(t)
	}
}

// ConvertBrandedIDColumnsInTable converts branded ID columns in a single table.
func ConvertBrandedIDColumnsInTable(t *schema.Table) {
	for _, c := range t.Columns {
		convertBrandedIDColumn(c)
	}
}

// convertBrandedIDColumn checks if a column is a branded ID based on its
// comment and type, and converts it to BrandedIDType if so.
func convertBrandedIDColumn(c *schema.Column) {
	if c.Type == nil {
		return
	}

	// Check if the column has a branded_id comment.
	comment := getComment(c)
	if comment == "" {
		return
	}

	ns, ok := branded.ParseComment(comment)
	if !ok {
		return
	}

	// Verify the underlying type is VARCHAR(14) or similar.
	if !isBrandedIDCompatible(c.Type.Type) {
		return
	}

	// Convert to BrandedIDType.
	c.Type.Type = branded.BrandedIDFromNamespace(ns)
}

// isBrandedIDCompatible checks if a schema.Type is compatible with branded IDs.
// Branded IDs are stored as VARCHAR(14), so we accept varchar/character varying
// with size 14.
func isBrandedIDCompatible(t schema.Type) bool {
	switch st := t.(type) {
	case *schema.StringType:
		// Accept varchar(14), character varying(14), char(14)
		return st.Size == 14
	default:
		return false
	}
}

// getComment extracts the comment text from a column's attributes.
func getComment(c *schema.Column) string {
	for _, a := range c.Attrs {
		if comment, ok := a.(*schema.Comment); ok {
			return comment.Text
		}
	}
	return ""
}

// IsBrandedIDColumn checks if a column is a branded ID column based on its type.
func IsBrandedIDColumn(c *schema.Column) bool {
	if c.Type == nil {
		return false
	}
	_, ok := c.Type.Type.(*branded.BrandedIDType)
	return ok
}

// GetBrandedIDNamespace returns the namespace if the column is a branded ID,
// otherwise returns empty string and false.
func GetBrandedIDNamespace(c *schema.Column) (string, bool) {
	if c.Type == nil {
		return "", false
	}
	bt, ok := c.Type.Type.(*branded.BrandedIDType)
	if !ok {
		return "", false
	}
	return string(bt.Namespace), true
}

// SetBrandedIDComment sets the branded_id:NS comment on a column.
// This should be called when generating COMMENT ON COLUMN statements.
func SetBrandedIDComment(c *schema.Column) {
	bt, ok := c.Type.Type.(*branded.BrandedIDType)
	if !ok {
		return
	}
	c.SetComment(branded.FormatComment(bt.Namespace))
}
