package main

import (
	"github.com/usememos/memos/bin"
)

func main() {
	if err := bin.Execute(); err != nil {
		panic(err)
	}
}
