package ast

import "github.com/oshima/lang/types"

// Metadata holds metadata of each AST node.
type Metadata struct {
	Refs  map[Node]Node
	Types map[Expr]types.Type
}
