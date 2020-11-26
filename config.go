package main

import (
	"fmt"

	"github.com/spf13/viper"
)

// Create private data struct to hold config options.
type config struct {
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

	return conf
}
