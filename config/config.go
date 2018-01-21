package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type celestial struct {
	Conf Config `yaml:"celestial"`
}

type Config struct {
	Endpoints      []string `yaml:"endpoints"`
	DialTimeout    int32    `yaml:"dialtimeout"`
	RequestTimeout int32    `yaml:"requesttimeout"`
}

func ReadConfig(n ...string) (*Config, error) {
	path := "config.yml"
	if len(n) > 0 {
		path = n[0]
	}

	yamlFile, err := ioutil.ReadFile(path)
	check(err)

	var conf celestial
	err = yaml.Unmarshal(yamlFile, &conf)
	check(err)

	return &conf.Conf, nil
}

func DefaultConfig() *Config {
	conf := Config{
		Endpoints:      []string{"0.0.0.0:2379"},
		DialTimeout:    2,
		RequestTimeout: 10,
	}

	return &conf
}

func (self *Config) SetEndpoints(endpoints []string) {
	self.Endpoints = endpoints
}

func (self *Config) SetDialTimeout(dialTime int) {
	self.DialTimeout = dialTime
}

func (self *Config) RequestTimeout(requestTimeout int) {
	self.RequestTimeout = requestTimeout
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
