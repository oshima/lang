# lang

lang is a toy language compiled into assembly code targetting x86_64-linux.

## Syntax

### Literals

lang has six types of values: `int`, `bool`, `string`, `range`, `array` and `function`.\
Each type has the literal to represent its value.

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
0..100

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
2 in 0..5 // => true
5 in 0..5 // => false
2 in [0, 1, 2, 3, 4] // => true
5 in [0, 1, 2, 3, 4] // => false
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

Variable type annotation can be omitted.\
In which case, it is inferred by the type of initial value.

```go
var name = "foo", age = 20;    // name: string, age: int

var primes = [2, 3, 5, 7, 11]; // primes: [5]int
```

A value of variable can be reassigned.

```go
var num = 0;
num = 3;
num += 7;
printf("%d\n", num) // => 10
```

### Functions

Using `func` statement, we can declare named function.\
It is required to annotate the types of parameters and return value.

```go
func mul(a: int, b: int) -> int {
  return a * b;
}

printf("%d\n", mul(3, 5)); // => 15
```

Function literal generates an anonymous function.

```go
(name: string) -> {
  printf("Hello, %s\n", name);
}("foo");
// => Hello, foo
```

Anonymous function can be stored in a variable and called by its name.

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

var n = 30;

if n in 0..10 {
  puts("small");
} else if n in 10..20 {
  puts("medium");
} else {
  puts("large");
}
// => large
```

```js
var n = 0;

while n < 10 {
  if n % 2 == 0 {
    continue;
  }
  printf("%d ", n);
  n += 1
}
// => 1 3 5 7 9

var n = 0;

while true {
  printf("%d ", n);
  n += 1;
  if n >= 5 {
    break;
  }
}
// => 0 1 2 3 4
```

``` js
for n in 0..5 {
  printf("%d ", n);
}
// => 0 1 2 3 4

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
