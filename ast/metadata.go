package ast

import "github.com/oshjma/lang/types"

type Metadata struct {
	Refs  map[Node]Node
	Types map[Expr]types.Type
}
