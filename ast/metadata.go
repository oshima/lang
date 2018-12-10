package ast

type Metadata struct {
	Refs  map[Node]Node
	Types map[Expr]string
}
