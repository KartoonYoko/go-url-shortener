package main

import "os"

func main() {
	os.Exit(1) // want "os exit in main function of main package"
}