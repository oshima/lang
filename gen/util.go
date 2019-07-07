package gen

import (
	"github.com/oshima/lang/token"
	"github.com/oshima/lang/types"
)

var setcc = map[token.Type]string{
	token.EQ: "sete",
	token.NE: "setne",
	token.LT: "setl",
	token.LE: "setle",
	token.GT: "setg",
	token.GE: "setge",
}

var paramRegs = map[int][6]string{
	1: [6]string{"dil", "sil", "dl", "cl", "r8b", "r9b"},
	8: [6]string{"rdi", "rsi", "rdx", "rcx", "r8", "r9"},
}

func sizeOf(typ types.Type) int {
	switch typ.(type) {
	case *types.Int:
		return 8
	case *types.Bool:
		return 1
	case *types.String:
		return 8
	case *types.Range:
		return 8
	case *types.Array:
		return 8
	case *types.Func:
		return 8
	default:
		return 0 // unreachable
	}
}

// https://en.wikipedia.org/wiki/Data_structure_alignment
func align(n int, boundary int) int {
	return (n + boundary - 1) & -boundary
}
