#!/bin/sh

function comment() {
  while IFS= read -r line
  do
    echo "// $line"
  done
}

echo "//go:generate .script/doc.sh" > doc.go

echo "" >> doc.go

go run main.go --help | comment >> doc.go

echo "package main" >> doc.go
