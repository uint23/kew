package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "usage: kew in/ out/\n")
		os.Exit(1)
	}

	src := os.Args[1]
	out := os.Args[2]

	fmt.Println("src:", src)
	fmt.Println("out:", out)
}
