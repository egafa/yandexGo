package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	envPrefix      = ""
	AddrServer     = "ADDRESS"
	PollInterval   = "POLLINTERVAL"
	ReportInterval = "REPORTINTERVAL"
	Timeout        = "TIMEOUT"
	DirName        = "DIRNAME"

	StoreInterval = "STORE_INTERVAL"
	StoreFile     = "STORE_FILE "
	Restore       = "RESTORE"
)

type Config_Agent struct {
	AddrServer     string `env:"ADDRESS"`
	PollInterval   int    //`env:"POLL_INTERVAL"`
	ReportInterval int    //`env:"REPORT_INTERVAL"`
	Timeout        int
	DirName        string
}

type Config_Server struct {
	AddrServer    string
	StoreInterval int
	StoreFile     string
	Restore       bool
	DirName       string
	Port          string
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
	v.SetDefault(DirName, getDir())

	return &Config_Agent{
		AddrServer:     v.GetString(AddrServer),
		PollInterval:   v.GetInt(PollInterval),
		ReportInterval: v.GetInt(ReportInterval),
		Timeout:        v.GetInt(Timeout),
		DirName:        v.GetString(DirName),
	}
}

func LoadConfigServer() *Config_Server {
	v := viper.New()
	//v.SetEnvPrefix(envPrefix)
	v.AutomaticEnv()

	v.SetDefault(AddrServer, "127.0.0.1:8080")
	v.SetDefault(StoreInterval, 300)
	//v.SetDefault(StoreFile, "D:\\gafa\\Go\\yandexGo\\tmp\\devops-metrics-db.json")
	v.SetDefault(StoreFile, "/tmp/devops-metrics-db.json")
	v.SetDefault(Restore, false)
	v.SetDefault(DirName, getDir())

	return &Config_Server{
		AddrServer:    v.GetString(AddrServer),
		StoreInterval: v.GetInt(StoreInterval),
		StoreFile:     v.GetString(StoreFile),
		Restore:       v.GetBool(Restore),
		DirName:       v.GetString(DirName),
	}
}

func getDir() string {
	ex, err := os.Executable()
	if err != nil {
		return ""
	}

	return filepath.Dir(ex)

}
