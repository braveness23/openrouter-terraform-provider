//go:build tools

// Package tools pins development tool dependencies so they appear in go.mod/go.sum.
package tools

import (
	// tfplugindocs generates the docs/ directory from schema descriptions and examples/.
	_ "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs"
)
