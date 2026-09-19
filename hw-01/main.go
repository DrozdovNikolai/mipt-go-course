package main

import (
	"fmt"
	"os"
	"runtime"
)

func main() {
	fmt.Println("Username:", os.Getenv("USERNAME"))
	fmt.Println("Arguments:", os.Args[1:])
	fmt.Println("Go version:", runtime.Version())
}
