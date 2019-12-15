package main

import (
	"compiler/cmd"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	cmd.Execute()
}
