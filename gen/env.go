package gen

import "errors"

type env struct {
	store map[string]*gvar
	outer *env
}

func (e *env) set(name string, v *gvar) error {
	if _, ok := e.store[name]; ok {
		return errors.New("Duplicate identifier")
	}
	e.store[name] = v
	return nil
}

func (e *env) get(name string) (*gvar, bool) {
	v, ok := e.store[name]
	if !ok && e.outer != nil {
		v, ok = e.outer.get(name)
	}
	return v, ok
}

type gvar struct {
	label string
	size int
}
