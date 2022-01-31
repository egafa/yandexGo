package storage

import (
	"database/sql"
	"log"

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

func SaveToDatabase(db *sql.DB, gaugeData map[string]float64) error {

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for name, val := range gaugeData {
		rows, errSelect := db.Query("select name from gauge where name=$1", name)
		if errSelect != nil {
			log.Println("database error Select", errSelect.Error())
			return errSelect
		}
		defer rows.Close()

		if rows.Next() {
			_, err = tx.Exec("UPDATE gauge SET val = $1, WHERE name=$2", val, name)
			log.Println("database error UPDATE", err.Error())
			return err
		} else {

			_, err := tx.Exec("INSERT INTO gauge (name, val) values($1, $2)", name, val)
			if err != nil {
				log.Println("database error INSERT", err.Error())
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
