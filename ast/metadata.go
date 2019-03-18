package ast

import "github.com/oshima/lang/types"

// Metadata of the AST
type Metadata struct {
	Refs  map[Node]Node
	Types map[Expr]types.Type
}
