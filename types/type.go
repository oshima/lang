package types

import (
	"fmt"
	"strings"
)

/*
 Interface
*/

type Type interface {
	String() string // for error messages
}

/*
 Basic types
*/

type Int struct{}

type Bool struct{}

type String struct{}

func (i *Int) String() string    { return "int" }
func (b *Bool) String() string   { return "bool" }
func (s *String) String() string { return "string" }

/*
 Array type
*/

type Array struct {
	Len      int
	ElemType Type
}

func (a *Array) String() string {
	return fmt.Sprintf("[%d]%s", a.Len, a.ElemType)
}

/*
 Func type
*/

type Func struct {
	ParamTypes []Type
	ReturnType Type
}

func (f *Func) String() string {
	var params []string
	for _, ty := range f.ParamTypes {
		params = append(params, ty.String())
	}
	if f.ReturnType == nil {
		return fmt.Sprintf("(%s) -> {}", strings.Join(params, ", "))
	} else {
		return fmt.Sprintf("(%s) -> %s", strings.Join(params, ", "), f.ReturnType)
	}
}
