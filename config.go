package main

import (
	env "github.com/caarlos0/env/v6"
	_ "github.com/joho/godotenv/autoload"
)

// Config is the main application config.
type Config struct {
	Port          int    `env:"PORT" envDefault:"8080"`
	MemoryStorage bool   `env:"MEMORY_STORAGE" envDefault:"false"`
	DataDir       string `env:"DATA_DIR" envDefault:"./data"`
	TemplatesDir  string `env:"TEMPLATES_DIR" envDefault:"./templates"`
	StaticDir     string `env:"STATIC_DIR" envDefault:"./static"`
}

// ReadConfig reads config from env.
func ReadConfig() (Config, error) {
	var conf Config
	if err := env.Parse(&conf); err != nil {
		return Config{}, err // nolint: wrapcheck
	}
	if conf.MemoryStorage {
		conf.DataDir = ""
	}
	return conf, nil
}
