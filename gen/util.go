package gen

var sizeof = map[string]int{
	"int":    8,
	"bool":   1,
	"string": 8,
}

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

var libFns = map[string]bool{
	"puts":   true,
	"printf": true,
}

// https://en.wikipedia.org/wiki/Data_structure_alignment
func align(n int, boundary int) int {
	return (n + boundary - 1) & -boundary
}
