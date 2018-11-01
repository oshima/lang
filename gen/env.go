package gen

import "errors"

type env struct {
	gvars map[string]*gvar
	dests map[string]*dest
	outer *env
}

// Global variable
type gvar struct {
	label string
	size  int
}

// Destination of jump instruction
type dest struct {
	label string
}

func newEnv(outer *env) *env {
	return &env{
		gvars: make(map[string]*gvar),
		dests: make(map[string]*dest),
		outer: outer,
	}
}

func (e *env) setGvar(name string, v *gvar) error {
	if _, ok := e.gvars[name]; ok {
		return errors.New("Duplicate identifier")
	}
	e.gvars[name] = v
	return nil
}

func (e *env) getGvar(name string) (*gvar, bool) {
	v, ok := e.gvars[name]
	if !ok && e.outer != nil {
		v, ok = e.outer.getGvar(name)
	}
	return v, ok
}

func (e *env) setDest(keyword string, d *dest) {
	e.dests[keyword] = d
}

func (e *env) getDest(keyword string) (*dest, bool) {
	d, ok := e.dests[keyword]
	if !ok && e.outer != nil {
		d, ok = e.outer.getDest(keyword)
	}
	return d, ok
}
