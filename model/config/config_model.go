package config

type Celestial struct {
	Conf Config `yaml:"celestial"`
}

type Config struct {
	ConfVersion    string       `yaml:"version"`
	Address        string       `yaml:"address"`
	ClientConf     ClientConfig `yaml:"client"`
	Endpoints      []string     `yaml:"db"`
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
