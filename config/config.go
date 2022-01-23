package config

import (
	"os"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const (
	envPrefix = ""
)

type Config_Agent struct {
	AddrServer     string
	PollInterval   int
	ReportInterval int
	Timeout        int
	DirName        string
}

type Config_Server struct {
	AddrServer    string
	StoreInterval float64
	StoreFile     string
	Restore       bool
	Timeout       int
}

// LoadConfig creates a Config object that is filled with values from environment variables or set default values
func LoadConfigAgent() *Config_Agent {

	AddrServerEnv := "ADDRESS"
	PollIntervalEnv := "POLLINTERVAL"
	ReportIntervalEnv := "REPORTINTERVAL"

	v := viper.New()
	v.BindEnv("AddrServer", AddrServerEnv)
	v.BindEnv("PollInterval", PollIntervalEnv)
	v.BindEnv("ReportInterval", ReportIntervalEnv)
	//v.SetEnvPrefix(envPrefix)
	v.AutomaticEnv()

	v.SetDefault("Timeout", 1)

	pflag.String("a", "127.0.0.1:8080", "адрес сервера")
	pflag.String("p", "2s", "интервал получения метрик")
	pflag.String("r", "10s", "интервал отправки метрик")
	pflag.Parse()
	v.BindPFlags(pflag.CommandLine)

	AddrServer := ""
	_, ok := os.LookupEnv(AddrServerEnv)
	if ok {
		AddrServer = v.GetString("AddrServer")
	} else {
		AddrServer = v.GetString("a")
	}

	ReportIntervalStr := ""
	_, ok = os.LookupEnv(ReportIntervalEnv)
	if ok {
		ReportIntervalStr = v.GetString("ReportInterval")
	} else {
		ReportIntervalStr = v.GetString("r")
	}
	ReportInterval, _ := time.ParseDuration(ReportIntervalStr)

	PollIntervalStr := ""
	_, ok = os.LookupEnv(PollIntervalEnv)
	if ok {
		PollIntervalStr = v.GetString("PollInterval")
	} else {
		PollIntervalStr = v.GetString("p")
	}
	PollInterval, _ := time.ParseDuration(PollIntervalStr)

	return &Config_Agent{
		AddrServer:     AddrServer,
		PollInterval:   int(PollInterval.Seconds()),
		ReportInterval: int(ReportInterval.Seconds()),
		Timeout:        v.GetInt("Timeout"),
	}
}

func LoadConfigServer() *Config_Server {

	AddrServerEnv := "ADDRESS"
	StoreIntervalEnv := "STORE_INTERVAL"
	StoreFileEnv := "STORE_FILE"
	Restorenv := "RESTORE"

	v := viper.New()
	v.BindEnv("AddrServer", AddrServerEnv)
	v.BindEnv("StoreInterval", StoreIntervalEnv)
	v.BindEnv("StoreFile", StoreFileEnv)
	v.BindEnv("Restore", Restorenv)

	//v.SetEnvPrefix(envPrefix)
	v.AutomaticEnv()

	v.SetDefault("Timeout", 1)

	pflag.String("a", "127.0.0.1:8080", "адрес сервера")
	pflag.String("r", "false", "Восстановить из файла")
	pflag.String("i", "5m", "Интервал сохранения в файл")
	pflag.String("f", "/tmp/devops-metrics-db.json", "имя файла")
	pflag.Parse()
	v.BindPFlags(pflag.CommandLine)

	RestoreStr := GetVal(v, Restorenv, "Restore", "r")
	Restore := true
	if RestoreStr == "false" {
		Restore = false
	}

	AddrServer := GetVal(v, AddrServerEnv, "AddrServer", "a")
	StoreFile := GetVal(v, StoreFileEnv, "StoreFile", "f")

	StoreIntervalStr := GetVal(v, StoreIntervalEnv, "StoreInterval", "i")
	StoreInterval, _ := time.ParseDuration(StoreIntervalStr)

	return &Config_Server{
		AddrServer:    AddrServer,
		StoreInterval: StoreInterval.Seconds(),
		StoreFile:     StoreFile,
		Restore:       Restore,
		Timeout:       v.GetInt("Timeout"),
	}
}

func GetVal(v *viper.Viper, env string, envName string, flagName string) string {
	_, ok := os.LookupEnv(env)
	if ok {
		return v.GetString(envName)
	} else {
		return v.GetString(flagName)
	}
}
