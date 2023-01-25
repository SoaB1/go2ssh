package config

var Conf Config

type Config struct {
	SSHConfigs sshConfig
}

type sshConfig struct {
	Hosts    string `env:"SCHEMA" envDefault:""`
	Hostname string `env:"USER" envDefault:"localhost"`
	Server   string `env:"PASSWORD" envDefault:"localhost"`
	Port     string `env:"PORT" envDefault:"22"`
	User     string `env:"USER" envDefault:""`
	Name     string `env:"NAME" envDefault:"root"`
	Keypath  string `env:"KEY-PATH" envDefault:"$HOME/.ssh/id_rsa"`
}
