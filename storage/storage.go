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

func NewDB(dsn string) (*sql.DB, error) {

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Println("database error ", err.Error())
		return nil, err
	}

	err = createTablles(db)
	if err != nil {
		log.Println("database error ", err.Error())
		return nil, err
	}
	return db, nil

}

func createTablles(db *sql.DB) error {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS metrics " +
		`("id" SERIAL PRIMARY KEY,` +
		`"typename" varchar(10), "name" varchar(100), delta bigint, val numeric(40,20))`)

	_, err = db.Exec("CREATE INDEX IF NOT EXISTS typename ON metrics (typename)")

	_, err = db.Exec("CREATE INDEX IF NOT EXISTS name ON metrics (name)")

	_, err = db.Exec("CREATE INDEX IF NOT EXISTS name_type ON metrics (typename, name)")

	return err
}
func GetFromDatabase(db *sql.DB, typename string, name string) (RowDB, bool) {
	var qtext string
	qtext = "select val from metrics where name = $1"
	if typename == "counter" {
		qtext = "select delta from metrics where name = $1"
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

func SaveMassiveDatabase(db *sql.DB, rows []RowDB) error {

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmtInsert, err := tx.Prepare("INSERT INTO metrics (typename, name, val, delta) values($1,$2,$3,$4)")
	if err != nil {
		return err
	}

	stmtUpdateCounter, err := tx.Prepare("UPDATE metrics SET delta = $1 WHERE name=$2")
	if err != nil {
		return err
	}

	stmtUpdateGauge, err := tx.Prepare("UPDATE metrics SET val = $1 WHERE name=$2")
	if err != nil {
		return err
	}

	stmSelect, err := tx.Prepare("select delta from metrics where name=$1")

	for _, v := range rows {

		row := stmSelect.QueryRow(v.Name)

		var delta int64

		flagRecord := false
		if err = row.Scan(&delta); err != sql.ErrNoRows {
			flagRecord = true
		}

		if flagRecord {

			if v.MType == "gauge" {
				_, err = stmtUpdateGauge.Exec(v.Value, v.Name)
			} else {
				_, err = stmtUpdateCounter.Exec(delta+v.Delta, v.Name)
			}

		} else {

			if v.MType == "gauge" {

				_, err = stmtInsert.Exec(v.MType, v.Name, v.Value, nil)
			} else {

				_, err = stmtInsert.Exec(v.MType, v.Name, nil, v.Delta)
			}

		}
	}

	return tx.Commit()

}
