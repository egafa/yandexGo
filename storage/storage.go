package storage

import (
	"database/sql"
	"log"

	_ "github.com/jackc/pgx/v4/stdlib"
)

type RowDB struct {
	MType string
	Name  string
	Delta int64
	Value float64
}

type MapData struct {
	GaugeData   map[string]float64
	CounterData map[string]int64
}

func NewRowDB(mtype string, name string, value float64, delta int64) RowDB {
	r := RowDB{}
	r.MType = mtype
	r.Name = name

	switch mtype {
	case "gauge":
		r.Value = value

	case "counter":
		r.Delta = delta
	}

	return r

}

func NewDB(dsn string) *sql.DB {

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Println("database error ", err.Error())
		return nil
	}

	err = createTablles(db)
	if err != nil {
		log.Println("database error ", err.Error())
		return nil
	}
	return db

}

func createTablles(db *sql.DB) error {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS metrics " +
		`("id" SERIAL PRIMARY KEY,` +
		`"typename" varchar(10), "name" varchar(100), delta bigint, val numeric(40,20))`)

	return err
}
func GetFromDatabase(db *sql.DB, typename string, name string) (RowDB, bool) {
	var qtext string
	qtext = "select val as val from metrics where name = $1"
	if typename == "counter" {
		qtext = "select delta as val from metrics where name = $1"
	}
	rows, err := db.Query(qtext, name)
	if err != nil {
		log.Println("database error Select", err.Error())
		return RowDB{}, false
	}

	defer rows.Close()

	var r RowDB
	r.MType = typename
	r.Name = name
	if rows.Next() {
		if typename == "counter" {
			err = rows.Scan(&r.Delta)
		} else {
			err = rows.Scan(&r.Value)
		}
		if err == nil {
			return r, true
		}

		log.Println("database error Select scan", err.Error())
	}

	return RowDB{}, false
}

func SaveToDatabase(db *sql.DB, r RowDB) error {

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	rows, errSelect := db.Query("select name, delta from metrics where name=$1", r.Name)
	if errSelect != nil {
		log.Println("database error Select", errSelect.Error())
		return errSelect
	}
	defer rows.Close()

	if rows.Next() {
		var err error
		if r.MType == "gauge" {
			_, err = tx.Exec("UPDATE metrics SET val = $1 WHERE name=$2", r.Value, r.Name)
		} else {

			var delta int64
			err = rows.Scan(&r.Name, &delta)
			if err != nil {
				log.Println("database error SAVE SCAN ", err.Error())
				return err
			}
			r.Delta = r.Delta + delta
			_, err = tx.Exec("UPDATE metrics SET delta = $1 WHERE name=$2", r.Delta, r.Name)
		}

		if err != nil {
			log.Println("database error UPDATE", err.Error())
			return err
		}

	} else {

		var err error
		if r.MType == "gauge" {
			//_, err = tx.Exec("INSERT INTO metrics (typename, name, val) values(@typename, @name, @val)", r.MType, r.Name, r.Value)
			_, err = tx.Exec("INSERT INTO metrics (typename, name, val) values($1, $2, $3)", r.MType, r.Name, r.Value)
		} else {
			_, err = tx.Exec("INSERT INTO metrics (typename, name, delta) values($1, $2, $3)", r.MType, r.Name, r.Delta)
		}

		if err != nil {
			log.Println("database error INSERT", err.Error())
			return err
		}

	}

	return tx.Commit()

}

func GetMapData(db *sql.DB, typename string) MapData {
	var qtext string
	qtext = "select name, val from metrics where typename = $1"
	if typename == "counter" {
		qtext = "select name, delta from metrics where typename = $1"
	}
	rows, err := db.Query(qtext, typename)
	if err != nil {
		log.Println("database error Select GetMapData", err.Error())
		return MapData{}
	}

	defer rows.Close()

	res := MapData{}
	res.GaugeData = make(map[string]float64)
	res.CounterData = make(map[string]int64)

	var r RowDB
	for rows.Next() {

		if typename == "counter" {
			err = rows.Scan(&r.Name, &r.Delta)
			res.CounterData[r.Name] = r.Delta
		} else {
			err = rows.Scan(&r.Name, &r.Value)
		}

		if err != nil {
			continue
		}

		if typename == "counter" {
			res.CounterData[r.Name] = r.Delta
		} else {
			res.GaugeData[r.Name] = r.Value
		}

	}

	return res
}
