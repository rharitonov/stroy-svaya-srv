package config

type Config struct {
	ServerAddress string
	DatabasePath  string
	DatabaseUrl   string
}

func Load() *Config {
	return &Config{
		ServerAddress: ":8080",
		DatabasePath:  "./db/stroy-svaya.db",
		DatabaseUrl:   "sqlite://db/stroy-svaya.db",
	}
}
