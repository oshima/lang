package ast

import "github.com/oshima/lang/types"

type Metadata struct {
	Refs  map[Node]Node
	Types map[Expr]types.Type
}
