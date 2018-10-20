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

echo OK
rm -f tmp*
