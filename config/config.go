package config

import (
	"github.com/BurntSushi/toml"
	options "github.com/mreiferson/go-options"

	"flag"
	"fmt"
	"os"
)

const (
	defaultTcpAddress = "localhost:5432"
	defaultConfigFile = "./config.toml"
)

type Config struct {
	PgAddress      string `flag:"pg-address" cfg:"pg-address"`
	PgUsername     string `flag:"pg-username" cfg:"pg-username"`
	PgPassword     string `flag:"pg-password" cfg:"pg-password"`
	PgDatabaseName string `flag:"pg-db-name" cfg:"pg-db-name"`
	PgPoolsize     int    `flag:"pg-poolsize" cfg:"pg-poolsize"`
	PgReadTimeout  int    `flag:"pg-readtimeout" cfg:"pg-readtimeout"`
	PgWriteTimeout int    `flag:"pg-writetimeout" cfg:"pg-writetimeout"`
	PgIdleTimeout  int    `flag:"pg-idletimeout" cfg:"pg-idletimeout"`
	InitDB         bool
}

func NewConfig() *Config {
	flagSet := consoleConfig()
	flagSet.Parse(os.Args[1:])

	config := defaultConfig()

	options.Resolve(config, flagSet, fileConfig(defaultConfigFile))

	initDbFlag := flagSet.Lookup("init")
	config.InitDB = initDbFlag.Value.(flag.Getter).Get().(bool)

	verbose := flagSet.Lookup("verbose")
	if verbose != nil && verbose.Value.(flag.Getter).Get().(bool) {
		fmt.Printf("pg-address: %s\n", config.PgAddress)
		fmt.Printf("pg-username: %s\n", config.PgUsername)
		fmt.Printf("pg-password: %s\n", config.PgPassword)
		fmt.Printf("pg-db-name: %s\n", config.PgDatabaseName)
		fmt.Printf("pg-poolsize: %d\n", config.PgPoolsize)
		fmt.Printf("pg-readtimeout: %d\n", config.PgReadTimeout)
		fmt.Printf("pg-writetimeout: %d\n", config.PgWriteTimeout)
		fmt.Printf("pg-idletimeout: %d\n", config.PgIdleTimeout)
		fmt.Printf("init: %t\n", config.InitDB)
	}
	return config
}

func defaultConfig() *Config {
	return &Config{defaultTcpAddress, "", "", "", 10, 5, 5, 5, false}
}

func fileConfig(configFile string) map[string]interface{} {
	var v map[string]interface{}
	_, err := toml.DecodeFile(configFile, &v)
	if err != nil {
		fmt.Errorf("FATAL: failed to load config file %s, %s", configFile, err.Error())
	}
	return v
}

func consoleConfig() *flag.FlagSet {
	flagSet := flag.NewFlagSet("tantan", flag.ExitOnError)
	flagSet.String("pg-address", defaultTcpAddress, "postgresql address, eg. 0:0:0:0:{port}")
	flagSet.String("pg-username", "", "postgresql db user name")
	flagSet.String("pg-password", "", "postgresql db user password")
	flagSet.String("pg-db-name", "", "postgresql db name")
	flagSet.Int("pg-poolsize", 10, "db connection pool size")
	flagSet.Int("pg-readtimeout", 5, "timeout in seconds when reading from postgresql")
	flagSet.Int("pg-writetimeout", 5, "timeout in seconds when writing to postgresql")
	flagSet.Int("pg-idletimeout", 5, "the amount of time in seconds after which client closes idle db connections")
	flagSet.Bool("verbose", false, "print config value")
	flagSet.Bool("init", false, "if set true, then init db schema and quit. ")

	return flagSet
}
