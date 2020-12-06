package main

import (
	"fmt"

	"github.com/spf13/viper"
)

// Create private data struct to hold config options.
type config struct {
	MysqlDatabase   string `yaml:"mysqlDatabase"`
	MysqlHost       string `yaml:"mysqlHost"`
	MysqlUser       string `yaml:"mysqlUser"`
	MysqlPass       string `yaml:"mysqlPass"`
	SearchDirectory string `yaml:"searchDirectory"`
	SSHServer       string `yaml:"sshServer"`
	SSHPort         string `yaml:"sshPort"`
	SSHUser         string `yaml:"sshUser"`
	SSHKey          string `yaml:"sshKey"`
	SSHHostKey      string `yaml:"SSHHostKey"`
	RemotePath      string `yaml:"remotePath"`
	RemoteOldPath   string `yaml:"remoteOldPath"`
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

	conf.SearchDirectory = appendTrailingSlashIfNotExist(conf.SearchDirectory)
	conf.RemotePath = appendTrailingSlashIfNotExist(conf.RemotePath)

	return conf
}

func appendTrailingSlashIfNotExist(s string) string {
	// if a dir was set
	if len(s) > 0 {
		// if it doesn't end in slash
		if s[len(s)-1:] != "/" {
			s = s + "/"
		}
	}

	return s
}
