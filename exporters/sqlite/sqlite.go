package sqlite

import (
	"database/sql"
	_ "embed"

	_ "github.com/mattn/go-sqlite3"

	"azuremyst.org/go-home-sensors/exporters"
	"azuremyst.org/go-home-sensors/log"
	"azuremyst.org/go-home-sensors/sensors"
)

//go:embed table.sql
var tableSQL string

type SqliteExporter struct {
	db *sql.DB
}

func CreateExporter(path string) exporters.Exporter {
	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		log.ErrorLog.Fatalf("Unable to open file %q\n", err)
	}

	_, err = db.Exec(tableSQL)
	if err != nil {
		log.ErrorLog.Fatalf("Failed to create tables %q: %s\n", err, tableSQL)
	}

	tx, err := db.Begin()
	if err != nil {
		log.ErrorLog.Fatal(err)
	}
	stmt, err := tx.Prepare("INSERT INTO measurement(id, description, unit) VALUES(?, ?, ?)")
	if err != nil {
		log.ErrorLog.Fatal(err)
	}
	defer stmt.Close()
	for _, m := range sensors.Measurements {
		_, err = stmt.Exec(m.ID, m.Description, m.Unit)
		if err != nil {
			log.ErrorLog.Fatal(err)
		}
	}
	err = tx.Commit()
	if err != nil {
		log.ErrorLog.Fatal(err)
	}
	return &SqliteExporter{db: db}
}

func (pe *SqliteExporter) Export(recordings []sensors.MeasurementRecording) {
	tx, err := pe.db.Begin()
	if err != nil {
		log.ErrorLog.Fatal(err)
	}
	stmtMeasurement, err := tx.Prepare("INSERT INTO measurement_recording(value, measure_id) VALUES(?, ?)")
	if err != nil {
		log.ErrorLog.Fatal(err)
	}
	defer stmtMeasurement.Close()
	stmtMetadata, err := tx.Prepare("INSERT INTO measurement_meta_data(key, value, recording_id) VALUES(?, ?, ?)")
	if err != nil {
		log.ErrorLog.Fatal(err)
	}
	defer stmtMetadata.Close()

	for _, recording := range recordings {
		res, err := stmtMeasurement.Exec(recording.Value, recording.Measure.ID)
		if err != nil {
			log.ErrorLog.Fatal(err)
		}

		insertId, err := res.LastInsertId()
		if err != nil {
			log.ErrorLog.Fatal(err)
		}
		for k, v := range recording.Metadata {
			_, err = stmtMeasurement.Exec(k, v, insertId)
			if err != nil {
				log.ErrorLog.Fatal(err)
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		log.ErrorLog.Println(err)
	}
}
