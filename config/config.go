package config

import (
	"os"
	"path/filepath"
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
	pflag.Bool("r", false, "Восстановить из файла")
	pflag.String("i", "5m", "Интервал сохранения в файл")
	pflag.String("f", "/tmp/devops-metrics-db.json", "имя файла")
	pflag.Parse()
	v.BindPFlags(pflag.CommandLine)

	_, ok := os.LookupEnv(Restorenv)
	Restore := false
	if ok {
		Restore = v.GetBool("Restore")
	} else {
		Restore = v.GetBool("r")
	}

	AddrServer := ""
	_, ok = os.LookupEnv(AddrServerEnv)
	if ok {
		AddrServer = v.GetString("AddrServer")
	} else {
		AddrServer = v.GetString("a")
	}

	StoreFile := ""
	_, ok = os.LookupEnv(StoreFileEnv)
	if ok {
		StoreFile = v.GetString("StoreFile")
	} else {
		StoreFile = v.GetString("f")
	}

	StoreIntervalStr := ""
	_, ok = os.LookupEnv(StoreIntervalEnv)
	if ok {
		StoreIntervalStr = v.GetString("StoreInterval")
	} else {
		StoreIntervalStr = v.GetString("i")
	}
	StoreInterval, _ := time.ParseDuration(StoreIntervalStr)

	return &Config_Server{
		AddrServer:    AddrServer,
		StoreInterval: StoreInterval.Seconds(),
		StoreFile:     StoreFile,
		Restore:       Restore,
		Timeout:       v.GetInt("Timeout"),
	}
}

func getDir() string {
	ex, err := os.Executable()
	if err != nil {
		return ""
	}

	return filepath.Dir(ex)

}
