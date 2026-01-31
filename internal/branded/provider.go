// Copyright 2021-present The Atlas Authors. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package branded

import (
	"github.com/jonny-novikov/jonnify/fiberfx"
)

// Provider generates and parses branded IDs.
// It wraps fiberfx.Generator for use in Atlas operations.
type Provider struct {
	gen *fiberfx.Generator
}

// NewProvider creates a provider with the given worker ID.
// Worker ID should be unique per process/node (0-1023).
func NewProvider(workerID int64) *Provider {
	return &Provider{
		gen: fiberfx.NewGenerator(workerID),
	}
}

// DefaultProvider returns a provider using the default generator (worker ID 0).
func DefaultProvider() *Provider {
	return &Provider{
		gen: fiberfx.DefaultGenerator(),
	}
}

// Generate creates a new branded ID for the namespace.
func (p *Provider) Generate(ns fiberfx.Namespace) fiberfx.ID {
	return p.gen.New(ns)
}

// Parse validates and parses a branded ID string.
func (p *Provider) Parse(value string) (fiberfx.ID, error) {
	return fiberfx.Parse(value)
}

// ParseWithNamespace validates against expected namespace.
func (p *Provider) ParseWithNamespace(value string, expected fiberfx.Namespace) (fiberfx.ID, error) {
	return fiberfx.ParseWithNamespace(value, expected)
}

// IsValid checks if a string is a valid branded ID format.
func (p *Provider) IsValid(value string) bool {
	return fiberfx.Valid(value)
}

// Generator returns the underlying Generator.
func (p *Provider) Generator() *fiberfx.Generator {
	return p.gen
}
