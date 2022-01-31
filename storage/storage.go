package storage

import (
	"database/sql"
	"log"

	model "github.com/egafa/yandexGo/api/model"
	_ "github.com/jackc/pgx/v4/stdlib"
)

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
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS gauge " +
		`("id" SERIAL PRIMARY KEY,` +
		`"name" varchar(100), val numeric(15,2))`)

	return err
}

func SaveToDatabase(m model.MapMetric) error {

	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for name, val := range m.GaugeData {
		rows, errSelect := m.DB.Query("SELECT * FROM phonebook WHERE name = &1", name)
		if errSelect != nil {
			return errSelect
		}
		defer rows.Close()

		if rows.Next() {
			_, err = tx.Exec("UPDATE gauge SET val = $1, "+
				"WHERE name=$2",
				val, name)

		} else {

			_, err := tx.Exec("INSERT INTO phonebook VALUES (default, $1, $2)",
				name, val)
			if err != nil {
				return err
			}
		}

		if err != nil {
			return err
		}
	}

	if err != nil {
		return err
	}
	return tx.Commit()

}
