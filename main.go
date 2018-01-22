package main

import (
	"github.com/c12s/celestial/client"
	"github.com/c12s/celestial/config"
	"log"
)

func main() {
	conf, err := config.ConfigFile()
	if err != nil {
		log.Fatal(err)
	}

	c := client.NewClient(conf)
	defer c.Close()

	c.PrintClusterNodes("novisad", "grbavica")
}
