package main

import (
	"fmt"
	"github.com/c12s/celestial/model/config"
	"github.com/c12s/celestial/service"
	"github.com/c12s/celestial/storage/etcd"
	"log"
)

func main() {
	// Run()
	// Load configurations
	conf, err := config.ConfigFile()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(conf)

	//Load database
	db, err := etcd.New(conf.GetClientConfig())
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(db)

	//Start Server
	service.Run(db, conf.GetConnectionConfig())
}
