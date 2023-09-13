package sqlite

import (
	"database/sql"
	_ "embed"
	"io/fs"
	"os"
	"path/filepath"

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
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.MkdirAll(filepath.Dir(path), fs.FileMode(os.O_RDWR))
		if err != nil {
			log.ErrorLog.Fatalf("Failed to create path to db file: %q", err)
		}

		file, err := os.Create(path)
		if err != nil {
			log.ErrorLog.Fatalf("Failed to create database file: %q", err)
		}
		file.Close()
	}

	db, err := sql.Open("sqlite3", path)
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
	stmt, err := tx.Prepare("INSERT OR REPLACE INTO measurement(id, description, unit) VALUES(?, ?, ?)")
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
			log.ErrorLog.Fatalf("Failed to insert %s: %q\n", recording.Measure.ID, err)
		}

		insertId, err := res.LastInsertId()
		if err != nil {
			log.ErrorLog.Fatal(err)
		}
		for k, v := range recording.Metadata {
			_, err = stmtMetadata.Exec(k, v, insertId)
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
