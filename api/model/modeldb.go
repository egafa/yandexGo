package model

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/egafa/yandexGo/config"
	"github.com/egafa/yandexGo/storage"
)

type MetricsDB struct {
	DB *sql.DB
}

func NewMetricDB(cfg *config.Config_Server) (MetricsDB, error) {
	m := MetricsDB{}

	db, err := storage.NewDB(cfg.DatabaseDSN)
	if err == nil {
		m.DB = db
		return m, nil
	}
	return m, err

}

func (m MetricsDB) Close() error {
	return m.DB.Close()
}

func (m MetricsDB) SaveGaugeVal(nameMetric string, value float64) error {
	var i int64
	return storage.SaveToDatabase(m.DB, storage.NewRowDB("gauge", nameMetric, value, i))
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

func (m MetricsDB) SaveMassiveMetric(dataMetrics []Metrics) error {

	massiveDB := make([]storage.RowDB, len(dataMetrics))
	for i, metric := range dataMetrics {

		if metric.MType == "gauge" {
			if metric.Value == nil {
				return fmt.Errorf("Для  ID ", metric.ID, "не задано значение")
			}
		} else {
			if metric.Delta == nil {
				return fmt.Errorf("Для  ID ", metric.ID, "не задано значение")
			}
		}

		r := storage.RowDB{
			Name:  metric.ID,
			MType: metric.MType,
		}
		if r.MType == "gauge" {

			r.Value = *metric.Value
		} else {
			r.Delta = *metric.Delta
		}
		massiveDB[i] = r
	}

	return storage.SaveMassiveDatabase(m.DB, massiveDB)
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
