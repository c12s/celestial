package config

type Celestial struct {
	Conf Config `yaml:"celestial"`
}

type Config struct {
	ConfVersion    string       `yaml:"version"`
	Address        string       `yaml:"address"`
	Apollo         string       `yaml:"apollo"`
	ClientConf     ClientConfig `yaml:"client"`
	Endpoints      []string     `yaml:"db"`
	SEndpoints     []string     `yaml:"sdb"`
	DialTimeout    int          `yaml:"dialtimeout"`
	RequestTimeout int          `yaml:"requesttimeout"`
}

type ClientSecurity struct {
	Cert    string `yaml:"cert"`
	Key     string `yaml:"key"`
	Trusted string `yaml:"trusted"`
}

type ClientConfig struct {
	Security ClientSecurity `yaml:"security"`
}
