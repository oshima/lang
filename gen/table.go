package gen

var sizeof = map[string]int{
	"int":  8,
	"bool": 1,
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
	8: [6]string{"rdi", "rsi", "rdx", "rcx", "r8", "r9"},
	1: [6]string{"dil", "sil", "dl", "cl", "r8b", "r9b"},
}
