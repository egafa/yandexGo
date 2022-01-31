package model

import (
	"context"
	"database/sql"

	"github.com/egafa/yandexGo/config"
	"github.com/egafa/yandexGo/storage"
)

type MetricsDB struct {
	DB *sql.DB
}

func NewMetricDB(cfg *config.Config_Server) MetricsDB {
	m := MetricsDB{}

	if len(cfg.DatabaseDSN) > 0 {
		db := storage.NewDB(cfg.DatabaseDSN)
		if db != nil {
			m.DB = db
		}
	}

	return m
}

func (m MetricsDB) Close() error {
	return m.DB.Close()
}

func (m MetricsDB) SaveGaugeVal(nameMetric string, value float64) {
	var i int64
	storage.SaveToDatabase(m.DB, storage.NewRowDB("gauge", nameMetric, value, i))
}

func (m MetricsDB) GetGaugeVal(nameMetric string) (float64, bool) {

	r, ok := storage.GetFromDatabase(m.DB, "gauge", nameMetric)
	return r.Value, ok

}

func (m MetricsDB) SaveCounterVal(nameMetric string, value int64) {
	var f float64
	storage.SaveToDatabase(m.DB, storage.NewRowDB("counter", nameMetric, f, value))
}

func (m MetricsDB) GetCounterVal(nameMetric string) (int64, bool) {

	r, ok := storage.GetFromDatabase(m.DB, "counter", nameMetric)
	return r.Delta, ok
}

func (m MetricsDB) GetGaugetMetricTemplate() GaugeTemplateMetric {

	r := storage.GetMapData(m.DB, "gauge")

	res := GaugeTemplateMetric{}

	res.Data = r.GaugeData
	res.Typemetric = "Gauge"

	return res
}

func (m MetricsDB) GetCounterMetricTemplate() CounterTemplateMetric {

	r := storage.GetMapData(m.DB, "Counter")

	res := CounterTemplateMetric{}

	res.Data = r.CounterData
	res.Typemetric = "Counter"

	return res
}

func (m MetricsDB) PingContext(ctx context.Context) error {
	return m.DB.PingContext(ctx)
}

func (m MetricsDB) SaveToFile() error {
	return nil
}

func (m MetricsDB) LoadFromFile() error {
	return nil
}
