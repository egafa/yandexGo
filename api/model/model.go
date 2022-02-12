package model

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/egafa/yandexGo/config"
)

type Metric interface {
	SaveGaugeVal(nameMetric string, value float64) error
	GetGaugeVal(nameMetric string) (float64, bool)
	SaveCounterVal(nameMetric string, value int64)
	GetCounterVal(nameMetric string) (int64, bool)
	SaveMassiveMetric([]Metrics) error
	GetGaugetMetricTemplate() GaugeTemplateMetric
	GetCounterMetricTemplate() CounterTemplateMetric
	SaveToFile() error
	PingContext(ctx context.Context) error
	Close() error
}

type MapMetric struct {
	GaugeData   map[string]float64 `json:"GaugeData"`
	CounterData map[string]int64   `json:"CounterData"`
	FlagSave    bool               `json:"-"`
	FileName    string             `json:"-"`
}

type GaugeTemplateMetric struct {
	Typemetric string
	Data       map[string]float64
}
type CounterTemplateMetric struct {
	Typemetric string
	Data       map[string]int64
}

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
	Hash  string   `json:"hash,omitempty"`  // значение хеш-функции
}

func GetHash(m Metrics, key string) string {

	if len(key) == 0 {
		return ""
	}

	var src string
	switch m.MType {
	case "counter":
		src = fmt.Sprintf("%s:counter:%d", m.ID, *m.Delta)
	case "gauge":
		src = fmt.Sprintf("%s:gauge:%f", m.ID, *m.Value)
	default:
		return ""

	}

	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(src))
	dst := h.Sum(nil)

	res := fmt.Sprintf("%x", dst)
	return res
}

func NewMetric(cfg *config.Config_Server) Metric {
	var mapMetric Metric
	var err error

	if len(cfg.DatabaseDSN) > 0 {
		mapMetric, err = NewMetricDB(cfg)
		if err == nil {
			return mapMetric
		}
	}

	return NewMapMetricCongig(cfg)

}

func NewMapMetric() MapMetric {
	mapMetricVal := MapMetric{}
	mapMetricVal.GaugeData = make(map[string]float64)
	mapMetricVal.CounterData = make(map[string]int64)
	return mapMetricVal
}

func NewMapMetricCongig(cfg *config.Config_Server) MapMetric {
	mapMetricVal := NewMapMetric()

	if len(cfg.StoreFile) > 0 {
		mapMetricVal.FileName = cfg.StoreFile
	}

	if cfg.StoreInterval == 0 {
		mapMetricVal.FlagSave = true
	}

	if cfg.Restore {
		mapMetricVal.LoadFromFile()
	}

	return mapMetricVal
}

func (m MapMetric) SetFileName(fname string) {
	m.FileName = fname
}

func (m *MapMetric) SetFlagSave(fl bool) {
	m.FlagSave = fl
}

func (m MapMetric) SaveToFile() error {
	//var MapMetricToSave MapMetric

	file, err := os.Create(m.FileName)
	if err != nil {
		log.Println("Ошибка создания файла: ", m.FileName, err.Error())
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	err = encoder.Encode(m)
	if err != nil {
		log.Println("Ошибка сериализации: ", err.Error())
		return err
	}

	return nil
}

func (m MapMetric) LoadFromFile() error {
	//var MapMetricToSave MapMetric

	if m.FileName == "" {
		return nil
	}

	file, err := os.OpenFile(m.FileName, os.O_RDONLY, 0777)
	if err != nil {
		log.Println("Ошибка открытия файла: ", m.FileName, err.Error())
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&m)
	if err != nil {
		log.Println("Ошибка десериализации: ", err.Error())
		return err

	}

	return nil
}

func (m MapMetric) SaveGaugeVal(nameMetric string, value float64) error {
	m.GaugeData[nameMetric] = value
	if m.FlagSave {
		return m.SaveToFile()
	}
	return nil
}

func (m MapMetric) GetGaugeVal(nameMetric string) (float64, bool) {
	res, ok := m.GaugeData[nameMetric]
	if ok {
		return res, true
	} else {
		return 0, false
	}

}

func (m MapMetric) SaveCounterVal(nameMetric string, value int64) {

	v, ok := m.CounterData[nameMetric]
	if ok {
		m.CounterData[nameMetric] = v + value
	} else {
		m.CounterData[nameMetric] = value
	}

	if m.FlagSave {
		m.SaveToFile()
	}
}

func (m MapMetric) GetCounterVal(nameMetric string) (int64, bool) {

	v, ok := m.CounterData[nameMetric]
	if ok {
		return v, true
	} else {
		return 0, false
	}
}

func (m MapMetric) SaveMassiveMetric(dataMetrics []Metrics) error {
	var err error

	for i := 0; i < len(dataMetrics); i++ {
		if dataMetrics[i].MType == "gauge" {
			err = m.SaveGaugeVal(dataMetrics[i].ID, *dataMetrics[i].Value)
		} else {
			m.SaveCounterVal(dataMetrics[i].ID, *dataMetrics[i].Delta)
		}
	}
	return err
}

func (m MapMetric) GetGaugetMetricTemplate() GaugeTemplateMetric {

	res := GaugeTemplateMetric{}

	res.Data = make(map[string]float64)

	res.Data = m.GaugeData
	res.Typemetric = "Gauge"

	return res
}

func (m MapMetric) GetCounterMetricTemplate() CounterTemplateMetric {

	res := CounterTemplateMetric{}

	res.Data = make(map[string]int64)
	res.Typemetric = "Counter"

	res.Data = m.CounterData

	return res
}

func (m MapMetric) SetData(GaugeData map[string]float64, CounterData map[string]int64) {

	m.GaugeData = make(map[string]float64)
	m.CounterData = make(map[string]int64)

	for key, value := range GaugeData {
		m.GaugeData[key] = value
	}

	for key1, value1 := range CounterData {
		m.CounterData[key1] = value1
	}

}

func (m MapMetric) PingContext(ctx context.Context) error {
	return fmt.Errorf("в этом режиме базы данных нет")
}

func (m MapMetric) Close() error {
	return nil
}
