package main

import (
	// "github.com/c12s/celestial/client"
	"fmt"
	"github.com/c12s/celestial/config"
)

func main() {
	conf, err := config.ConfigFile()
	if err != nil {
		fmt.Println("Error")
	} else {
		fmt.Println(conf)
	}

	// c := client.NewClient(conf)
	// defer c.Close()

	// c.PrintClusterNodes("novisad", "grbavica")
}
