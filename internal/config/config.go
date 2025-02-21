package config

import (
	"fmt"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

const (
	DiscoModeNone     = ""
	DiscoModeConsulKV = "consul-kv"
	DiscoModeEtcdKV   = "etcd-kv"
	DiscoModeDNS      = "dns"
	DiscoModeDNSSRV   = "dns-srv"
)

type Config struct {
	Server Server  `yaml:"server" json:"server"` // configuration of the public REST server
	Name   string  `yaml:"name" json:"name"`     // used for OTEL as an application identifier
	RqLite *RqLite `yaml:"rqlite" json:"rqlite"`
}

type Server struct {
	Context string `yaml:"context" json:"context" env:"REST_API_CONTEXT" env-default:"/"`
	Addr    string `yaml:"addr" json:"addr" env:"REST_API_ADDR" env-default:":8080"`
}

func InitConfig() Config {
	c := Config{}
	var fileName string
	confFile := os.Getenv("CONFIG_FILE")
	if confFile == "" {
		wd, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		fileName = fmt.Sprintf("%s/conf/zenbpm/conf.yaml", wd)
	} else {
		fileName = confFile
	}
	err := cleanenv.ReadConfig(fileName, &c)
	if err != nil {
		fmt.Printf("Error occurred while reading the configuration file: %s Error: %v", fileName, err)
		panic(err)
	}
	return c
}
