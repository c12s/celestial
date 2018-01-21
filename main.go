package main

import (
	"github.com/c12s/celestial/client"
	"github.com/c12s/celestial/config"
)

func main() {
	conf := config.DefaultConfig()
	c := client.NewClient(conf)
	defer c.Close()

	c.PrintClusterNodes("novisad", "grbavica")
}
