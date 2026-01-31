// Copyright 2021-present The Atlas Authors. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package branded

import (
	"testing"

	"ariga.io/atlas/sql/schema"
	"github.com/jonny-novikov/jonnify/fiberfx"
	"github.com/stretchr/testify/require"
)

func TestBrandedID(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		wantNS    fiberfx.Namespace
	}{
		{
			name:      "task namespace",
			namespace: "TSK",
			wantNS:    fiberfx.NS_TASK,
		},
		{
			name:      "epic namespace",
			namespace: "EPC",
			wantNS:    fiberfx.NS_EPIC,
		},
		{
			name:      "lowercase normalized",
			namespace: "tsk",
			wantNS:    fiberfx.NS_TASK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bid := BrandedID(tt.namespace)
			require.NotNil(t, bid)
			require.Equal(t, tt.wantNS, bid.Namespace)
		})
	}
}

func TestBrandedIDType_String(t *testing.T) {
	bid := BrandedID("TSK")
	require.Equal(t, "varchar(14)", bid.String())
}

func TestBrandedIDType_Underlying(t *testing.T) {
	bid := BrandedID("TSK")
	underlying := bid.Underlying()

	st, ok := underlying.(*schema.StringType)
	require.True(t, ok)
	require.Equal(t, "character varying", st.T)
	require.Equal(t, 14, st.Size)
}

func TestBrandedIDType_Is(t *testing.T) {
	bid1 := BrandedID("TSK")
	bid2 := BrandedID("TSK")
	bid3 := BrandedID("EPC")

	require.True(t, bid1.Is(bid2))
	require.False(t, bid1.Is(bid3))
	require.False(t, bid1.Is(&schema.StringType{}))
}

func TestParseComment(t *testing.T) {
	tests := []struct {
		name    string
		comment string
		wantNS  fiberfx.Namespace
		wantOK  bool
	}{
		{
			name:    "valid branded_id comment",
			comment: "branded_id:TSK",
			wantNS:  "TSK",
			wantOK:  true,
		},
		{
			name:    "branded_id with other text",
			comment: "User task identifier branded_id:TSK for tracking",
			wantNS:  "TSK",
			wantOK:  true,
		},
		{
			name:    "epic namespace",
			comment: "branded_id:EPC",
			wantNS:  "EPC",
			wantOK:  true,
		},
		{
			name:    "no branded_id marker",
			comment: "Regular column comment",
			wantNS:  "",
			wantOK:  false,
		},
		{
			name:    "empty comment",
			comment: "",
			wantNS:  "",
			wantOK:  false,
		},
		{
			name:    "invalid namespace format",
			comment: "branded_id:T",
			wantNS:  "",
			wantOK:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns, ok := ParseComment(tt.comment)
			require.Equal(t, tt.wantOK, ok)
			if ok {
				require.Equal(t, tt.wantNS, ns)
			}
		})
	}
}

func TestFormatComment(t *testing.T) {
	tests := []struct {
		ns   fiberfx.Namespace
		want string
	}{
		{fiberfx.NS_TASK, "branded_id:TSK"},
		{fiberfx.NS_EPIC, "branded_id:EPC"},
		{fiberfx.NS_FEATURE, "branded_id:FTR"},
	}

	for _, tt := range tests {
		t.Run(string(tt.ns), func(t *testing.T) {
			got := FormatComment(tt.ns)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{
			name:    "valid task ID",
			id:      "TSK0Ij1P13FRDM",
			wantErr: false,
		},
		{
			name:    "too short",
			id:      "TSK0Ij1P13FR",
			wantErr: true,
		},
		{
			name:    "invalid characters",
			id:      "TSK0Ij1P13FR!M",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Validate(tt.id)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateWithNamespace(t *testing.T) {
	_, err := ValidateWithNamespace("TSK0Ij1P13FRDM", fiberfx.NS_TASK)
	require.NoError(t, err)

	_, err = ValidateWithNamespace("TSK0Ij1P13FRDM", fiberfx.NS_EPIC)
	require.Error(t, err)
}

func TestCheckConstraintExpr(t *testing.T) {
	expr := CheckConstraintExpr("id", fiberfx.NS_TASK)
	require.Equal(t, "id ~ '^TSK[0-9A-Za-z]{11}$'", expr)
}

func TestCheckConstraintName(t *testing.T) {
	name := CheckConstraintName("tasks", "id")
	require.Equal(t, "chk_tasks_id_format", name)
}
