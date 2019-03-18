package types

import (
	"fmt"
	"strings"
)

// ----------------------------------------------------------------
// Interface

// Type is the interface for all types in this language
type Type interface {
	String() string
}

// ----------------------------------------------------------------
// Types

// Int represents the integer type
type Int struct{}

// Bool represents the boolean type
type Bool struct{}

// String represents the string type
type String struct{}

// Range represents the range type
type Range struct{}

// Array represents the array type
type Array struct {
	Len      int
	ElemType Type
}

// Func represents the function type
type Func struct {
	ParamTypes []Type
	ReturnType Type
}

func (i *Int) String() string {
	return "int"
}

func (b *Bool) String() string {
	return "bool"
}

func (s *String) String() string {
	return "string"
}

func (r *Range) String() string {
	return "range"
}

func (a *Array) String() string {
	return fmt.Sprintf("[%d]%s", a.Len, a.ElemType)
}

func (f *Func) String() string {
	var params []string
	for _, ty := range f.ParamTypes {
		params = append(params, ty.String())
	}
	if f.ReturnType == nil {
		return fmt.Sprintf("(%s) -> {}", strings.Join(params, ", "))
	}
	return fmt.Sprintf("(%s) -> %s", strings.Join(params, ", "), f.ReturnType)
}
