package gen

import (
	"errors"
	"github.com/oshjma/lang/ast"
)

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

func (e *env) set(key string, node ast.Node) error {
	if _, ok := e.store[key]; ok {
		return errors.New("Duplicate entries")
	}
	e.store[key] = node
	return nil
}

func (e *env) get(key string) (ast.Node, bool) {
	node, ok := e.store[key]
	if !ok && e.outer != nil {
		node, ok = e.outer.get(key)
	}
	return node, ok
}
