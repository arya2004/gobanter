package config

import "flag"

type Config struct {
	Port string
}

func Load() Config {
	cfg := Config{}

	flag.StringVar(&cfg.Port, "port", "8080", "Port to run the server on")
	flag.Parse()

	return cfg
}
