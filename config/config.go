package config

import (
	"os"
	"path/filepath"

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
	StoreInterval int
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
	pflag.Int("p", 2, "интервал получения метрик")
	pflag.Int("r", 10, "интервал отправки метрик")
	pflag.Parse()
	v.BindPFlags(pflag.CommandLine)

	AddrServer := ""
	_, ok := os.LookupEnv(AddrServerEnv)
	if ok {
		AddrServer = v.GetString("AddrServer")
	} else {
		AddrServer = v.GetString("a")
	}

	ReportInterval := 0
	_, ok = os.LookupEnv(ReportIntervalEnv)
	if ok {
		ReportInterval = v.GetInt("ReportInterval")
	} else {
		ReportInterval = v.GetInt("r")
	}

	PollInterval := 0
	_, ok = os.LookupEnv(PollIntervalEnv)
	if ok {
		PollInterval = v.GetInt("PollInterval")
	} else {
		PollInterval = v.GetInt("p")
	}

	return &Config_Agent{
		AddrServer:     AddrServer,
		PollInterval:   PollInterval,
		ReportInterval: ReportInterval,
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
	pflag.Int("i", 300, "Интервал сохранения в файл")
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

	StoreInterval := 0
	_, ok = os.LookupEnv(StoreIntervalEnv)
	if ok {
		StoreInterval = v.GetInt("StoreInterval")
	} else {
		StoreInterval = v.GetInt("i")
	}

	return &Config_Server{
		AddrServer:    AddrServer,
		StoreInterval: StoreInterval,
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
