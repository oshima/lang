package sema

import (
	"errors"

	"github.com/oshima/lang/ast"
)

// env represents a scope of names.
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

func (e *env) set(name string, node ast.Node) error {
	if _, ok := e.store[name]; ok {
		return errors.New("duplicate entries")
	}
	e.store[name] = node
	return nil
}

func (e *env) get(name string) (ast.Node, bool) {
	node, ok := e.store[name]
	if !ok && e.outer != nil {
		node, ok = e.outer.get(name)
	}
	return node, ok
}
