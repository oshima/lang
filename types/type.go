package types

import (
	"fmt"
	"strings"
)

// ----------------------------------------------------------------
// Interface

// Type is the interface for all types in this language.
type Type interface {
	String() string
}

// ----------------------------------------------------------------
// Types

// Int represents the integer type.
type Int struct{}

func (i *Int) String() string {
	return "int"
}

// Bool represents the boolean type.
type Bool struct{}

func (b *Bool) String() string {
	return "bool"
}

// String represents the string type.
type String struct{}

func (s *String) String() string {
	return "string"
}

// Range represents the range type.
type Range struct{}

func (r *Range) String() string {
	return "range"
}

// Array represents the array type.
type Array struct {
	Len      int
	ElemType Type
}

func (a *Array) String() string {
	return fmt.Sprintf("[%d]%s", a.Len, a.ElemType)
}

// Func represents the function type.
type Func struct {
	ParamTypes []Type
	ReturnType Type
}

func (f *Func) String() string {
	var params []string
	for _, typ := range f.ParamTypes {
		params = append(params, typ.String())
	}
	if f.ReturnType == nil {
		return fmt.Sprintf("(%s) -> void", strings.Join(params, ", "))
	}
	return fmt.Sprintf("(%s) -> %s", strings.Join(params, ", "), f.ReturnType)
}
