package gen

import "github.com/oshjma/lang/types"

var setcc = map[string]string{
	"==": "sete",
	"!=": "setne",
	"<":  "setl",
	"<=": "setle",
	">":  "setg",
	">=": "setge",
}

var paramRegs = map[int][6]string{
	1: [6]string{"dil", "sil", "dl", "cl", "r8b", "r9b"},
	8: [6]string{"rdi", "rsi", "rdx", "rcx", "r8", "r9"},
}

func sizeOf(ty types.Type) int {
	switch ty.(type) {
	case *types.Int:
		return 8
	case *types.Bool:
		return 1
	case *types.String:
		return 8
	case *types.Array:
		return 8
	default:
		return 0 // unreachable here
	}
}

// https://en.wikipedia.org/wiki/Data_structure_alignment
func align(n int, boundary int) int {
	return (n + boundary - 1) & -boundary
}
