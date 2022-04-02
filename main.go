package main

import "fmt"

// injected at build time
var (
	version string
	gitHash string
	date    string
)

func main() {
	fmt.Println(version, gitHash, date)
}
