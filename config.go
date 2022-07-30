package main

import (
	"errors"
	"fmt"
	"strconv"

	env "github.com/caarlos0/env/v6"
	_ "github.com/joho/godotenv/autoload"
)

// Config is the main application config.
type Config struct {
	Port            int    `env:"PORT" envDefault:"8080"`
	DataFile        string `env:"DATA_FILE" envDefault:"./data.db"`
	DataFileMaxSize Size   `env:"DATA_FILE_MAX_SIZE" envDefault:"100M"`
	TemplatesDir    string `env:"TEMPLATES_DIR" envDefault:"./templates"`
	StaticDir       string `env:"STATIC_DIR" envDefault:"./static"`
}

// ReadConfig reads config from env.
func ReadConfig() (Config, error) {
	var conf Config
	if err := env.Parse(&conf); err != nil {
		return Config{}, err // nolint: wrapcheck
	}
	return conf, nil
}

// Size is a size in bytes. Used for parsing config values from string
// like 10B, 100K, 1000G, etc.
type Size int64

// SizeUnits is a set of available units for `Size` type.
// nolint: gochecknoglobals
var SizeUnits = map[string]int64{
	"B": 1,
	"K": 1024,
	"M": 1024 * 1024,
	"G": 1024 * 1024 * 1024,
	"T": 1024 * 1024 * 1024 * 1024,
}

// UnmarshalText unmarshals size from its string representation.
func (s *Size) UnmarshalText(text []byte) error {
	if len(text) < 2 {
		return errors.New("invalid size string")
	}

	chars := []rune(string(text))

	unit := string(chars[len(chars)-1])
	mul, ok := SizeUnits[unit]
	if !ok {
		return fmt.Errorf("unknown unit '%s'", unit)
	}

	nums := string(chars[:len(chars)-1])
	n, err := strconv.ParseInt(nums, 10, 64)
	if err != nil {
		return fmt.Errorf("size '%s' is not a number", nums)
	}

	*s = Size(n * mul)
	return nil
}
