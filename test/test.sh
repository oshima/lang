#!/bin/bash

try() {
  input="$1"
  expected="$2"
  echo "$input" | lang > tmp.s
  gcc -no-pie -o tmp tmp.s
  ./tmp
  actual="$?"
  if [ "$actual" != "$expected" ]; then
    echo "$input => Expected $expected but got $actual"
    exit 1
  fi
}

try-file() {
  file="$1"
  expected="$2"
  cat "$file" | lang > tmp.s
  gcc -no-pie -o tmp tmp.s
  actual=`./tmp`
  if [ "$actual" != "$expected" ]; then
    echo "$file => Expected $expected but got $actual"
    exit 1
  fi
}

try "0;" 0
try "42;" 42
try "10; 100;" 100

try "-0;" 0
try "-5;" 251 # -5

try "7 + 8;" 15
try "1+2+3+4+5+6+7+8+9+10;" 55
try "1 - 6;" 251 # -5
try "-1 - -6;" 5
try "1 - 2 + 3 - 4 + 5;" 3
try "11; 12 + 13;" 25
try "12 + 13; 11;" 11

try "7-3;" 4
try "7 -3;" 4
try "7+-3;" 4
try "7--3;" 10

try "3 * 10;" 30
try "-6 * -9;" 54
try "6*2-3;" 9
try "6+2*3;" 12
try "27 / 3;" 9
try "-42/-6;" 7
try "3+12*8-9/3;" 96
try "16 % 2;" 0
try "11 % -7;" 4

try "(3+12)*8-(9/3);" 117
try "4+(6+8*2)/2;" 15
try "((((26))))-3;" 23

try "- 1;" 255 # -1
try "--1;" 1
try "- -1;" 1
try "-(1-2);" 1
try "- 1-2;" 253 # -3

try "true;" 1
try "false;" 0
try "1; false;" 0

try '!true;' 0
try '!false;' 1
try '!!true;' 1
try '!(!true);' 1

try "true && true;" 1
try "true && false;" 0
try "false && false;" 0
try "true || true;" 1
try "true || false;" 1
try "false || false;" 0
try "true || false && false;" 1
try "(true || false) && false;" 0

try "33 == 33;" 1
try "4 == 29;" 0
try "-89 != -3;" 1
try "10 != 10;" 0
try "1 == 2 == false;" 1
try "true == (2 == 0);" 0

try "4 < 2 + 3;" 1
try "4 <= 2 + 3;" 1
try "-5 < -(2 + 3);" 0
try "-5 <= -(2 + 3);" 1
try "2 + 3 > 4;" 1
try "2 + 3 >= 4;" 1
try "-(2 + 3) > -5;" 0
try "-(2 + 3) >= -5;" 1

try "{ 10; 20; }" 20
try "if true { 10; }" 10
try "if true { 10; } else { 20; }" 10
try "if false { 10; } else { 20; }" 20
try "if true { if false { 10; } else { 20; } } else { 30; }" 20
try "if false { if false { 10; } else { 20; } } else { 30; }" 30
try "if false { 10; } else if false { 20; } else { 30; }" 30

try "var x int = 3 + 4; x;" 7
try "var x int = 10; { var x int = 20; x; }" 20
try "var x int = 10; { var x int = 20; } x;" 10

try "var x = 2; x = x * 2; x = x * 2; x;" 8
try "var x = false; x = x || true; x;" 1
try "var x = 10; { var x = 20; x = x + 10; x; }" 30
try "var x = 10; { var x = 20; x = x + 10; } x;" 10

try-file ./test/for1 55
try-file ./test/for2 55
try-file ./test/for3 225

try-file ./test/func1 40
try-file ./test/func2 true
try-file ./test/func3 15
try-file ./test/func4 91
try-file ./test/func-fib 102334155

try-file ./test/array1 6,15
try-file ./test/array2 6

echo OK
