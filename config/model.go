package config

type Celestial struct {
	Conf Config `yaml:"celestial"`
}

type Config struct {
	ConfVersion    string           `yaml:"version"`
	ClientConf     ClientConfig     `yaml:"client"`
	ConnectionConf ConnectionConfig `yaml:"connection"`
}

type ClientSecurity struct {
	Cert    string `yaml:"cert"`
	Key     string `yaml:"key"`
	Trusted string `yaml:"trusted"`
}

type ClientConfig struct {
	Security       ClientSecurity `yaml:"security"`
	Endpoints      []string       `yaml:"endpoints"`
	DialTimeout    int            `yaml:"dialtimeout"`
	RequestTimeout int            `yaml:"requesttimeout"`
}

type ConnectionConfig struct {
	Rpc        RPC  `yaml:"rpc"`
	Rest       REST `yaml:"rest"`
	Standalone bool `yaml:"standalone"`
}

type RPC struct {
	Address string `yaml:"address"`
}

type REST struct {
	Address string `yaml:"address"`
}
