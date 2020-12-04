package main

import (
	"fmt"

	"github.com/spf13/viper"
)

// Create private data struct to hold config options.
type config struct {
	MysqlDatabase   string `yaml:"MysqlDatabase"`
	MysqlHost       string `yaml:"mysqlHost"`
	MysqlUser       string `yaml:"mysqlUser"`
	MysqlPass       string `yaml:"mysqlPass"`
	SearchDirectory string `yaml:"searchDirectory"`
}

// Create a new config instance.
var (
	conf *config
)

// Read the config file from the current directory and marshal
// into the conf config struct.
func getConf() *config {
	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	err := viper.ReadInConfig()

	if err != nil {
		fmt.Printf("%v", err)
	}

	conf := &config{}
	err = viper.Unmarshal(conf)
	if err != nil {
		fmt.Printf("unable to decode into config struct, %v", err)
	}

	sd := conf.SearchDirectory

	if len(sd) == 0 {
		panic("Please set a `searchDirectory` in config.yml")
	}

	// if a dir was set
	if len(sd) > 0 {
		// if it doesn't end in slash
		if sd[len(sd)-1:] != "/" {
			conf.SearchDirectory = conf.SearchDirectory + "/"
		}
	}

	return conf
}
