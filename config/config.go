package config

import (
	"github.com/spf13/viper"
)

const (
	envPrefix      = ""
	AddrServer     = "ADDRESS"
	PollInterval   = "POLLINTERVAL"
	ReportInterval = "REPORTINTERVAL"
	Timeout        = "TIMEOUT"
	DirName        = "DIRNAME"
)

type Config_Agent struct {
	AddrServer     string `env:"ADDRESS"`
	PollInterval   int    //`env:"POLL_INTERVAL"`
	ReportInterval int    //`env:"REPORT_INTERVAL"`
	Timeout        int
	DirName        string
	Address        string
}

// LoadConfig creates a Config object that is filled with values from environment variables or set default values
func LoadConfigAgent() *Config_Agent {
	v := viper.New()
	//v.SetEnvPrefix(envPrefix)
	v.AutomaticEnv()

	v.SetDefault(AddrServer, "127.0.0.1:8080")
	v.SetDefault(PollInterval, 2)
	v.SetDefault(ReportInterval, 10)
	v.SetDefault(Timeout, 1)
	v.SetDefault(DirName, "")

	return &Config_Agent{
		AddrServer:     v.GetString(AddrServer),
		PollInterval:   v.GetInt(PollInterval),
		ReportInterval: v.GetInt(ReportInterval),
		Timeout:        v.GetInt(Timeout),
		DirName:        v.GetString(DirName),
	}
}
