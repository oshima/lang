#!/bin/bash

try() {
  input="$1"
  expected="$2"
  echo "$input" | lang > tmp.s
  gcc -o tmp tmp.s
  ./tmp
  actual="$?"
  if [ "$actual" != "$expected" ]; then
    echo "Expected $expected but got $actual"
    rm -f tmp*
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

echo OK
rm -f tmp*
