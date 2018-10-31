package gen

var sizes = map[string]int{
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
