package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"time"
)

func ConfigFile(n ...string) (*Config, error) {
	path := "config.yml"
	if len(n) > 0 {
		path = n[0]
	}

	yamlFile, err := ioutil.ReadFile(path)
	check(err)

	var conf Celestial
	err = yaml.Unmarshal(yamlFile, &conf)
	check(err)

	return &conf.Conf, nil
}

func DefaultConfig() *Config {
	sec := ClientSecurity{
		Cert:    "",
		Key:     "",
		Trusted: "",
	}

	rpc := RPC{
		Address: "",
	}

	rest := REST{
		Address: "",
	}

	conn := ConnectionConfig{
		Rpc:        rpc,
		Rest:       rest,
		Standalone: true,
	}

	client := ClientConfig{
		Security:       sec,
		Endpoints:      []string{"0.0.0.0:2379"},
		DialTimeout:    2,
		RequestTimeout: 10,
	}

	conf := Config{
		ConfVersion:    "v1",
		ClientConf:     client,
		ConnectionConf: conn,
	}

	return &conf
}

func (self *Config) SetEndpoints(endpoints []string) {
	self.ClientConf.Endpoints = endpoints
}

func (self *Config) SetDialTimeout(dialTime int) {
	self.ClientConf.DialTimeout = dialTime
}

func (self *Config) SetRequestTimeout(requestTimeout int) {
	self.ClientConf.RequestTimeout = requestTimeout
}

func (self *Config) GetRequestTimeout() time.Duration {
	return time.Duration(self.ClientConf.RequestTimeout) * time.Second
}

func (self *Config) GetDialTimeout() time.Duration {
	return time.Duration(self.ClientConf.DialTimeout) * time.Second
}

func (self *Config) GetEndpoints() []string {
	return self.ClientConf.Endpoints
}

func (self *Config) GetApiAddress() string {
	return self.ConnectionConf.Rest.Address
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
