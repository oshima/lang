package gen

// global variable
type gvar struct {
	label string
	size  int
}

// global range
type grng struct {
	label string
}

// global array
type garr struct {
	label    string
	len      int
	elemSize int
}

// local variable
type lvar struct {
	offset int
	size   int
}

// local range
type lrng struct {
	offset int
}

// local array
type larr struct {
	offset   int
	len      int
	elemSize int
}

// string
type str struct {
	label string
	value string
}

// function
type fn struct {
	label     string
	localArea int
}

// branch labels
type br struct {
	labels []string
}
