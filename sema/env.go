package sema

import (
	"errors"
	"github.com/oshjma/lang/ast"
)

/*
 Environment - create the scope of name bindings
*/

type env struct {
	store map[string]ast.Node
	outer *env
}

func newEnv(outer *env) *env {
	return &env{
		store: make(map[string]ast.Node),
		outer: outer,
	}
}

func (e *env) set(ident string, node ast.Node) error {
	if _, ok := e.store[ident]; ok {
		return errors.New("Duplicate entries")
	}
	e.store[ident] = node
	return nil
}

func (e *env) get(ident string) (ast.Node, bool) {
	node, ok := e.store[ident]
	if !ok && e.outer != nil {
		node, ok = e.outer.get(ident)
	}
	return node, ok
}
