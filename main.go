package main

import (
	"github.com/c12s/celestial/model/config"
	"github.com/c12s/celestial/service"
	"github.com/c12s/celestial/storage/etcd"
	"github.com/c12s/celestial/storage/vault"
	"log"
	"time"
)

func main() {
	// Run()
	// Load configurations
	conf, err := config.ConfigFile()
	if err != nil {
		log.Fatal(err)
	}

	//Load database
	db, err := etcd.New(conf.Endpoints, 10*time.Second)
	if err != nil {
		log.Fatal(err)
	}

	//Load secrets database
	sdb, err := vault.New(conf.SEndpoints, 10*time.Second)
	if err != nil {
		log.Fatal(err)
	}

	//Start Server
	service.Run(db, sdb, conf.Address)
}
