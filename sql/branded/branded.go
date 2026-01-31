// Copyright 2021-present The Atlas Authors. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

// Package branded provides the BrandedIDType for Atlas schema management.
// Branded IDs are 14-character identifiers with a 3-character namespace prefix
// and 11-character Base62-encoded snowflake (e.g., "TSK0Ij1P13FRDM").
package branded

import (
	"fmt"
	"regexp"
	"strings"

	"ariga.io/atlas/sql/schema"
	"github.com/jonny-novikov/jonnify/fiberfx"
)

// BrandedIDType represents a branded ID column type.
// Stored as VARCHAR(14) but carries namespace metadata.
type BrandedIDType struct {
	schema.Type
	// Namespace is the 3-character prefix (e.g., "TSK", "EPC", "FTR").
	Namespace fiberfx.Namespace
}

// typ implements schema.Type interface marker.
func (*BrandedIDType) typ() {}

// String returns the SQL type representation.
func (t *BrandedIDType) String() string {
	return "varchar(14)"
}

// Underlying returns the physical SQL type.
func (t *BrandedIDType) Underlying() schema.Type {
	return &schema.StringType{
		T:    "character varying",
		Size: 14,
	}
}

// Is reports if t is the same type as x.
func (t *BrandedIDType) Is(x schema.Type) bool {
	b, ok := x.(*BrandedIDType)
	return ok && t.Namespace == b.Namespace
}

// BrandedID creates a new branded ID type with the given namespace.
func BrandedID(ns string) *BrandedIDType {
	return &BrandedIDType{
		Namespace: fiberfx.Namespace(strings.ToUpper(ns)),
	}
}

// BrandedIDFromNamespace creates a new branded ID type from a fiberfx.Namespace.
func BrandedIDFromNamespace(ns fiberfx.Namespace) *BrandedIDType {
	return &BrandedIDType{
		Namespace: ns,
	}
}

// CommentMarker is the column comment prefix used to identify branded ID columns.
const CommentMarker = "branded_id:"

// reComment matches the branded_id:NS pattern in column comments.
var reComment = regexp.MustCompile(`branded_id:([A-Z]{3})`)

// ParseComment extracts the namespace from a column comment.
// Returns the namespace and true if found, empty string and false otherwise.
func ParseComment(comment string) (fiberfx.Namespace, bool) {
	matches := reComment.FindStringSubmatch(comment)
	if len(matches) != 2 {
		return "", false
	}
	return fiberfx.Namespace(matches[1]), true
}

// FormatComment returns the column comment for a branded ID with the given namespace.
func FormatComment(ns fiberfx.Namespace) string {
	return fmt.Sprintf("%s%s", CommentMarker, ns)
}

// IsValidNamespace checks if the given namespace is valid.
func IsValidNamespace(ns string) bool {
	return fiberfx.IsValidNamespace(fiberfx.Namespace(ns))
}

// Validate validates a branded ID string.
// Returns the parsed ID and nil error if valid.
func Validate(id string) (fiberfx.ID, error) {
	return fiberfx.Parse(id)
}

// ValidateWithNamespace validates a branded ID string against an expected namespace.
func ValidateWithNamespace(id string, ns fiberfx.Namespace) (fiberfx.ID, error) {
	return fiberfx.ParseWithNamespace(id, ns)
}

// CheckConstraintExpr returns a CHECK constraint expression for validating
// the branded ID format in SQL. The column name is parameterized.
func CheckConstraintExpr(column string, ns fiberfx.Namespace) string {
	return fmt.Sprintf("%s ~ '^%s[0-9A-Za-z]{11}$'", column, ns)
}

// CheckConstraintName returns the conventional name for a branded ID check constraint.
func CheckConstraintName(table, column string) string {
	return fmt.Sprintf("chk_%s_%s_format", table, column)
}
