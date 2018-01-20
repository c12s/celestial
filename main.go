package main

import "fmt"

type Test struct {
	Name string
}

func main() {
	fmt.Println("Hello world!")

	t := getTest("hello")

	fmt.Println(t)
}

func getTest(name string) *Test {
	test := Test{
		Name: name,
	}

	return &test
}
