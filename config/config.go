package config

import (
	"flag"
	"os"
	"path/filepath"
	"time"
)

type Config_Agent struct {
	AddrServer     string
	PollInterval   int
	ReportInterval int
	Timeout        int
	DirName        string
	Key            string
}

type Config_Server struct {
	AddrServer    string
	StoreInterval float64
	StoreFile     string
	Restore       bool
	Timeout       int
	TemplateDir   string
	Key           string
}

// LoadConfig creates a Config object that is filled with values from environment variables or set default values
func LoadConfigAgent() *Config_Agent {

	AddrServerEnv := "ADDRESS"
	PollIntervalEnv := "POLLINTERVAL"
	ReportIntervalEnv := "REPORTINTERVAL"
	KeyEnv := "KEY"

	AddrServer := flag.String("a", "127.0.0.1:8080", "адрес сервера")
	PollIntervalStr := flag.String("p", "2s", "интервал получения метрик")
	ReportIntervalStr := flag.String("r", "10s", "интервал отправки метрик")
	KeyStr := flag.String("k", "", "ключ для хеш")
	flag.Parse()

	SetVal(AddrServerEnv, AddrServer)
	SetVal(PollIntervalEnv, PollIntervalStr)
	SetVal(ReportIntervalEnv, ReportIntervalStr)
	SetVal(KeyEnv, KeyStr)

	PollInterval, _ := time.ParseDuration(*PollIntervalStr)
	ReportInterval, _ := time.ParseDuration(*ReportIntervalStr)

	return &Config_Agent{
		AddrServer:     *AddrServer,
		PollInterval:   int(PollInterval.Seconds()),
		ReportInterval: int(ReportInterval.Seconds()),
		Timeout:        1,
		Key:            *KeyStr,
	}
}

func LoadConfigServer() *Config_Server {

	AddrServerEnv := "ADDRESS"
	StoreIntervalEnv := "STORE_INTERVAL"
	StoreFileEnv := "STORE_FILE"
	Restorenv := "RESTORE"
	TemplateDirEnv := "TEMPLATE_DIR"
	KeyEnv := "KEY"

	p, err := os.Executable()
	var TemplateDirStr string
	if err == nil {
		TemplateDirStr = filepath.Dir(p) + "/" // + "/internal/"
	}

	AddrServer := flag.String("a", "127.0.0.1:8080", "адрес сервера")
	StoreFile := flag.String("f", "/tmp/devops-metrics-db.json", "имя файла")
	RestoreStr := flag.String("r", "false", "Восстановить из файла")
	StoreIntervalStr := flag.String("i", "5m", "Интервал сохранения в файл")
	KeyStr := flag.String("k", "", "ключ для хеш")

	flag.Parse()

	SetVal(AddrServerEnv, AddrServer)
	SetVal(StoreFileEnv, StoreFile)
	SetVal(Restorenv, RestoreStr)
	SetVal(StoreIntervalEnv, StoreIntervalStr)
	SetVal(TemplateDirEnv, &TemplateDirStr)
	SetVal(KeyEnv, KeyStr)

	Restore := false
	if *RestoreStr == "true" {
		Restore = true
	}

	StoreInterval, _ := time.ParseDuration(*StoreIntervalStr)

	return &Config_Server{
		AddrServer:    *AddrServer,
		StoreInterval: StoreInterval.Seconds(),
		StoreFile:     *StoreFile,
		Restore:       Restore,
		Timeout:       1,
		TemplateDir:   TemplateDirStr,
		Key:           *KeyStr,
	}
}

func SetVal(env string, val *string) {
	valEnv, ok := os.LookupEnv(env)
	if ok {
		*val = valEnv
	}
}
