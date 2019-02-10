# lang

lang is a toy language compiled to assembly code for x86_64-linux.

## Syntax

### Literals

lang has six types of values: `int`, `bool`, `string`, `range`, `array` and `function`.
There are literals for representing values of each types.

```go
// int (signed 64-bit)
42
-7

// bool
true
false

// string
"foo"

// range
1..100

// array
["apple", "banana", "orange"] // [3]string

// function
(a: int, b: int) -> int { return a + b; } // (int, int) -> int
```

### Operators

lang has the following operators:

```go
// arithmetic operators
5 + 2
5 - 2
5 * 2
5 / 2
5 % 2
-(5)

// comparison operators
5 == 2
5 != 2
5 < 2
5 <= 2
5 > 2
5 >= 2

// logical operators
true && false
true || false
!true

// `in` operator
3 in 1..5 // => true
6 in 1..5 // => false
3 in [1, 2, 3, 4, 5] // => true
6 in [1, 2, 3, 4, 5] // => false
```

Operator priority is similar to other languages.

```go
1 + 2 * 3   // => 7
(1 + 2) * 3 // => 9
```


### Variables

Using `var` statement, we can declare a variable that holds a value.

```go
var num: int = 3 + 5;

```

Variable type annotation can be omitted.
In which case, it is inferred by the type of initial value.

```go
var name = "bar", age = 20;   // name: string, age: int

var primes = [2, 3, 5, 7, 9]; // primes: [5]int
```

A value of variable can be reassigned.

```go
var num = 0;
num = 3;
num += 7;
printf("%d\n", num) // => 10
```

### Functions

Using `func` statement, we can declare named function.
It is required to annotate the types of parameters and return value.

```go
// mul: (int, int) -> int
func mul(a: int, b: int) -> int {
  return a * b;
}

printf("%d\n", mul(3, 5)); // => 15
```

Function literal generates anonymous function.

```go
(name: string) -> {
  printf("Hello, %s\n", name);
}("baz"); // => Hello, baz
```

Anonymous function can be stored in a variable and called by the variable name.

```go
var fib = (n: int) -> int {
  if n < 2 {
    return n;
  }
  return fib(n - 2) + fib(n - 1); // recursive call
}

printf("%d\n", fib(10)); // => 55
```

### Flow control

lang has `if`, `while` and `for` statements like other languages.

```js
if true {
  puts("foo");
}
// => foo

if false {
  puts("foo");
} else {
  puts("bar");
}
// => bar

var n = 7;

if n in 1..4 {
  puts("foo");
} else if n in 5..9 {
  puts("bar");
} else {
  puts("baz");
}
// => bar
```

```js
var n = 1;

while true {
  if n % 2 == 0 {
    continue;
  }
  printf("%d ", n);
  n += 1;
  if n >= 10 {
    break;
  }
}
// => 1 3 5 7 9
```

``` js
for n in 1..5 {
  printf("%d ", n);
}
// => 1 2 3 4 5

for s in ["a", "b", "c"] {
  printf("%s ", s);
}
// => a b c
```

## References

- [Writing An Interpreter In Go](https://interpreterbook.com/)
- [An Incremental Approach to Compiler Construction](http://scheme2006.cs.uchicago.edu/11-ghuloum.pdf)
- [x86-64 Assembly Language Programming with Ubuntu](http://www.egr.unlv.edu/~ed/x86.html)
- [x86 and amd64 instruction reference](https://www.felixcloutier.com/x86/)
- [Compiler Explorer](https://godbolt.org/)
- [rui314/9cc: A Small C Compiler](https://github.com/rui314/9cc)
