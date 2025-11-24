package config

type Config struct {
	Port  string
	DBUrl string
}

func Load() *Config {
	return &Config{
		Port:  ":8080",
		DBUrl: "postgresql://postgres:123@localhost:5432/postgres",
	}
}
