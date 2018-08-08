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

func (self *ClientConfig) SetEndpoints(endpoints []string) {
	self.Endpoints = endpoints
}

func (self *ClientConfig) SetDialTimeout(dialTime int) {
	self.DialTimeout = dialTime
}

func (self *ClientConfig) SetRequestTimeout(requestTimeout int) {
	self.RequestTimeout = requestTimeout
}

func (self *ClientConfig) GetRequestTimeout() time.Duration {
	return time.Duration(self.RequestTimeout) * time.Second
}

func (self *ClientConfig) GetDialTimeout() time.Duration {
	return time.Duration(self.DialTimeout) * time.Second
}

func (self *ClientConfig) GetEndpoints() []string {
	return self.Endpoints
}

func (self *Config) GetApiAddress() string {
	return self.ConnectionConf.Rest.Address
}

func (self *Config) GetClientConfig() *ClientConfig {
	return &self.ClientConf
}

func (self *Config) GetConnectionConfig() *ConnectionConfig {
	return &self.ConnectionConf
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
