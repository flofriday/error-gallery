package main

import (
	"fmt"
)

func hello(name string) string {
	return fmt.Sprintf("Hello %s!", name)
}

func main() {
	greeting := hallo("friday")
	fmt.Println(greeting)
}
